package prowpath

import (
	"testing"
)

func TestProwHome(t *testing.T) {
	ph := Home("r:\\")
	isEq := func(t *testing.T, a, b string) {
		if a != b {
			t.Errorf("Expected %q, got %q", b, a)
		}
	}

	isEq(t, ph.String(), "r:\\")
	isEq(t, ph.Starters(), "r:\\starters")
}
