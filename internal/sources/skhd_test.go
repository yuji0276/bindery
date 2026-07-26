package sources

import (
	"bindery/internal/bindery"
	"os"
	"path/filepath"
	"testing"
)

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

func TestParseSkhd(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []bindery.Binding
	}{
		{
			"one line with mods",
			"cmd -h: yabai -m window --focus west",
			[]bindery.Binding{
				{Source: "skhd", Mods: "cmd", Key: "h", Desc: "yabai -m window --focus west", Line: 1},
			},
		},
		{
			"ignore empty line and comment line",
			"# comment\n\ncmd -h: echo hi\n",
			[]bindery.Binding{
				{Source: "skhd", Mods: "cmd", Key: "h", Desc: "echo hi", Line: 3},
			},
		},
		{
			"end line comment becomes Desc",

			"cmd + shift -j: echo hi # フォーカスを下へ",
			[]bindery.Binding{
				{Source: "skhd", Mods: "cmd+shift", Key: "j", Desc: "フォーカスを下へ", Line: 1},
			},
		},
		{
			"ignore line without colon",
			"this is not a keybind\ncmd -h: echo hi\n",
			[]bindery.Binding{
				{Source: "skhd", Mods: "cmd", Key: "h", Desc: "echo hi", Line: 2},
			},
		},
		{
			"ignore line with empty keydef",
			": echo hi\n",
			nil,
		},
		{
			"ignore mode declaration and directive",
			":: default\n.blacklist [\n\"terminal\"\n]\ncmd -h: echo hi\n",
			[]bindery.Binding{
				{Source: "skhd", Mods: "cmd", Key: "h", Desc: "echo hi", Line: 5},
			},
		},
		{
			"no mods",
			"f1 : echo hi",
			[]bindery.Binding{
				{Source: "skhd", Mods: "", Key: "f1", Desc: "echo hi", Line: 1},
			},
		},
		{
			"multiple hash marks",
			"cmd - h : echo # a # b",
			[]bindery.Binding{
				{Source: "skhd", Mods: "cmd", Key: "h", Desc: "a # b", Line: 1},
			},
		},
		{
			"comment only command",
			"cmd - h : # コメントのみ",
			[]bindery.Binding{
				{Source: "skhd", Mods: "cmd", Key: "h", Desc: "コメントのみ", Line: 1},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "skhdrc")
			if err := os.WriteFile(path, []byte(tt.in), 0o600); err != nil {
				t.Fatal(err)
			}

			want := make([]bindery.Binding, len(tt.want))
			for i, b := range tt.want {
				b.File = path
				want[i] = b
			}

			got, err := parseSkhd(path)
			if err != nil {
				t.Fatalf("parseSkhd(%q) returned error: %v", path, err)
			}
			if len(got) != len(want) {
				t.Fatalf("parseSkhd() = %+v (len %d), want %+v (len %d)", got, len(got), want, len(want))
			}
			for i := range want {
				if got[i] != want[i] {
					t.Errorf("parseSkhd()[%d] = %+v, want %+v", i, got[i], want[i])
				}
			}
		})
	}
}

func TestParseSkhdOpenError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notexist")

	got, err := parseSkhd(path)
	if err == nil {
		t.Fatalf("parseSkhd(%q) error = nil, want error", path)
	}
	if got != nil {
		t.Errorf("parseSkhd(%q) = %+v, want nil", path, got)
	}
}
