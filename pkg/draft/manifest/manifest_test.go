package manifest

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	m := New()
	expected := "&{   default [] false false 2}"
	actual := fmt.Sprintf("%v", m.Environments[DefaultEnvironmentName])
	if expected != actual {
		t.Errorf("wanted %s, got %s", expected, actual)
	}
}
