package installer

import (
	"os"
	"path/filepath"

	"github.com/Azure/draft/pkg/plugin"
	"github.com/Azure/draft/pkg/plugin/repository"
)

// Installer provides an interface for installing client rigs.
type Installer interface {
	// Install adds a rig to a path
	Install() error
	// Path is the directory of the installed plugin.
	Path() string
	// Update updates a plugin.
	Update() error
}

// Install installs a plugin.
func Install(i Installer) error {
	basePath := filepath.Dir(i.Path())
	if _, pathErr := os.Stat(basePath); os.IsNotExist(pathErr) {
		if err := os.MkdirAll(basePath, 0755); err != nil {
			return err
		}
	}

	if _, pathErr := os.Stat(i.Path()); !os.IsNotExist(pathErr) {
		return i.Update()
	}

	return i.Install()
}

// Update updates a plugin.
func Update(i Installer) error {
	if _, pathErr := os.Stat(i.Path()); os.IsNotExist(pathErr) {
		return repository.ErrDoesNotExist
	}

	return i.Update()
}

// FindSource determines the correct Installer for the given source.
func FindSource(location string, home plugin.Home) (Installer, error) {
	installer, err := existingVCSRepo(location, home)
	if err != nil && err.Error() == "Cannot detect VCS" {
		return installer, repository.ErrMissingSource
	}
	return installer, err
}

// New determines and returns the correct Installer for the given source
func New(source, version string, home plugin.Home) (Installer, error) {
	if isLocalReference(source) {
		return NewLocalInstaller(source, home)
	}

	return NewVCSInstaller(source, version, home)
}

// isLocalReference checks if the source exists on the filesystem.
func isLocalReference(source string) bool {
	_, err := os.Stat(source)
	return err == nil
}

// isRig checks if the directory contains a "Plugins" directory.
func isRig(dirname string) bool {
	_, err := os.Stat(filepath.Join(dirname, "Plugins"))
	return err == nil
}
