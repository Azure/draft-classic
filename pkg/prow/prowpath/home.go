package prowpath

import (
	"path/filepath"
)

// Home describes the location of a CLI configuration.
//
// This helper builds paths relative to a Prow Home directory.
type Home string

// String returns Home as a string.
//
// Implements fmt.Stringer.
func (h Home) String() string {
	return string(h)
}

// Packs returns the path to the Prow starter packs.
func (h Home) Packs() string {
	return filepath.Join(string(h), "packs")
}
