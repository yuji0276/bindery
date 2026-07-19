package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"bindery/internal/bindery"
)

// RunFzf は fzf に渡してインタラクティブに選択させ、選んだ定義を $EDITOR で開く。
func RunFzf(bs []bindery.Binding) {
	// 各行の末尾に \x1f で file:line を隠しフィールドとして付与し、--with-nth で非表示にする。
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
