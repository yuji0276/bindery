> **This document is a translation. It is not the source of truth.**
> The authoritative version is the Japanese [design.md](design.md).
> If the two disagree, the Japanese version wins.

## Data model

- action: what operation is performed
- mode: holds e.g. nvim's insert, zellij's tab
- tool: the name of the tool
- key: the assigned key binding

Fields such as column number and filepath will be added later so that features like jump-to-definition can be built on top; the model is meant to be extensible. We start from the minimum required for conflict detection.

### Value domain (type) of the `key` column

The `key` column can be seen as a map from the set of rows to a single value. This section describes the value domain (type) that `key` can take. The following have to be accounted for.

- Chord (pressed simultaneously) or sequence (pressed one after another)
- Whether a binding carries more than one modifier key

With these in mind, the value domain is built in three stages.

M := the vocabulary of modifier keys (Ctrl, Alt, Cmd, ...; how they are spelled is undecided)
B := the vocabulary of base keys (t, =, left, ...; how they are spelled is undecided)

1. A combination of modifier keys (the power set of M)
2. One keystroke (a modifier combination together with a base key; the Cartesian product of the power set of M and B)
3. One binding (a non-empty finite sequence of keystrokes)

Examples:

|Input|As an element of the value domain|
|---|---|
|`Ctrl+Shift+t`|sequence of length 1 `< ({Ctrl,Shift},t) >`|
|`<leader>ff`|sequence of length 3 `< (∅,leader),(∅,f),(∅,f) >`|
|wezterm `SUPER` + `=`|sequence of length 1 `< ({Super},=) >`|
|zellij `Ctrl+t` -> `n`|sequence of length 2 `< ({Ctrl},t),(∅,n) >`|

Why the empty sequence is not allowed: length 0 would mean pressing nothing, which is not meaningful as a key.

### Conflict detection

There are two conditions for a conflict. Two bindings conflict when both conditions hold.

- Whether the two modes can be active at the same time
- Whether one binding's key is a prefix of the other's

`tool` does not appear in the conflict conditions themselves. It is used only inside the decision of whether two modes can be active at the same time: there we ask whether the two modes belong to the same tool, and in what order the tools are nested.

#### Whether the two modes can be active at the same time

This is decided by whether the modes of the two bindings can be active at the same time.

The table below is a check on C, defined further down. All three rows must be derivable from C.

|Pair|Conflict?|Can be active at the same time?|
|--|--|--|
|(zellij tab, nvim normal)|no conflict|mutually exclusive|
|(nvim normal, nvim insert)|no conflict|mutually exclusive|
|(nvim normal, zellij normal)|conflict|**can be active at the same time**|

##### Deciding co-activity with bindery.conf

bindery.conf records the transparency of each tool's modes and how the tools are nested.

The nesting is placed in a configuration file because it is not a fixed value but a declaration about the user's environment. Some setups open nvim directly in wezterm without zellij; over ssh the nesting can differ again.

Let C(m1, m2) be the deciding function, where m1 and m2 are the modes held by the two bindings being compared.

1. m1 and m2 belong to the same tool -> mutually exclusive
2. m1 and m2 belong to different tools -> look at the nesting, then look at the transparency of the **mode** of the outer tool. Transparent means they can be active at the same time; not transparent means mutually exclusive.

#### Whether one binding's key is a prefix of the other's

This condition is needed because `<leader>ff` and `<leader>f` conflict.

An exact match is a prefix of equal length, so it is covered by this condition. The case where the same key is bound in two different tools is caught here as well.

## Open Question

- Value domain of the `action` column (normalizing meaning)
- Difference in severity between an exact match and a prefix (timeoutlen)
- `Shift+a` versus `A`
- DisableDefaultAssignment
- CSI u / Kitty keyboard protocol
- wezterm's key_tables and leader
- How to choose the vocabularies for base keys and modifier keys (a or A, Cmd or CMD, ...)
- With three or more layers, is it acceptable that the modes of the layers in between do not affect the decision? (for a pair of a wezterm mode and an nvim mode, zellij sits in between)
