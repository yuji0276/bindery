<h1 align="center">bindery</h1>

<p align="center">
  Aggregate every keybinding from your <a href="https://www.chezmoi.io/">chezmoi</a>-managed dotfiles into a single, searchable list.
</p>

<p align="center">
  <img alt="Go" src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white">
  <img alt="Platform" src="https://img.shields.io/badge/platform-macOS-lightgrey">
  <img alt="License" src="https://img.shields.io/badge/license-MIT-green">
</p>

`bindery` is a small Go CLI that scans the config files managed by chezmoi, parses
the keybindings out of each one, and shows them all together — in an interactive
`fzf` picker, a plain table, Markdown, or JSON.

Currently supported sources: **skhd** · **Neovim** · **WezTerm**
(karabiner / yabai are not yet parsed because they carry no key assignments today —
add a parser to `internal/sources/` when that changes.)

> 🇯🇵 日本語の説明は [下部](#日本語) にあります。

## Features

- 🔍 **One list for every tool** — skhd, Neovim, and WezTerm bindings side by side.
- ⚡ **Interactive search** — fuzzy-find with `fzf`, then hit <kbd>Enter</kbd> to jump straight to the definition in `$EDITOR`.
- 📄 **Multiple outputs** — interactive picker, plain table, Markdown, or JSON.
- 🎯 **chezmoi-aware** — only parses files that chezmoi actually manages (falls back to file existence when chezmoi is absent).
- 🧠 **Real Neovim introspection** — launches headless Neovim and reads the _live_ keymaps via `nvim_get_keymap()`, including `desc`, source file, and line number.

## Installation

```sh
go build -o ~/.local/bin/bindery .
```

Make sure `~/.local/bin` is on your `PATH`. Installing [`fzf`](https://github.com/junegunn/fzf)
is optional but unlocks interactive search.

Requirements: **Go 1.25+**, macOS. `chezmoi`, `fzf`, and `nvim` are optional and enable extra behavior when present.

## Usage

```sh
bindery            # interactive fzf search (falls back to a table if fzf is missing)
bindery -l         # always print a table
bindery -l | less  # pipe to a pager
bindery --md       # emit Markdown (handy for generating keybind.md)
bindery --json     # emit JSON
bindery -s skhd    # filter by source (skhd | nvim | wezterm)
bindery -h         # help
```

In `fzf` mode, narrow the list and press <kbd>Enter</kbd> to open the file
containing that binding in `$EDITOR` (defaults to `nvim`).

### Generate a keybind.md

```sh
bindery --md > ~/.config/keybind.md
```

## How it works

1. Run `chezmoi managed --path-style=absolute` to get the set of managed files.
2. Parse only the known config files (skhd / nvim / wezterm) that are **under chezmoi management**.
   In an environment without chezmoi, presence of the file is used instead.
3. Extract key, modifiers, description, and source line with a format-specific parser.

### Supported formats

| Source  | File                   | Extraction                                                                                                                                                                                                                                                                                                                                              |
| ------- | ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| skhd    | `skhd/skhdrc`          | `<mods> - <key> : <command>` form; trailing `# comment` becomes the description.                                                                                                                                                                                                                                                                        |
| nvim    | (live Neovim)          | Runs `nvim --headless` and dumps every active mapping via `nvim_get_keymap()` as JSON — capturing `desc`, source file, and line number. Mode is shown in the Mods column. String-`rhs` maps lack an `lnum`, so `desc` is grepped against `~/.config/nvim` to recover `file:line`. Falls back to regex-parsing `keymaps.lua` when Neovim is unavailable. |
| wezterm | `wezterm/keybinds.lua` | Extracts `key =` / `mods =` from the Lua table; the preceding `--` comment becomes the description.                                                                                                                                                                                                                                                     |

## Project layout

```
main.go                     # CLI shell: flag parsing and wiring — Collect(sources.List(), filter)
internal/bindery/           # domain types (leaf, unit-tested)
  binding.go                #   Binding / Source types, Sort
  collect.go                #   chezmoi integration, Collect()
internal/sources/           # per-source parsers
  sources.go                #   List() — the source registry
  skhd.go                   #   parseSkhd
  nvim.go                   #   parseNvim (live Neovim) + regex fallback
  wezterm.go                #   parseWezterm
internal/cli/               # presentation
  render.go                 #   table / Markdown / JSON output
  fzf.go                    #   fzf interaction + $EDITOR launch
```

Dependency direction: `main` → `internal/sources` → `internal/bindery` ← `internal/cli`.
`bindery` is a type leaf that never imports `sources`, so there are no cycles.

## Adding a parser

Implement `func parseXxx(path string) ([]bindery.Binding, error)` in `internal/sources/`,
then add one line to `List()` in `internal/sources/sources.go`. That's it.

## Contributing

Issues and pull requests are welcome. Please run the checks before submitting:

```sh
go test ./...
go vet ./...
```

## License

Released under the [MIT License](LICENSE).

---

## 日本語

chezmoi 管理下の dotfile からキーバインド定義を集約して一覧表示する Go 製 CLI です。

対応ソース: **skhd** / **Neovim** / **WezTerm**
（karabiner / yabai は現状キー割り当てが無いため未対応。定義が増えたら `internal/sources/` にパーサを追加します）

### インストール

```sh
go build -o ~/.local/bin/bindery .
```

`~/.local/bin` が PATH に含まれている前提。fzf があるとインタラクティブ検索が使えます。

### 使い方

```sh
bindery            # fzf でインタラクティブ検索 (fzf 未導入なら表を表示)
bindery -l         # 常に表で一覧表示
bindery -l | less  # ページャに流す
bindery --md       # Markdown で出力 (keybind.md 生成用)
bindery --json     # JSON で出力
bindery -s skhd    # ソースで絞り込み (skhd|nvim|wezterm)
bindery -h         # ヘルプ
```

fzf モードでは、絞り込んで **Enter** を押すとその定義がある行を `$EDITOR`（既定 `nvim`）で開きます。

### 仕組み

1. `chezmoi managed --path-style=absolute` で管理対象ファイルの集合を取得
2. 既知の設定ファイル（skhd/nvim/wezterm）のうち **chezmoi 管理下のものだけ** をパース
   （chezmoi が無い環境ではファイルの存在のみで判断）
3. 各フォーマット固有のパーサでキー・修飾・説明・定義行を抽出

パーサの追加は、`internal/sources/` に `func parseXxx(path string) ([]bindery.Binding, error)` を実装し、
`internal/sources/sources.go` の `List()` に 1 行追加するだけです。
