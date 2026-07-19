# keybinds

chezmoi 管理下の dotfile からキーバインド定義を集約して一覧表示する Go 製 CLI。

対応ソース: **skhd** / **Neovim** / **WezTerm**
（karabiner / yabai は現状キー割り当てが無いため未対応。定義が増えたら `parsers.go` にパーサを追加する）

## インストール

```sh
cd ~/.config/keybinds
go build -o ~/.local/bin/keybinds .
```

`~/.local/bin` が PATH に含まれている前提。fzf があるとインタラクティブ検索が使える。

## 使い方

```sh
keybinds            # fzf でインタラクティブ検索 (fzf 未導入なら表を表示)
keybinds -l         # 常に表で一覧表示
keybinds -l | less  # ページャに流す
keybinds --md       # Markdown で出力 (keybind.md 生成用)
keybinds --json     # JSON で出力
keybinds -s skhd    # ソースで絞り込み (skhd|nvim|wezterm)
keybinds -h         # ヘルプ
```

fzf モードでは、絞り込んで **Enter** を押すとその定義がある行を `$EDITOR`（既定 `nvim`）で開く。

## keybind.md を自動生成する

```sh
keybinds --md > ~/.config/keybind.md
```

## 仕組み

1. `chezmoi managed --path-style=absolute` で管理対象ファイルの集合を取得
2. 既知の設定ファイル（skhd/nvim/wezterm）のうち **chezmoi 管理下のものだけ** をパース
   - chezmoi が無い環境ではファイルの存在のみで判断
3. 各フォーマット固有のパーサでキー・修飾・説明・定義行を抽出

## 対応フォーマット

| ソース | ファイル | 抽出方法 |
|--------|----------|----------|
| skhd | `skhd/skhdrc` | `<mods> - <key> : <command>` 形式。行末 `# コメント` を説明に |
| nvim | `nvim/lua/config/keymaps.lua` | `map("mode", "key", ..., { desc = "..." })` を正規表現で抽出。Mods 列にモードを表示 |
| wezterm | `wezterm/keybinds.lua` | Lua テーブルの `key = / mods =` を抽出。直前の `--` コメントを説明に |

## パーサを追加する

`parsers.go` に `func parseXxx(path string) ([]Binding, error)` を実装し、
`main.go` の `sources()` に 1 行追加するだけ。
