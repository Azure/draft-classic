package prowpath

import "path/filepath"

// Home describes the location of a CLI configuration.
//
// This helper builds paths relative to a Helm Home directory.
type Home string

// String returns Home as a string.
//
// Implements fmt.Stringer.
func (h Home) String() string {
	return string(h)
}

// Plugins returns the path to the plugins directory.
func (h Home) Plugins() string {
	return filepath.Join(string(h), "plugins")
}
