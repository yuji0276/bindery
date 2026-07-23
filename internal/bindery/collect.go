package bindery

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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

func Collect(srcs []Source, filterSrc string) ([]Binding, []string) {
	managed := managedSet()
	var all []Binding
	var warnings []string
	for _, s := range srcs {
		if filterSrc != "" && s.Name != filterSrc {
			continue
		}
		if _, err := os.Stat(s.Path); err != nil {
			continue
		}
		if managed != nil && !managed[s.Path] {
			warnings = append(warnings, fmt.Sprintf("%s は chezmoi 管理外のためスキップ: %s", s.Name, s.Path))
			continue
		}
		bs, err := s.Parse(s.Path)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s のパースに失敗: %v", s.Name, err))
			continue
		}
		all = append(all, bs...)
	}
	return all, warnings
}
