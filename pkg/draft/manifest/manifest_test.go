package manifest

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewCustomName(t *testing.T) {
	m := New("foobar")
	expected := "&{foobar   default [] false false 2}"
	actual := fmt.Sprintf("%v", m.Environments[DefaultEnvironmentName])
	if expected != actual {
		t.Errorf("wanted %s, got %s", expected, actual)
	}
}

func TestNewDefault(t *testing.T) {
	m := New("")

	if m.Environments[DefaultEnvironmentName].Name == "" {
		t.Errorf("expected name to be generated")
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
