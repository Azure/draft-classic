package pack

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// Create creates a new Pack in a directory.
//
// Inside of dir, this will create a directory based on the name. It will
// then write the Chart.yaml into this directory and create the (empty)
// appropriate directories.
//
// The returned string will point to the newly created directory. It will be
// an absolute path, even if the provided base directory was relative.
//
// If dir does not exist, this will return an error.
// If Chart.yaml or any directories cannot be created, this will return an
// error. In such a case, this will attempt to clean up by removing the
// new pack directory.
func Create(name, dir string, files []*File) (string, error) {
	path, err := filepath.Abs(dir)
	if err != nil {
		return path, err
	}

	if fi, err := os.Stat(path); err != nil {
		return path, err
	} else if !fi.IsDir() {
		return path, fmt.Errorf("%s is not a directory", path)
	}

	pdir := filepath.Join(path, name)
	cdir := filepath.Join(pdir, ChartDir)
	if fi, err := os.Stat(pdir); err == nil && !fi.IsDir() {
		return pdir, fmt.Errorf("file %s already exists and is not a directory", pdir)
	}
	if err := os.MkdirAll(pdir, 0755); err != nil {
		return pdir, err
	}

	if err := os.MkdirAll(cdir, 0755); err != nil {
		return pdir, err
	}

	for _, d := range []string{TemplatesDir, ChartsDir} {
		if err := os.MkdirAll(filepath.Join(cdir, d), 0755); err != nil {
			return pdir, err
		}
	}

	cf := filepath.Join(cdir, ChartfileName)
	if _, err := os.Stat(cf); err != nil {
		if err := chartutil.SaveChartfile(cf, &chart.Metadata{Name: name}); err != nil {
			return pdir, err
		}
	}

	for _, file := range files {
		if _, err := os.Stat(filepath.Join(pdir, file.Path)); err == nil {
			// File exists and is okay. Skip it.
			continue
		}
		if err := ioutil.WriteFile(filepath.Join(pdir, file.Path), file.Content, file.Perm); err != nil {
			return pdir, err
		}
	}
	return pdir, nil
}

// CreateFrom scaffolds a directory with the src pack.
func CreateFrom(chartMeta *chart.Metadata, dest, src string) error {
	pack, err := FromDir(src)
	if err != nil {
		return fmt.Errorf("could not load %s: %s", src, err)
	}

	pack.Chart.Metadata = chartMeta
	return pack.SaveDir(dest, false)
}
