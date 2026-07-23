package sources

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

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"bindery/internal/bindery"
)

var (
	reWezKey  = regexp.MustCompile(`key\s*=\s*["']([^"']*)["']`)
	reWezMods = regexp.MustCompile(`mods\s*=\s*["']([^"']*)["']`)
	reWezCmt  = regexp.MustCompile(`^\s*--+\s?(.*)$`)
)

func parseWezterm(path string) ([]bindery.Binding, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []bindery.Binding
	sc := bufio.NewScanner(f)
	line := 0
	pendingDesc := ""

	// key を見つけたら 1 エントリを開始し、続く行の mods を拾う。
	var cur *bindery.Binding
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
			cur = &bindery.Binding{
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
