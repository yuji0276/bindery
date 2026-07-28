package cli

import "testing"

func TestDispLen(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"空文字は幅0として扱う", "", 0},
		{"判定上限の最大値", string(rune(0x07F)), 1},
		{"判定条件の最小上界", string(rune(0x080)), 2},
		{"半角カタカナは幅2として扱う", "ｧ", 2},
		{"半角カタカナ,アルファベットの複合", "aｧb", 4},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := dispLen(tt.in)
			if got != tt.want {
				t.Errorf("dispLen(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestPad(t *testing.T) {
	cases := []struct {
		name string
		s    string
		w    int
		want string
	}{
		{"境界の最大値", "a", 1, "a"},
		{"境界の最小上界", "a", 2, "a "},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := pad(tt.s, tt.w)
			if got != tt.want {
				t.Errorf("pad(%q,%d) = %q, want %q", tt.s, tt.w, got, tt.want)
			}
		})
	}
}
