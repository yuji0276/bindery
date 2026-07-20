package bindery

import "sort"

type Binding struct {
	Source string `json:"source"`
	Mods   string `json:"mods"`
	Key    string `json:"key"`
	Desc   string `json:"desc"`
	File   string `json:"file"`
	Line   int    `json:"line"`
}

type source struct {
	name  string
	path  string
	parse func(path string) ([]Binding, error)
}

func Sort(bs []Binding) {
	sort.SliceStable(bs, func(i, j int) bool {
		if bs[i].Source != bs[j].Source {
			return bs[i].Source < bs[j].Source
		}
		return bs[i].Line < bs[j].Line
	})
}
