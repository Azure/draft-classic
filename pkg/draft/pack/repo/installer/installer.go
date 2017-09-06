package installer

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/plugin/installer"
)

// ErrMissingPackDir indicates that the packs dir is missing.
var ErrMissingPackDir = errors.New("packs dir missing or not found")

type base struct {
	// Source is the reference to a pack repo
	Source string

	// DraftHome is the $DRAFT_HOME directory
	DraftHome draftpath.Home
}

func newBase(source string, home draftpath.Home) base {
	return base{source, home}
}

// link creates a symlink from the pack repo source to $DRAFT_HOME
func (b *base) link(from string) error {
	//debug("symlinking %s to %s", from, b.Path())
	return os.Symlink(from, b.Path())
}

// Path is where the pack repo will be symlinked to.
func (b *base) Path() string {
	if b.Source == "" {
		return ""
	}
	return filepath.Join(b.DraftHome.Packs(), filepath.Base(b.Source))
}

// isPackRepo checks if the directory contains a packs directory.
func isPackRepo(dirname string) bool {
	fi, err := os.Stat(filepath.Join(dirname, "packs"))
	return err == nil && fi.IsDir()
}

// isLocalReference checks if the source exists on the filesystem.
func isLocalReference(source string) bool {
	_, err := os.Stat(source)
	return err == nil
}

// Install installs a pack repo to $DRAFT_HOME
func Install(i installer.Installer) error {
	if _, pathErr := os.Stat(path.Dir(i.Path())); os.IsNotExist(pathErr) {

		return fmt.Errorf("pack home %s does not exist", path.Dir(i.Path()))
	}

	if _, pathErr := os.Stat(i.Path()); !os.IsNotExist(pathErr) {
		return errors.New("pack repo already exists")
	}

	return i.Install()
}

// Update updates a pack repo in $DRAFT_HOME.
func Update(i installer.Installer) error {
	if _, pathErr := os.Stat(i.Path()); os.IsNotExist(pathErr) {
		return errors.New("pack repo does not exist")
	}

	return i.Update()
}

// FindSource determines the correct Installer for the given source.
func FindSource(location string, home draftpath.Home) (installer.Installer, error) {
	installer, err := existingVCSRepo(location, home)
	if err != nil && err.Error() == "Cannot detect VCS" {
		return installer, errors.New("cannot get information about pack repo source")
	}
	return installer, err
}

// New determines and returns the correct Installer for the given source
func New(source, version string, home draftpath.Home) (installer.Installer, error) {
	if isLocalReference(source) {
		return NewLocalInstaller(source, home)
	}

	return NewVCSInstaller(source, version, home)
}
