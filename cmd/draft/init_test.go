package main

import (
	"errors"
	"testing"
)

func TestIsAlreadyExists(t *testing.T) {
	tests := []struct {
		descrip  string
		err      error
		expected bool
	}{
		{"Should not match", errors.New("a release named \"Foo1 baR2_v4.0\" already exists"), false},
		{"Should match", errors.New("a release named \"Foo1-baR1_v4.0\" already exists"), true},
	}

	for _, test := range tests {
		got := IsReleaseAlreadyExists(test.err)
		if got != test.expected {
			t.Errorf("%s: Expected %v, got %v", test.descrip, test.expected, got)
		}
	}

}
