package plugin

import (
	"path/filepath"
)

type (
	// Home designates where plugins should store data.
	Home string
)

// Path returns Home with elements appended.
func (h Home) Path(elem ...string) string {
	p := []string{h.String()}
	p = append(p, elem...)
	return filepath.Join(p...)
}

// Installed returns the path to the directory where plugins are installed to.
func (h Home) Installed() string {
	return h.Path("installed")
}

// Repositories returns the path to the fishing Repositories.
func (h Home) Repositories() string {
	return h.Path("repositories")
}

// Cache returns the path to the plugin cache.
func (h Home) Cache() string {
	return h.Path("cache")
}

// String returns Home as a string.
//
// Implements fmt.Stringer.
func (h Home) String() string {
	return string(h)
}
