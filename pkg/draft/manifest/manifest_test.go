package manifest

import (
	"fmt"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	m := New()
	m.Environments[DefaultEnvironmentName].Name = "foobar"
	expected := "&{foobar   default [] false true 2}"
	actual := fmt.Sprintf("%v", m.Environments[DefaultEnvironmentName])
	if expected != actual {
		t.Errorf("wanted %s, got %s", expected, actual)
	}
}

func TestGenerateName(t *testing.T) {
	name := generateName()
	if name == "" {
		t.Error("expected name to be generated")
	}
	if !strings.Contains(name, "-") {
		t.Errorf("expected dash in name, got %s", name)
	}
}
