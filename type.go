package bindery

type ShortCutKeys struct {
	keyStrokes []OneKeyStroke
}

type OneKeyStroke struct {
	modifierKey Modifiers
	baseKey     string
}

type Modifiers int

const (
	ctrl Modifiers = 1 << iota
	shift
	cmd
	//修飾キーの選択、揺れについては後に考える
)
