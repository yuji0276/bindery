package main

// truncateString returns s cut to at most max runes. If truncated, an
// ellipsis is appended so the result never exceeds max runes.
func truncateString(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}
