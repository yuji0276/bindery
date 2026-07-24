package sources

import "testing"

func TestNormalizeMods(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"単一のmod", "cmd", "cmd"},
		{"前後の空白を除去", "  cmd  ", "cmd"},
		{"+区切りの各要素をトリム", "cmd + shift", "cmd+shift"},
		{"空の要素は除去される", "cmd +  + shift", "cmd+shift"},
		{"空文字列は空文字列のまま", "", ""},
		{"3つのmod", "cmd + shift + alt", "cmd+shift+alt"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeMods(tt.in)

			if got != tt.want {
				t.Errorf("normalizeMods(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
