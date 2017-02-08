package prowpath

import (
	"runtime"
	"testing"
)

func TestProwHome(t *testing.T) {
	ph := Home("/r")
	isEq := func(t *testing.T, a, b string) {
		if a != b {
			t.Error(runtime.GOOS)
			t.Errorf("Expected %q, got %q", a, b)
		}
	}

	isEq(t, ph.String(), "/r")
	isEq(t, ph.Packs(), "/r/packs")
}
