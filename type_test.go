package bindery

import (
	"testing"
)

func TestOneKeyStrokeLiteral(t *testing.T) {
	// ここに t と Ctrl+Shift+t の2つを OneKeyStroke の値として書く。
	// 書けたら、意図した通りかを == で確かめて t.Errorf で報告する。
	key1 := OneKeyStroke{
		modifierKey: ctrl | shift,
		baseKey:     "t",
	}
	key2 := OneKeyStroke{
		baseKey: "t",
	}

	if key1.modifierKey != ctrl|shift {
		t.Errorf("Expected 11, but got %d", key1.modifierKey)
	}
	if key2.modifierKey != 0 {
		t.Errorf("Expected 0, but got %d", key2.modifierKey)
	}
}
