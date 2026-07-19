// Package bindery は chezmoi 管理下の dotfile からキーバインド定義を収集する
// ドメインロジックを提供する。CLI 表示層 (internal/cli) はこのパッケージに依存する。
package bindery

import "sort"

// Binding は 1 つのキー割り当てを表す。
type Binding struct {
	Source string `json:"source"` // skhd / nvim / wezterm
	Mods   string `json:"mods"`   // 修飾キー (nvim ではモード)
	Key    string `json:"key"`    // キー
	Desc   string `json:"desc"`   // 説明 (コメント or コマンド)
	File   string `json:"file"`   // 定義元ファイル (絶対パス)
	Line   int    `json:"line"`   // 定義元の行番号
}

// source は 1 つの設定ソースと、そのパーサを結びつける。
type source struct {
	name  string
	path  string // ~ 展開済みの既定パス
	parse func(path string) ([]Binding, error)
}

// Sort はソース名→行番号の順で安定ソートする。
func Sort(bs []Binding) {
	sort.SliceStable(bs, func(i, j int) bool {
		if bs[i].Source != bs[j].Source {
			return bs[i].Source < bs[j].Source
		}
		return bs[i].Line < bs[j].Line
	})
}
