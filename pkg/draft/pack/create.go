package pack

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/draft/pkg/draft/pack/generated"
	"github.com/Azure/draft/pkg/osutil"
)

// ErrPackExists is returned when a user calls Create() but the pack already exists.
var ErrPackExists = errors.New("pack already exists")

// File represents a file within a Pack
type File struct {
	Path    string
	Content []byte
	Perm    os.FileMode
}

// Builtins fetches all built-in packs as a map of packname=>files.
func Builtins() (map[string][]*File, error) {
	res := map[string][]*File{}
	fnames := generated.AssetNames()
	for _, fname := range fnames {
		parts := strings.SplitN(fname, "/", 2)
		if len(parts) != 2 {
			// Skip files and empty directories.
			continue
		}
		pname := parts[0]
		if files, ok := res[pname]; ok {
			f, err := file(fname)
			if err != nil {
				return res, err
			}
			files = append(files, f)
			res[pname] = files
			continue
		}
		f, err := file(fname)
		if err != nil {
			return res, err
		}
		res[pname] = []*File{f}
	}
	return res, nil
}

// file loads an Asset as a *File
func file(name string) (*File, error) {
	fi, err := generated.AssetInfo(name)
	if err != nil {
		return nil, err
	}

	asset, err := generated.Asset(name)
	if err != nil {
		return nil, err
	}
	return &File{Path: name, Content: asset, Perm: fi.Mode()}, nil
}

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

	ret := filepath.Join(path, name)
	if fi, err := os.Stat(ret); err == nil && !fi.IsDir() {
		return ret, fmt.Errorf("file %s already exists and is not a directory", ret)
	}

	if dirExists, err := osutil.Exists(ret); err == nil && dirExists {
		return ret, ErrPackExists
	}

	// Next, we can simply loop through files and create each.
	// We call MkdirAll for a safe way to create any missing dirs.
	for _, f := range files {
		fullpath := filepath.Join(path, f.Path)
		dir := filepath.Dir(fullpath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return ret, err
		}
		// Don't everwrite existing files.
		if _, err := os.Stat(fullpath); err == nil {
			continue
		}
		if err := ioutil.WriteFile(fullpath, f.Content, f.Perm); err != nil {
			return ret, err
		}
	}

	return ret, nil
}

// CreateFrom scaffolds a directory with the src pack.
func CreateFrom(dest, src string) error {
	pack, err := FromDir(src)
	if err != nil {
		return fmt.Errorf("could not load %s: %s", src, err)
	}
	return pack.SaveDir(dest, false)
}
