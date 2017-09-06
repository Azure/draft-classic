package draftpath

import (
	"os"
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
	return os.ExpandEnv(string(h))
}

// Path returns Home with elements appended.
func (h Home) Path(elem ...string) string {
	p := []string{h.String()}
	p = append(p, elem...)
	return filepath.Join(p...)
}

// Packs returns the path to the Draft starter packs.
func (h Home) Packs() string {
	return h.Path("packs")
}

// Plugins returns the path to the Draft plugins.
func (h Home) Plugins() string {
	return h.Path("plugins")
}
