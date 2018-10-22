package pack

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/helm/pkg/chartutil"
)

// FromDir takes a string name, tries to resolve it to a file or directory, and then loads it.
//
// This is the preferred way to load a pack. It will discover the pack encoding
// and hand off to the appropriate pack reader.
func FromDir(dir string) (*Pack, error) {
	pack := new(Pack)
	pack.Files = make(map[string]PackFile)

	topdir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	pack.Chart, err = chartutil.LoadDir(filepath.Join(topdir, ChartsDir))
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(topdir)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %s", topdir, err)
	}

	// load all files in pack directory
	for _, fInfo := range files {
		if !fInfo.IsDir() {
			f, err := os.Open(filepath.Join(topdir, fInfo.Name()))
			if err != nil {
				return nil, err
			}
			if fInfo.Name() != "README.md" {
				pack.Files[fInfo.Name()] = PackFile{f, fInfo.Mode().Perm()}
			}
		} else {
			if fInfo.Name() != "charts" {
				packFiles, err := extractFiles(filepath.Join(topdir, fInfo.Name()), "")
				if err != nil {
					return nil, err
				}
				for k, packFile := range packFiles {
					pack.Files[k] = packFile
				}
			}
		}
	}

	return pack, nil
}

func extractFiles(dir, base string) (map[string]PackFile, error) {
	baseDir := filepath.Join(base, filepath.Base(dir))
	packFiles := make(map[string]PackFile)

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(absDir)
	if err != nil {
		return packFiles, fmt.Errorf("error reading %s: %s", dir, err)
	}

	for _, fInfo := range files {
		if !fInfo.IsDir() {
			fPath := filepath.Join(dir, fInfo.Name())
			f, err := os.Open(fPath)
			if err != nil {
				return nil, err
			}
			packFiles[filepath.Join(baseDir, fInfo.Name())] = PackFile{f, fInfo.Mode().Perm()}
		} else {
			nestedPackFiles, err := extractFiles(filepath.Join(dir, fInfo.Name()), baseDir)
			if err != nil {
				return nil, err
			}
			for k, v := range nestedPackFiles {
				packFiles[k] = v
			}
		}
	}
	return packFiles, nil
}
