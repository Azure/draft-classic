package manifest

import (
	"os"
	"path/filepath"

	"github.com/technosophos/moniker"
)

const (
	// DefaultEnvironmentName is the name invoked from draft.toml on `draft up` when
	// --environment is not supplied.
	DefaultEnvironmentName = "development"
	// DefaultNamespace specifies the namespace apps should be deployed to by default.
	DefaultNamespace = "default"
	// DefaultWatchDelaySeconds is the time delay between files being changed and when a
	// new draft up` invocation is called when --watch is supplied.
	DefaultWatchDelaySeconds = 2
	// DefaultDockerfile is the Dockerfile being used by default
	DefaultDockerfile = "Dockerfile"
)

// Manifest represents a draft.toml
type Manifest struct {
	Environments map[string]*Environment `toml:"environments"`
}

// Environment represents the environment for a given app at build time
type Environment struct {
	Name              string            `toml:"name,omitempty"`
	ContainerBuilder  string            `toml:"container-builder,omitempty"`
	Registry          string            `toml:"registry,omitempty"`
	ResourceGroupName string            `toml:"resource-group-name,omitempty"`
	BuildTarPath      string            `toml:"build-tar,omitempty"`
	ChartTarPath      string            `toml:"chart-tar,omitempty"`
	Namespace         string            `toml:"namespace,omitempty"`
	Values            []string          `toml:"set,omitempty"`
	Wait              bool              `toml:"wait"`
	Watch             bool              `toml:"watch"`
	WatchDelay        int               `toml:"watch-delay,omitempty"`
	OverridePorts     []string          `toml:"override-ports,omitempty"`
	AutoConnect       bool              `toml:"auto-connect"`
	CustomTags        []string          `toml:"custom-tags,omitempty"`
	Dockerfile        string            `toml:"dockerfile"`
	Chart             string            `toml:"chart"`
	ImageBuildArgs    map[string]string `toml:"image-build-args,omitempty"`
}

// New creates a new manifest with the Environments intialized.
func New() *Manifest {
	m := Manifest{
		Environments: make(map[string]*Environment),
	}
	m.Environments[DefaultEnvironmentName] = &Environment{
		Name:        generateName(),
		Namespace:   DefaultNamespace,
		Wait:        true,
		Watch:       false,
		WatchDelay:  DefaultWatchDelaySeconds,
		AutoConnect: false,
		Dockerfile:  DefaultDockerfile,
	}
	return &m
}

// generateName generates a name based on the current working directory or a random name.
func generateName() string {
	var name string
	cwd, err := os.Getwd()
	if err == nil {
		name = filepath.Base(cwd)
	} else {
		namer := moniker.New()
		name = namer.NameSep("-")
	}
	return name
}
