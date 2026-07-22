# bindery

chezmoi 管理下の dotfile からキーバインド定義を集約して一覧表示する Go 製 CLI。

対応ソース: **skhd** / **Neovim** / **WezTerm**
（karabiner / yabai は現状キー割り当てが無いため未対応。定義が増えたら `parsers.go` にパーサを追加する）

## インストール

```sh
cd ~/.config/keybinds
go build -o ~/.local/bin/bindery .
```

`~/.local/bin` が PATH に含まれている前提。fzf があるとインタラクティブ検索が使える。

## 使い方

```sh
bindery            # fzf でインタラクティブ検索 (fzf 未導入なら表を表示)
bindery -l         # 常に表で一覧表示
bindery -l | less  # ページャに流す
bindery --md       # Markdown で出力 (keybind.md 生成用)
bindery --json     # JSON で出力
bindery -s skhd    # ソースで絞り込み (skhd|nvim|wezterm)
bindery -h         # ヘルプ
```

fzf モードでは、絞り込んで **Enter** を押すとその定義がある行を `$EDITOR`（既定 `nvim`）で開く。

## keybind.md を自動生成する

```sh
bindery --md > ~/.config/keybind.md
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
| nvim | (nvim 実機) | `nvim --headless` を起動し `nvim_get_keymap()` で**実際に効いている**全モードのマッピングを JSON 出力させて取り込む。desc・定義元ファイル・行番号まで取得。Mods 列にモードを表示。文字列 rhs のマップは lnum を取れないため、desc で設定ツリー(`~/.config/nvim`)を grep して file:line を補完する。nvim 未導入時は keymaps.lua の正規表現パースにフォールバック |
| wezterm | `wezterm/keybinds.lua` | Lua テーブルの `key = / mods =` を抽出。直前の `--` コメントを説明に |

## 構成

```
main.go                     # CLI の殻: フラグ処理と配線 (Collect(sources.List(), filter))
internal/bindery/           # ドメイン型 (リーフ・テスト対象)
  binding.go                #   Binding 型・Source 型・Sort
  collect.go                #   chezmoi 連携・Collect()
internal/sources/           # ソース別パーサ
  sources.go                #   List() — 供給元リスト
  skhd.go                   #   parseSkhd
  nvim.go                   #   parseNvim (nvim 実機) + 正規表現フォールバック
  wezterm.go                #   parseWezterm
internal/cli/               # 表示層
  render.go                 #   表 / Markdown / JSON 出力
  fzf.go                    #   fzf 対話・$EDITOR 起動
```

依存の向きは `main` → `internal/sources` → `internal/bindery` ←  `internal/cli`。
`bindery` は型のリーフで `sources` に依存しないため循環しない。

## パーサを追加する

`internal/sources/` に `func parseXxx(path string) ([]bindery.Binding, error)` を実装し、
同 `internal/sources/sources.go` の `List()` に 1 行追加するだけ。
