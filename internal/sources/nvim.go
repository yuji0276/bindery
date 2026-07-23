package sources

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"bindery/internal/bindery"
)

// parseNvim は nvim を headless で起動し、実際に効いているマッピングを
// nvim_get_keymap() で JSON として吐かせてから取り込む。
// 正規表現で Lua を舐めるより確実で、desc・定義元ファイル・行番号まで取れる。
// nvim が無い環境では従来の正規表現パーサ (parseNvimFile) にフォールバックする。
func parseNvim(path string) ([]bindery.Binding, error) {
	if _, err := exec.LookPath("nvim"); err != nil {
		return parseNvimFile(path)
	}
	return nvimDumpKeymaps()
}

// nvim_get_keymap() の 1 エントリを Lua 側で整形したもの。
type nvimEntry struct {
	Mods string `json:"mods"` // モード (n/i/v ...) を "/" 連結
	Key  string `json:"key"`  // lhs
	Desc string `json:"desc"`
	File string `json:"file"` // 定義元スクリプト (解決できなければ空)
	Line int    `json:"line"`
}

// nvim 内で全モードの keymap を集約し JSON を書き出す Lua スクリプト。
// 出力先は環境変数 BINDERY_OUT。VimEnter 後に実行してプラグイン等が
// 設定するマッピングも拾えるようにする。
const nvimDumpLua = `
local out = assert(os.getenv("BINDERY_OUT"))
local scripts = {}
local ok, infos = pcall(vim.fn.getscriptinfo)
if ok then
  for _, s in ipairs(infos) do scripts[s.sid] = s.name end
end
local modes = {"n", "i", "v", "x", "s", "o", "t", "c"}
local order, byKey = {}, {}
for _, qm in ipairs(modes) do
  for _, m in ipairs(vim.api.nvim_get_keymap(qm)) do
    local desc = m.desc or ""
    if desc == "" then desc = m.rhs or "" end
    -- rhs が Lua 関数のマップは lnum=0 なので callback の定義元から復元する。
    local file, line = "", 0
    if m.callback then
      local info = debug.getinfo(m.callback, "S")
      if info and info.source and info.source:sub(1, 1) == "@" then
        file = info.source:sub(2)
        line = info.linedefined or 0
      end
    end
    if file == "" then
      file = scripts[m.sid] or ""
      line = m.lnum or 0
    end
    local id = table.concat({m.lhs, file, tostring(line), desc}, "\0")
    local e = byKey[id]
    if not e then
      e = {modeset = {}, modes = {}, lhs = m.lhs, desc = desc, file = file, line = line}
      byKey[id] = e
      order[#order + 1] = e
    end
    local mode = m.mode
    if mode == " " or mode == "" then mode = "n/v/o" end
    if not e.modeset[mode] then
      e.modeset[mode] = true
      e.modes[#e.modes + 1] = mode
    end
  end
end
local result = {}
for _, e in ipairs(order) do
  result[#result + 1] = {
    mods = table.concat(e.modes, "/"),
    key = e.lhs,
    desc = e.desc,
    file = e.file,
    line = e.line,
  }
end
local f = assert(io.open(out, "w"))
f:write(vim.json.encode(result))
f:close()
vim.cmd("qa!")
`

func nvimDumpKeymaps() ([]bindery.Binding, error) {
	script, err := os.CreateTemp("", "bindery-*.lua")
	if err != nil {
		return nil, err
	}
	defer os.Remove(script.Name())
	if _, err := script.WriteString(nvimDumpLua); err != nil {
		script.Close()
		return nil, err
	}
	script.Close()

	outFile, err := os.CreateTemp("", "bindery-*.json")
	if err != nil {
		return nil, err
	}
	outFile.Close()
	defer os.Remove(outFile.Name())

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// VimEnter で実行することで遅延ロードされるマッピングも拾う。
	cmd := exec.CommandContext(ctx, "nvim", "--headless", "-n",
		"-c", "autocmd VimEnter * luafile "+script.Name())
	cmd.Env = append(os.Environ(), "BINDERY_OUT="+outFile.Name())
	if err := cmd.Run(); err != nil && ctx.Err() == nil {
		return nil, fmt.Errorf("nvim 実行に失敗: %w", err)
	}
	if ctx.Err() != nil {
		return nil, fmt.Errorf("nvim がタイムアウトしました")
	}

	data, err := os.ReadFile(outFile.Name())
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, nil
	}
	var entries []nvimEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("nvim 出力の解析に失敗: %w", err)
	}
	// callback を持たない文字列 rhs のマップは lnum=0 で、さらに sid が
	// ラッパー関数経由だと呼び出し元スクリプトに誤って紐づく。
	// desc(または rhs) は設定ファイルにそのまま書かれているので、nvim 設定
	// ツリー全体を grep して本当の file:line を引き当てる。
	idx := buildConfigIndex()
	out := make([]bindery.Binding, 0, len(entries))
	for _, e := range entries {
		if e.Line == 0 && e.Desc != "" {
			if file, line := idx.find(e.Desc); line > 0 {
				e.File, e.Line = file, line
			}
		}
		out = append(out, bindery.Binding{
			Source: "nvim",
			Mods:   e.Mods,
			Key:    e.Key,
			Desc:   e.Desc,
			File:   e.File,
			Line:   e.Line,
		})
	}
	return out, nil
}

type configIndex struct {
	files []indexedFile
}

type indexedFile struct {
	path  string
	lines []string
}

// buildConfigIndex は ~/.config/nvim 配下の .lua / .vim を読み込んでおく。
func buildConfigIndex() configIndex {
	root := filepath.Join(home(), ".config/nvim")
	var idx configIndex
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if ext := filepath.Ext(p); ext != ".lua" && ext != ".vim" {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		idx.files = append(idx.files, indexedFile{p, strings.Split(string(data), "\n")})
		return nil
	})
	return idx
}

// find は needle を含む最初の行を持つファイルパスと行番号を返す。
func (idx configIndex) find(needle string) (string, int) {
	for _, f := range idx.files {
		for i, l := range f.lines {
			if strings.Contains(l, needle) {
				return f.path, i + 1
			}
		}
	}
	return "", 0
}

var (
	reNvimMap  = regexp.MustCompile(`(?:map|vim\.keymap\.set)\(\s*(\{[^}]*\}|"[^"]*")\s*,\s*"([^"]+)"`)
	reNvimDesc = regexp.MustCompile(`desc\s*=\s*"([^"]*)"`)
)

// parseNvimFile は nvim が使えない環境向けの正規表現フォールバック。
func parseNvimFile(path string) ([]bindery.Binding, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []bindery.Binding
	sc := bufio.NewScanner(f)
	line := 0
	for sc.Scan() {
		line++
		text := sc.Text()
		trimmed := strings.TrimSpace(text)
		if strings.HasPrefix(trimmed, "--") { // Lua コメント行
			continue
		}
		m := reNvimMap.FindStringSubmatch(text)
		if m == nil {
			continue
		}
		mode := strings.Trim(m[1], `"`)
		mode = strings.Trim(mode, "{}")
		mode = strings.ReplaceAll(mode, `"`, "")
		mode = strings.Join(strings.Fields(strings.ReplaceAll(mode, ",", " ")), "/")

		key := m[2]
		desc := ""
		if d := reNvimDesc.FindStringSubmatch(text); d != nil {
			desc = d[1]
		}
		out = append(out, bindery.Binding{
			Source: "nvim",
			Mods:   mode,
			Key:    key,
			Desc:   desc,
			File:   path,
			Line:   line,
		})
	}
	return out, sc.Err()
}
