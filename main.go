package main

import (
	"fmt"
	"os"
	"os/exec"

	"bindery/internal/bindery"
	"bindery/internal/cli"
)

func main() {
	var (
		list   bool
		asMD   bool
		asJSON bool
		noFzf  bool
		filter string
	)
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-l", "--list":
			list = true
		case "--md", "--markdown":
			asMD = true
		case "--json":
			asJSON = true
		case "--no-fzf":
			noFzf = true
		case "-s", "--source":
			if i+1 < len(args) {
				i++
				filter = args[i]
			}
		case "-h", "--help":
			printHelp()
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n\n", args[i])
			printHelp()
			os.Exit(2)
		}
	}

	bindings, warnings := bindery.Collect(filter)
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}
	if len(bindings) == 0 {
		fmt.Fprintln(os.Stderr, "キーバインドが見つかりませんでした。")
		os.Exit(1)
	}

	bindery.Sort(bindings)

	switch {
	case asJSON:
		cli.JSON(os.Stdout, bindings)
	case asMD:
		cli.Markdown(os.Stdout, bindings)
	case list || noFzf || !stdoutIsTTY() || !hasFzf():
		cli.Table(os.Stdout, bindings)
	default:
		cli.RunFzf(bindings)
	}
}

func stdoutIsTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func hasFzf() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

func printHelp() {
	fmt.Print(`bindery — chezmoi 管理下の dotfile からキーバインド一覧を集約

USAGE:
  bindery [flags]

FLAGS:
  (なし)          fzf でインタラクティブ検索 (fzf 未導入なら表を表示)
  -l, --list      常に表で一覧表示
  --md            Markdown で出力 (keybind.md 生成用)
  --json          JSON で出力
  -s, --source X  ソースで絞り込み (skhd|nvim|wezterm)
  -h, --help      このヘルプ

EXAMPLES:
  bindery
  bindery -l | less
  bindery --md > ~/.config/keybind.md
  bindery -s skhd
`)
}
