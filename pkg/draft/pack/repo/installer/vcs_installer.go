package installer

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/Masterminds/vcs"
	"k8s.io/helm/pkg/plugin/cache"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/plugin/installer"
)

//VCSInstaller installs packs from a remote repository
type VCSInstaller struct {
	Repo    vcs.Repo
	Version string
	base
}

// NewVCSInstaller creates a new VCSInstaller.
func NewVCSInstaller(source, version string, home draftpath.Home) (*VCSInstaller, error) {
	// create a system safe cache key
	key, err := cache.Key(source)
	if err != nil {
		return nil, err
	}
	cachedpath := home.Path("cache", "packs", key)
	repo, err := vcs.NewRepo(source, cachedpath)
	if err != nil {
		return nil, err
	}

	i := &VCSInstaller{
		Repo:    repo,
		Version: version,
		base:    newBase(source, home),
	}
	return i, err
}

// Install clones a remote repository and creates a symlink to the pack repo directory in DRAFT_HOME
//
// Implements Installer
func (i *VCSInstaller) Install() error {
	if err := i.sync(i.Repo); err != nil {
		return err
	}

	ref, err := i.solveVersion(i.Repo)
	if err != nil {
		return err
	}

	if err := i.setVersion(i.Repo, ref); err != nil {
		return err
	}

	if !isPackRepo(i.Repo.LocalPath()) {
		return ErrMissingPackDir
	}

	return i.link(i.Repo.LocalPath())
}

// Update updates a remote repository
func (i *VCSInstaller) Update() error {
	if i.Repo.IsDirty() {
		return errors.New("pack repo was modified")
	}
	if err := i.Repo.Update(); err != nil {
		return err
	}
	if !isPackRepo(i.Repo.LocalPath()) {
		return ErrMissingPackDir
	}
	return nil
}

func existingVCSRepo(location string, home draftpath.Home) (installer.Installer, error) {
	repo, err := vcs.NewRepo("", location)
	if err != nil {
		return nil, err
	}
	i := &VCSInstaller{
		Repo: repo,
		base: newBase(repo.Remote(), home),
	}

	return i, err
}

// Filter a list of versions to only included semantic versions. The response
// is a mapping of the original version to the semantic version.
func getSemVers(refs []string) []*semver.Version {
	var sv []*semver.Version
	for _, r := range refs {
		if v, err := semver.NewVersion(r); err == nil {
			sv = append(sv, v)
		}
	}
	return sv
}

// setVersion attempts to checkout the version
func (i *VCSInstaller) setVersion(repo vcs.Repo, ref string) error {
	return repo.UpdateVersion(ref)
}

func (i *VCSInstaller) solveVersion(repo vcs.Repo) (string, error) {
	if i.Version == "" {
		return "", nil
	}

	if repo.IsReference(i.Version) {
		return i.Version, nil
	}

	// Create the constraint first to make sure it's valid before
	// working on the repo.
	constraint, err := semver.NewConstraint(i.Version)
	if err != nil {
		return "", err
	}

	// Get the tags
	refs, err := repo.Tags()
	if err != nil {
		return "", err
	}

	// Convert and filter the list to semver.Version instances
	semvers := getSemVers(refs)

	// Sort semver list
	sort.Sort(sort.Reverse(semver.Collection(semvers)))
	for _, v := range semvers {
		if constraint.Check(v) {
			// If the constrint passes get the original reference
			ver := v.Original()
			return ver, nil
		}
	}

	return "", fmt.Errorf("requested version %q does not exist for pack repo %q", i.Version, i.Repo.Remote())
}

// sync will clone or update a remote repo.
func (i *VCSInstaller) sync(repo vcs.Repo) error {

	if _, err := os.Stat(repo.LocalPath()); os.IsNotExist(err) {
		return repo.Get()
	}
	return repo.Update()
}
