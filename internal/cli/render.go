// Package cli は bindery が収集した Binding を各種フォーマットで出力する表示層。
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"bindery/internal/bindery"
)

// JSON は整形済み JSON を書き出す。
func JSON(w io.Writer, bs []bindery.Binding) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(bs)
}

// Table は tabwriter で桁揃えした表を書き出す。
func Table(w io.Writer, bs []bindery.Binding) {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "SOURCE\tMODS\tKEY\tDESCRIPTION")
	for _, b := range bs {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", b.Source, b.Mods, b.Key, b.Desc)
	}
	tw.Flush()
}

// Markdown はソースごとに見出し + 表で書き出す。
func Markdown(w io.Writer, bs []bindery.Binding) {
	fmt.Fprintln(w, "# キーバインディング一覧 (自動生成)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "`bindery --md` で生成。手で編集しないでください。")
	fmt.Fprintln(w)
	cur := ""
	for _, b := range bs {
		if b.Source != cur {
			cur = b.Source
			fmt.Fprintf(w, "\n## %s\n\n", b.Source)
			fmt.Fprintln(w, "| Mods | Key | 説明 |")
			fmt.Fprintln(w, "|------|-----|------|")
		}
		fmt.Fprintf(w, "| %s | `%s` | %s |\n", mdEsc(b.Mods), b.Key, mdEsc(b.Desc))
	}
}

func mdEsc(s string) string { return strings.ReplaceAll(s, "|", "\\|") }
