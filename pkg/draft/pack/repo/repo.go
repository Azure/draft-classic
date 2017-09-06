package repo

import (
	"os"
	"path/filepath"
	"strings"
)

// Repository represents a pack repository.
type Repository struct {
	Name string
	Dir  string
}

// FindRepositories takes a given path and returns a list of repositories.
//
// Repositories are defined as directories with a "packs" directory present.
func FindRepositories(path string) []Repository {
	var repos []Repository
	filepath.Walk(path, func(walkPath string, f os.FileInfo, err error) error {
		// find all directories in walkPath that have a child directory called "packs"
		if f.IsDir() && f.Name() == "packs" && walkPath != path {
			repos = append(repos, Repository{
				Name: filepath.Base(filepath.Dir(walkPath)),
				Dir:  strings.TrimSuffix(walkPath, "/"+f.Name()),
			})
		}
		return nil
	})
	return repos
}
