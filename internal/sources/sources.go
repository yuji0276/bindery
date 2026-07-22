// Package sources は各ツール固有のパーサと、その供給元リストを提供する。
// 型 (bindery.Binding / bindery.Source) は bindery パッケージに依存するが、
// bindery 側は sources に依存しない (循環回避)。配線は main が行う。
package sources

import (
	"os"
	"path/filepath"

	"bindery/internal/bindery"
)

func home() string {
	h, _ := os.UserHomeDir()
	return h
}

// List は既知の供給元 (skhd/nvim/wezterm) を返す。
func List() []bindery.Source {
	h := home()
	return []bindery.Source{
		{Name: "skhd", Path: filepath.Join(h, ".config/skhd/skhdrc"), Parse: parseSkhd},
		{Name: "nvim", Path: filepath.Join(h, ".config/nvim/lua/config/keymaps.lua"), Parse: parseNvim},
		{Name: "wezterm", Path: filepath.Join(h, ".config/wezterm/keybinds.lua"), Parse: parseWezterm},
	}
}
