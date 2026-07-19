package bindery

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// ---- skhd ---------------------------------------------------------------
//
// 例:
//   cmd - h : yabai -m window --focus west
//   shift + alt - h : yabai -m window --swap west
//   alt - r     : yabai -m space --rotate 90   # ツリーを90度回転
//
// 形式: <mods> - <key> : <command>   [# 説明]
// mods は "+" で連結。command 中に ":" を含むため最初の ":" のみで分割する。

func parseSkhd(path string) ([]Binding, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []Binding
	sc := bufio.NewScanner(f)
	line := 0
	for sc.Scan() {
		line++
		raw := strings.TrimSpace(sc.Text())
		if raw == "" || strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "::") || strings.HasPrefix(raw, ".") {
			continue
		}
		// keydef : command で分割 (最初の ":" のみ)
		colon := strings.Index(raw, ":")
		if colon < 0 {
			continue
		}
		keydef := strings.TrimSpace(raw[:colon])
		command := strings.TrimSpace(raw[colon+1:])
		if keydef == "" {
			continue
		}

		// 末尾コメントを説明として取り出す (" #" 以降)。
		desc := ""
		if h := strings.Index(command, "#"); h >= 0 {
			desc = strings.TrimSpace(command[h+1:])
			command = strings.TrimSpace(command[:h])
		}
		if desc == "" {
			desc = command
		}

		// mods と key を分割 (mods は "+" 連結なので最初の "-" が区切り)。
		mods, key := "", keydef
		if d := strings.Index(keydef, "-"); d >= 0 {
			mods = normalizeMods(keydef[:d])
			key = strings.TrimSpace(keydef[d+1:])
		}

		out = append(out, Binding{
			Source: "skhd",
			Mods:   mods,
			Key:    key,
			Desc:   desc,
			File:   path,
			Line:   line,
		})
	}
	return out, sc.Err()
}

func normalizeMods(s string) string {
	parts := strings.Split(s, "+")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	// 空要素を除去
	cleaned := parts[:0]
	for _, p := range parts {
		if p != "" {
			cleaned = append(cleaned, p)
		}
	}
	return strings.Join(cleaned, "+")
}

// ---- Neovim -------------------------------------------------------------
//
// 例:
//   map("n", "<leader>w", "<cmd>w<cr>", { desc = "保存" })
//   map("n", "<C-h>", "<C-w>h")
//   vim.keymap.set({ "n", "v" }, "<leader>x", ..., { desc = "..." })
//
// Mods 列にはモード ("n"/"v"/...) を、Key 列にはキーコードを入れる。

var (
	reNvimMap  = regexp.MustCompile(`(?:map|vim\.keymap\.set)\(\s*(\{[^}]*\}|"[^"]*")\s*,\s*"([^"]+)"`)
	reNvimDesc = regexp.MustCompile(`desc\s*=\s*"([^"]*)"`)
)

func parseNvim(path string) ([]Binding, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []Binding
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
		out = append(out, Binding{
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

// ---- WezTerm ------------------------------------------------------------
//
// 例:
//   -- workspaceの切り替え
//   {
//       key = "w",
//       mods = "LEADER",
//       action = act.ShowLauncherArgs(...),
//   },
//
// key と mods は別行に現れることが多い。直前の "--" コメントを説明として使う。

var (
	reWezKey  = regexp.MustCompile(`key\s*=\s*["']([^"']*)["']`)
	reWezMods = regexp.MustCompile(`mods\s*=\s*["']([^"']*)["']`)
	reWezCmt  = regexp.MustCompile(`^\s*--+\s?(.*)$`)
)

func parseWezterm(path string) ([]Binding, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []Binding
	sc := bufio.NewScanner(f)
	line := 0
	pendingDesc := ""

	// key を見つけたら 1 エントリを開始し、続く行の mods を拾う。
	var cur *Binding
	flush := func() {
		if cur != nil {
			out = append(out, *cur)
			cur = nil
		}
	}

	for sc.Scan() {
		line++
		text := sc.Text()

		if c := reWezCmt.FindStringSubmatch(text); c != nil {
			s := strings.TrimSpace(c[1])
			// コメントアウトされた定義 (例: -- { key = 'X', ... }) は説明にしない。
			if s != "" && !strings.Contains(s, "{") && !reWezKey.MatchString(s) {
				pendingDesc = s
			}
			continue
		}
		if k := reWezKey.FindStringSubmatch(text); k != nil {
			flush() // 直前のエントリを確定
			cur = &Binding{
				Source: "wezterm",
				Key:    k[1],
				Desc:   pendingDesc,
				File:   path,
				Line:   line,
			}
			pendingDesc = ""
			// key と mods が同一行のこともあるので continue せず mods も見る。
		}
		if m := reWezMods.FindStringSubmatch(text); m != nil && cur != nil {
			cur.Mods = m[1]
		}
	}
	flush()
	return out, sc.Err()
}
