package cli

import "testing"

func TestMdEsc(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		skip bool
	}{
		{"空文字はそのまま空文字として扱う", "", "", false},
		{"|を含む文字列", "| a | b |", "\\| a \\| b \\|", false},
		{"|を含まない文字列", "abc", "abc", false},
		{"エスケープ済みの|はリテラルとして表示させる", "\\| a \\| b \\|", "\\\\\\| a \\\\\\| b \\\\\\|", true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("issue#2:render.go:50を修正")
			}
			got := mdEsc(tt.in)
			if got != tt.want {
				t.Errorf("mdEsc(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
