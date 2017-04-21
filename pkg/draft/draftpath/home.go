package draftpath

import (
	"path/filepath"
)

// Home describes the location of a CLI configuration.
//
// This helper builds paths relative to a Draft Home directory.
type Home string

// String returns Home as a string.
//
// Implements fmt.Stringer.
func (h Home) String() string {
	return string(h)
}

// Packs returns the path to the Draft starter packs.
func (h Home) Packs() string {
	return filepath.Join(string(h), "packs")
}

// Plugins returns the path to the Draft plugins.
func (h Home) Plugins() string {
	return filepath.Join(string(h), "plugins")
}
