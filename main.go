// bindery — chezmoi 管理下の dotfile からキーバインド定義を集約して一覧表示するツール。
//
// 対応ソース: skhd / Neovim / WezTerm
//
//	bindery            # fzf があればインタラクティブ検索、なければ表を表示
//	bindery -l         # 常に表で一覧表示 (fzf を使わない)
//	bindery --md       # Markdown で出力 (keybind.md 生成用)
//	bindery --json     # JSON で出力
//	bindery -s skhd    # ソースで絞り込み (skhd/nvim/wezterm)
//
// fzf で選択して Enter すると、その定義がある行を $EDITOR (既定 nvim) で開く。
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

// Binding は 1 つのキー割り当てを表す。
type Binding struct {
	Source string `json:"source"` // skhd / nvim / wezterm
	Mods   string `json:"mods"`   // 修飾キー (nvim ではモード)
	Key    string `json:"key"`    // キー
	Desc   string `json:"desc"`   // 説明 (コメント or コマンド)
	File   string `json:"file"`   // 定義元ファイル (絶対パス)
	Line   int    `json:"line"`   // 定義元の行番号
}

// source は 1 つの設定ソースと、そのパーサを結びつける。
type source struct {
	name  string
	path  string // ~ 展開済みの既定パス
	parse func(path string) ([]Binding, error)
}

func home() string {
	h, _ := os.UserHomeDir()
	return h
}

func sources() []source {
	h := home()
	return []source{
		{"skhd", filepath.Join(h, ".config/skhd/skhdrc"), parseSkhd},
		{"nvim", filepath.Join(h, ".config/nvim/lua/config/keymaps.lua"), parseNvim},
		{"wezterm", filepath.Join(h, ".config/wezterm/keybinds.lua"), parseWezterm},
	}
}

// managedSet は `chezmoi managed` の絶対パス集合を返す。
// chezmoi が無い場合は nil を返し、呼び出し側はファイルの存在のみで判断する。
func managedSet() map[string]bool {
	if _, err := exec.LookPath("chezmoi"); err != nil {
		return nil
	}
	out, err := exec.Command("chezmoi", "managed", "--path-style=absolute").Output()
	if err != nil {
		return nil
	}
	set := map[string]bool{}
	for _, line := range strings.Split(string(out), "\n") {
		if p := strings.TrimSpace(line); p != "" {
			set[p] = true
		}
	}
	return set
}

func collect(filterSrc string) ([]Binding, []string) {
	managed := managedSet()
	var all []Binding
	var warnings []string
	for _, s := range sources() {
		if filterSrc != "" && s.name != filterSrc {
			continue
		}
		if _, err := os.Stat(s.path); err != nil {
			continue // ファイルが無いソースは静かにスキップ
		}
		// chezmoi 管理下かどうかを確認 (chezmoi が使える場合のみ)。
		if managed != nil && !managed[s.path] {
			warnings = append(warnings, fmt.Sprintf("%s は chezmoi 管理外のためスキップ: %s", s.name, s.path))
			continue
		}
		bs, err := s.parse(s.path)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s のパースに失敗: %v", s.name, err))
			continue
		}
		all = append(all, bs...)
	}
	return all, warnings
}

func main() {
	var (
		list     bool
		asMD     bool
		asJSON   bool
		noFzf    bool
		filter   string
	)
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-l", "--list":
			list = true
		case "--md", "--markdown":
			asMD = true
		case "--json":
			asJSON = true
		case "--no-fzf":
			noFzf = true
		case "-s", "--source":
			if i+1 < len(args) {
				i++
				filter = args[i]
			}
		case "-h", "--help":
			printHelp()
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n\n", args[i])
			printHelp()
			os.Exit(2)
		}
	}

	bindings, warnings := collect(filter)
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}
	if len(bindings) == 0 {
		fmt.Fprintln(os.Stderr, "キーバインドが見つかりませんでした。")
		os.Exit(1)
	}

	sortBindings(bindings)

	switch {
	case asJSON:
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		enc.Encode(bindings)
	case asMD:
		printMarkdown(os.Stdout, bindings)
	case list || noFzf || !stdoutIsTTY() || !hasFzf():
		printTable(os.Stdout, bindings)
	default:
		runFzf(bindings)
	}
}

func sortBindings(bs []Binding) {
	sort.SliceStable(bs, func(i, j int) bool {
		if bs[i].Source != bs[j].Source {
			return bs[i].Source < bs[j].Source
		}
		return bs[i].Line < bs[j].Line
	})
}

func stdoutIsTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func hasFzf() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

// printTable は tabwriter で桁揃えした表を書き出す。
func printTable(w *os.File, bs []Binding) {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "SOURCE\tMODS\tKEY\tDESCRIPTION")
	for _, b := range bs {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", b.Source, b.Mods, b.Key, b.Desc)
	}
	tw.Flush()
}

// printMarkdown はソースごとに見出し + 表で書き出す。
func printMarkdown(w *os.File, bs []Binding) {
	fmt.Fprintln(w, "# キーバインディング一覧 (自動生成)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "`bindery --md` で生成。手で編集しないでください。")
	fmt.Fprintln(w)
	cur := ""
	for _, b := range bs {
		if b.Source != cur {
			cur = b.Source
			fmt.Fprintf(w, "\n## %s\n\n", b.Source)
			fmt.Fprintln(w, "| Mods | Key | 説明 |")
			fmt.Fprintln(w, "|------|-----|------|")
		}
		fmt.Fprintf(w, "| %s | `%s` | %s |\n", mdEsc(b.Mods), b.Key, mdEsc(b.Desc))
	}
}

func mdEsc(s string) string { return strings.ReplaceAll(s, "|", "\\|") }

// runFzf は fzf に渡してインタラクティブに選択させ、選んだ定義を $EDITOR で開く。
func runFzf(bs []Binding) {
	// 各行の末尾に \t で file:line を隠しフィールドとして付与し、--with-nth で非表示にする。
	const sep = "\x1f" // Unit Separator: 表示テキストに現れない区切り
	var sb strings.Builder
	// 見やすさのため列を固定幅で整形する。
	wSrc, wMods, wKey := 0, 0, 0
	for _, b := range bs {
		wSrc = max(wSrc, len(b.Source))
		wMods = max(wMods, dispLen(b.Mods))
		wKey = max(wKey, dispLen(b.Key))
	}
	for _, b := range bs {
		display := fmt.Sprintf("%-*s  %s  %s  %s",
			wSrc, b.Source,
			pad(b.Mods, wMods),
			pad(b.Key, wKey),
			b.Desc)
		fmt.Fprintf(&sb, "%s%s%s:%d\n", display, sep, b.File, b.Line)
	}

	cmd := exec.Command("fzf",
		"--ansi",
		"--delimiter", sep,
		"--with-nth", "1", // 表示は先頭フィールドのみ (file:line を隠す)
		"--prompt", "bindery> ",
		"--header", "Enter: 定義元を開く",
		"--height", "80%",
		"--layout", "reverse",
	)
	cmd.Stdin = strings.NewReader(sb.String())
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		// fzf を ESC で抜けた場合など。静かに終了。
		return
	}
	sel := strings.TrimRight(string(out), "\n")
	if sel == "" {
		return
	}
	parts := strings.SplitN(sel, sep, 2)
	if len(parts) != 2 {
		return
	}
	openInEditor(parts[1])
}

// openInEditor は "file:line" を $EDITOR (既定 nvim) で開く。
func openInEditor(fileLine string) {
	idx := strings.LastIndex(fileLine, ":")
	file, lineStr := fileLine, ""
	if idx >= 0 {
		file, lineStr = fileLine[:idx], fileLine[idx+1:]
	}
	line, _ := strconv.Atoi(lineStr)
	if line == 0 {
		line = 1
	}

	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "nvim"
	}

	var cmd *exec.Cmd
	base := filepath.Base(editor)
	switch {
	case strings.Contains(base, "code"): // VS Code
		cmd = exec.Command(editor, "-g", fmt.Sprintf("%s:%d", file, line))
	default: // vim / nvim 系
		cmd = exec.Command(editor, fmt.Sprintf("+%d", line), file)
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "エディタの起動に失敗:", err)
	}
}

func printHelp() {
	fmt.Print(`bindery — chezmoi 管理下の dotfile からキーバインド一覧を集約

USAGE:
  bindery [flags]

FLAGS:
  (なし)          fzf でインタラクティブ検索 (fzf 未導入なら表を表示)
  -l, --list      常に表で一覧表示
  --md            Markdown で出力 (keybind.md 生成用)
  --json          JSON で出力
  -s, --source X  ソースで絞り込み (skhd|nvim|wezterm)
  -h, --help      このヘルプ

EXAMPLES:
  bindery
  bindery -l | less
  bindery --md > ~/.config/keybind.md
  bindery -s skhd
`)
}

// pad は表示幅 (全角=2) を考慮して右詰めスペースを付ける。
func pad(s string, w int) string {
	diff := w - dispLen(s)
	if diff <= 0 {
		return s
	}
	return s + strings.Repeat(" ", diff)
}

// dispLen は端末表示幅を概算する (ASCII=1, それ以外=2)。
func dispLen(s string) int {
	n := 0
	for _, r := range s {
		if r < 0x80 {
			n++
		} else {
			n += 2
		}
	}
	return n
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
