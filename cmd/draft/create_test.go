package main

import (
	"strings"
	"testing"
)

func TestGenerateName(t *testing.T) {
	name := generateName()
	if name == "" {
		t.Error("expected name to be generated")
	}
	if !strings.Contains(name, "-") {
		t.Errorf("expected dash in name, got %s", name)
	}
}
