package pack

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"k8s.io/helm/pkg/chartutil"

	"github.com/deis/draft/pkg/osutil"
)

// FromDir takes a string name, tries to resolve it to a file or directory, and then loads it.
//
// This is the preferred way to load a pack. It will discover the pack encoding
// and hand off to the appropriate pack reader.
func FromDir(dir string) (*Pack, error) {
	pack := new(Pack)

	topdir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	pack.Chart, err = chartutil.LoadDir(filepath.Join(topdir, ChartDir))
	if err != nil {
		return nil, err
	}

	dockerfile := filepath.Join(topdir, DockerfileName)
	pack.Dockerfile, err = ioutil.ReadFile(dockerfile)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %s", dockerfile, err)
	}

	detect := filepath.Join(topdir, DetectName)
	detectExists, err := osutil.Exists(detect)
	if err != nil {
		return nil, fmt.Errorf("error checking if the detect script exists: %v", err)
	}
	if detectExists {
		pack.DetectScript, err = ioutil.ReadFile(detect)
		if err != nil {
			return nil, fmt.Errorf("error reading %s: %s", detect, err)
		}
	}

	return pack, nil
}
