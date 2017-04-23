package manifest

// Manifest represents a draft.yaml
type Manifest struct {
	Environments map[string]*Environment `json:"environments"`
}

// Environment represents the environment for a given app at build time
type Environment struct {
	AppName           string   `json:"name,omitempty"`
	BuildTarPath      string   `json:"build_tar,omitempty"`
	ChartTarPath      string   `json:"chart_tar,omitempty"`
	Namespace         string   `json:"namespace,omitempty"`
	Values            []string `json:"set,omitempty"`
	RawValueFilePaths []string `json:"values,omitempty"`
	Wait              bool     `json:"wait,omitempty"`
	Watch             bool     `json:"watch,omitempty"`
	WatchDelay        int      `json:"watch_delay,omitempty"`
}

// New creates a new manifest with the Environments intialized.
func New() *Manifest {
	return &Manifest{
		Environments: make(map[string]*Environment),
	}
}
