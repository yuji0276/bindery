package bindery

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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

func Collect(filterSrc string) ([]Binding, []string) {
	managed := managedSet()
	var all []Binding
	var warnings []string
	for _, s := range sources() {
		if filterSrc != "" && s.name != filterSrc {
			continue
		}
		if _, err := os.Stat(s.path); err != nil {
			continue
		}
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
