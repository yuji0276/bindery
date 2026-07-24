package bindery

import (
	"reflect"
	"testing"
)

func TestSort(t *testing.T) {
	cases := []struct {
		name string
		in   []Binding
		want []string
	}{
		{"同一のソース", []Binding{{"nvim", "cmd", "a", "test", "user/test.lua", 3}, {"nvim", "cmd", "b", "test", "user/test.lua", 2}, {"nvim", "cmd", "c", "test", "user/test.lua", 1}}, []string{"c", "b", "a"}},
		{"異なるソース", []Binding{{"nvim", "cmd", "a", "test", "user/test.lua", 3}, {"wezterm", "cmd", "b", "test", "user/test.lua", 2}, {"nvim", "cmd", "c", "test", "user/test.lua", 1}}, []string{"c", "a", "b"}},
		{"check stability", []Binding{{"nvim", "cmd", "a", "test", "user/test.lua", 3}, {"nvim", "cmd", "b", "test", "user/test.lua", 3}}, []string{"a", "b"}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			Sort(tt.in)
			got := make([]string, 0, len(tt.in))
			for i := 0; i < len(tt.in); i++ {
				got = append(got, tt.in[i].Key)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sort(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
