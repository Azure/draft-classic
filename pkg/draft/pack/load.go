package pack

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"k8s.io/helm/pkg/chartutil"
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

	files, err := ioutil.ReadDir(topdir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			chart, err := chartutil.LoadDir(filepath.Join(topdir, file.Name()))
			if err != nil {
				return nil, err
			}
			pack.Charts = append(pack.Charts, chart)
		} else {
			var rootFile ChartRootFiles
			rootFile.Filename = file.Name()

			rootFileName := filepath.Join(topdir, file.Name())
			rootFile.File, err = ioutil.ReadFile(rootFileName)
			if err != nil {
				return nil, err
			}
			pack.Files = append(pack.Files, &rootFile)

			if err != nil {
				return nil, fmt.Errorf("error reading %s: %s", file.Name(), err)
			}
		}
	}

	return pack, nil
}
