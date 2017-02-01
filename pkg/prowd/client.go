package prowd

import (
	"github.com/deis/prow/pkg/version"
)

// Client is the interface for creating clients to communicate with prowd.
type Client interface {
	// Up uploads the contents of appDir to prowd, installs it in the specified namespace and
	// returns a Helm Release.
	Up(appDir, namespace string) error
	// Version returns the server version.
	Version() (*version.Version, error)
}
