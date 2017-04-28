package manifest

const (
	// DefaultEnvironmentName is the name invoked from draft.toml on `draft up` when
	// --environment is not supplied.
	DefaultEnvironmentName = "development"
	// DefaultNamespace specifies the namespace apps should be deployed to by default.
	DefaultNamespace = "default"
	// DefaultWatchDelaySeconds is the time delay between files being changed and when a
	// new draft up` invocation is called when --watch is supplied.
	DefaultWatchDelaySeconds = 2
)

// Manifest represents a draft.yaml
type Manifest struct {
	Environments map[string]*Environment `toml:"environments"`
}

// Environment represents the environment for a given app at build time
type Environment struct {
	AppName      string   `toml:"name,omitempty"`
	BuildTarPath string   `toml:"build_tar,omitempty"`
	ChartTarPath string   `toml:"chart_tar,omitempty"`
	Namespace    string   `toml:"namespace,omitempty"`
	Values       []string `toml:"set,omitempty"`
	Wait         bool     `toml:"wait,omitempty"`
	Watch        bool     `toml:"watch,omitempty"`
	WatchDelay   int      `toml:"watch_delay,omitempty"`
}

// New creates a new manifest with the Environments intialized.
func New() *Manifest {
	m := Manifest{
		Environments: make(map[string]*Environment),
	}
	m.Environments[DefaultEnvironmentName] = &Environment{
		Namespace:  DefaultNamespace,
		WatchDelay: DefaultWatchDelaySeconds,
	}
	return &m
}
