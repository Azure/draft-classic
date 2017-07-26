package installer // import "github.com/Azure/draft/pkg/plugin/installer"

import (
	"os"
	"path/filepath"

	"github.com/Azure/draft/pkg/draft/draftpath"
)

type base struct {
	// Source is the reference to a plugin
	Source string

	//TODO: change DraftHome to ProjectHome or something more generic
	// or add ProjectHomeEnv as an attribute

	// DraftHome is the $DRAFT_HOME directory
	DraftHome draftpath.Home
}

func newBase(source string, home draftpath.Home) base {
	return base{source, home}
}

// link creates a symlink from the plugin source to $DRAFT_HOME
func (b *base) link(from string) error {
	//debug("symlinking %s to %s", from, b.Path())
	return os.Symlink(from, b.Path())
}

// Path is where the plugin will be symlinked to.
func (b *base) Path() string {
	if b.Source == "" {
		return ""
	}
	return filepath.Join(b.DraftHome.Plugins(), filepath.Base(b.Source))
}
