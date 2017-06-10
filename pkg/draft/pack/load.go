package pack

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/facebookgo/symwalk"
	"github.com/ghodss/yaml"
	"k8s.io/helm/pkg/chartutil"
)

// BufferedFile represents an archive file buffered for later processing.
type BufferedFile struct {
	Name string
	Data []byte
}

// FromDir takes a string name, tries to resolve it to a file or directory, and then loads it.
//
// This is the preferred way to load a pack. It will discover the pack encoding
// and hand off to the appropriate pack reader.
func FromDir(dir string) (*Pack, error) {

	topdir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	prefix := topdir + string(filepath.Separator)

	files := []*BufferedFile{}

	err = symwalk.Walk(topdir, func(name string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		n := strings.TrimPrefix(name, prefix)

		// normalize to / to ensure it works on Windows
		n = filepath.ToSlash(n)

		data, err := ioutil.ReadFile(name)
		if err != nil {
			return fmt.Errorf("error reading %s: %s", n, err)
		}

		files = append(files, &BufferedFile{Name: n, Data: data})
		return nil
	})
	if err != nil {
		return nil, err
	}

	pck, err := LoadFiles(files)
	if err != nil {
		return pck, err
	}

	// Assume that the name is the name of the directory if no
	// metadata is not present.
	if pck.Metadata == nil {
		_, name := filepath.Split(topdir)
		pck.Metadata = &Metadata{
			Name:    name,
			Version: "0.1.0",
		}
	}

	return pck, err
}

func LoadFiles(files []*BufferedFile) (*Pack, error) {

	pck := new(Pack)

	// grab all the files from the chart
	chartFiles := []*chartutil.BufferedFile{}

	chartDir := ChartDir + "/"
	for _, f := range files {

		if strings.HasPrefix(f.Name, chartDir) {
			chartName := strings.TrimPrefix(f.Name, chartDir)
			chartFiles = append(chartFiles, &chartutil.BufferedFile{Name: chartName, Data: f.Data})
			continue
		}

		if f.Name == MetadataName {
			err := yaml.Unmarshal(f.Data, &pck.Metadata)
			if err != nil {
				return nil, err
			}
		}

		if f.Name == DockerfileName {
			pck.Dockerfile = f.Data
		}

		if f.Name == DetectName {
			pck.DetectScript = f.Data
		}
	}

	// TODO (rodcoutier) once metadata is required, check it here

	if pck.Dockerfile == nil {
		return nil, fmt.Errorf("failed to find `%s` file", DockerfileName)
	}

	chart, err := chartutil.LoadFiles(chartFiles)
	if err != nil {
		return nil, err
	}
	pck.Chart = chart

	return pck, nil
}

func Load(path string) (*Pack, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return FromDir(path)
	}
	return LoadArchive(path)
}

func LoadArchive(name string) (*Pack, error) {

	fi, err := os.Stat(name)
	if err != nil {
		return nil, err
	} else if fi.IsDir() {
		return nil, errors.New("cannot load a directory")
	}

	raw, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer raw.Close()

	unzipped, err := gzip.NewReader(raw)
	if err != nil {
		return nil, nil
	}
	defer unzipped.Close()

	files := []*BufferedFile{}
	tr := tar.NewReader(unzipped)
	for {
		b := bytes.NewBuffer(nil)
		hd, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if hd.FileInfo().IsDir() {
			// Use this instead of hd.Typeflag because we don't have to do any
			// inference chasing.
			continue
		}

		parts := strings.Split(hd.Name, "/")
		n := path.Join(parts[1:]...)

		// TODO (rodcloutier) Check if Pack.yaml is in the base directory as an error?

		if _, err := io.Copy(b, tr); err != nil {
			return nil, err
		}

		files = append(files, &BufferedFile{Name: n, Data: b.Bytes()})
		b.Reset()
	}

	if len(files) == 0 {
		return nil, errors.New("no files in chart archive")
	}

	return LoadFiles(files)
}
