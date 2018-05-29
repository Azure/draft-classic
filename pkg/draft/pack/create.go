package pack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/draft/pkg/draft/pack/repo"
)

// CreateFrom scaffolds a directory with the src pack
func CreateFrom(dest, src string, appName string) error {
	// first do some validation that we are copying from a valid pack directory
	pack, err := FromDir(src)
	if err != nil {
		return fmt.Errorf("could not load pack: %s\nTry running:\n\t$ draft pack-repo update", err)
	}

	// Update ChartName if appName is non-empty
	if appName != "" {
		pack.Chart.Metadata.Name = appName
	} else {
		cwd, err := os.Getwd()
		if err == nil {
			pack.Chart.Metadata.Name = filepath.Base(cwd)
		}
	}
	return pack.SaveDir(dest)
}

// Find loops through each pack repo in packsDir to find pack with given name
func Find(packsDir, name string) ([]string, error) {
	packs := []string{}
	for _, r := range repo.FindRepositories(packsDir) {
		pack, err := r.Pack(name)
		if err != nil && err != repo.ErrPackNotFoundInRepo {
			return packs, err
		}
		if pack != "" {
			packs = append(packs, pack)
		}
	}

	return packs, nil
}
