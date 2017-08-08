package installer

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/Azure/draft/pkg/draft/draftpath"
)

// special thanks to the Kubernets Helm plugin installer pkg

// ErrMissingMetadata indicates that plugin.yaml is missing.
var ErrMissingMetadata = errors.New("plugin metadata (plugin.yaml) missing")

// Debug enables verbose output.
var Debug bool

// Installer provides an interface for installing client plugins.
type Installer interface {
	// Install adds a plugin to a path
	Install() error
	// Path is the directory of the installed plugin.
	Path() string
	// Update updates a plugin to $DRAFT_HOME.
	Update() error
}

// Install installs a plugin to $DRAFT_HOME
func Install(i Installer) error {
	if _, pathErr := os.Stat(path.Dir(i.Path())); os.IsNotExist(pathErr) {

		return errors.New(`plugin home "$DRAFT_HOME/plugins" does not exist`)
	}

	if _, pathErr := os.Stat(i.Path()); !os.IsNotExist(pathErr) {
		return errors.New("plugin already exists")
	}

	return i.Install()
}

// Update updates a plugin in $DRAFT_HOME.
func Update(i Installer) error {
	if _, pathErr := os.Stat(i.Path()); os.IsNotExist(pathErr) {
		return errors.New("plugin does not exist")
	}

	return i.Update()
}

// FindSource determines the correct Installer for the given source.
func FindSource(location string, home draftpath.Home) (Installer, error) {
	installer, err := existingVCSRepo(location, home)
	if err != nil && err.Error() == "Cannot detect VCS" {
		return installer, errors.New("cannot get information about plugin source")
	}
	return installer, err
}

// New determines and returns the correct Installer for the given source
func New(source, version string, home draftpath.Home) (Installer, error) {
	if isLocalReference(source) {
		return NewLocalInstaller(source, home)
	}

	return NewVCSInstaller(source, version, home)
}

func debug(format string, args ...interface{}) {
	if Debug {
		format = fmt.Sprintf("[debug] %s\n", format)
		fmt.Printf(format, args...)
	}
}

// isLocalReference checks if the source exists on the filesystem.
func isLocalReference(source string) bool {
	_, err := os.Stat(source)
	return err == nil
}

// isPlugin checks if the directory contains a plugin.yaml file.
func isPlugin(dirname string) bool {
	_, err := os.Stat(filepath.Join(dirname, "plugin.yaml"))
	return err == nil
}
