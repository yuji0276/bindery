package sources

import (
	"bufio"
	"os"
	"strings"

	"bindery/internal/bindery"
)

func parseSkhd(path string) ([]bindery.Binding, error) {
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
		raw := strings.TrimSpace(sc.Text())
		if raw == "" || strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "::") || strings.HasPrefix(raw, ".") {
			continue
		}
		keydef, command, ok := strings.Cut(raw, ":")
		if !ok {
			continue
		}
		keydef = strings.TrimSpace(keydef)
		command = strings.TrimSpace(command)
		if keydef == "" {
			continue
		}

		desc := ""
		if before, after, ok := strings.Cut(command, "#"); ok {
			command = strings.TrimSpace(before)
			desc = strings.TrimSpace(after)
		}
		if desc == "" {
			desc = command
		}

		mods, key := "", keydef
		if before, after, ok := strings.Cut(keydef, "-"); ok {
			mods = normalizeMods(before)
			key = strings.TrimSpace(after)
		}

		out = append(out, bindery.Binding{
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
	cleaned := parts[:0]
	for _, p := range parts {
		if p != "" {
			cleaned = append(cleaned, p)
		}
	}
	return strings.Join(cleaned, "+")
}
