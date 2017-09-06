package installer

import (
	"path/filepath"

	"github.com/Azure/draft/pkg/draft/draftpath"
)

// LocalInstaller installs pack repos from the filesystem
type LocalInstaller struct {
	base
}

// NewLocalInstaller creates a new LocalInstaller
func NewLocalInstaller(source string, home draftpath.Home) (*LocalInstaller, error) {

	i := &LocalInstaller{
		base: newBase(source, home),
	}

	return i, nil
}

// Install creates a symlink to the pack repo directory in $DRAFT_HOME
func (i *LocalInstaller) Install() error {
	if !isPackRepo(i.Source) {
		return ErrMissingPackDir
	}

	src, err := filepath.Abs(i.Source)
	if err != nil {
		return err
	}

	return i.link(src)
}

// Update updates a local repository
func (i *LocalInstaller) Update() error {
	return nil
}
