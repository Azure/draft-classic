package installer

import (
	"net/url"
	"os"
	"path/filepath"
	"sort"

	"github.com/Azure/draft/pkg/plugin"
	"github.com/Azure/draft/pkg/plugin/repository"
	"github.com/Masterminds/semver"
	"github.com/Masterminds/vcs"
)

//VCSInstaller installs rigs from a remote repository
type VCSInstaller struct {
	Repo    vcs.Repo
	Version string
	Source  string
	Home    plugin.Home
}

// NewVCSInstaller creates a new VCSInstaller.
func NewVCSInstaller(source, version string, home plugin.Home) (*VCSInstaller, error) {
	i := &VCSInstaller{
		Version: version,
		Source:  source,
		Home:    home,
	}
	repo, err := vcs.NewRepo(source, i.Path())
	if err != nil {
		return nil, err
	}

	i.Repo = repo
	return i, err
}

// Install clones a remote repository to the rig directory
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

	if ref != "" {
		if err := i.setVersion(i.Repo, ref); err != nil {
			return err
		}
	}

	if !isRig(i.Repo.LocalPath()) {
		return repository.ErrMissingMetadata
	}

	return nil
}

// Update updates a remote repository
func (i *VCSInstaller) Update() error {
	if i.Repo.IsDirty() {
		return repository.ErrRepoDirty
	}
	if err := i.Repo.Update(); err != nil {
		return err
	}
	if !isRig(i.Repo.LocalPath()) {
		return repository.ErrMissingMetadata
	}
	return nil
}

func existingVCSRepo(location string, home plugin.Home) (Installer, error) {
	repo, err := vcs.NewRepo("", location)
	if err != nil {
		return nil, err
	}
	i := &VCSInstaller{
		Repo:   repo,
		Source: repo.Remote(),
		Home:   home,
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
			// If the constraint passes get the original reference
			ver := v.Original()
			return ver, nil
		}
	}

	return "", repository.ErrVersionDoesNotExist
}

// sync will clone or update a remote repo.
func (i *VCSInstaller) sync(repo vcs.Repo) error {

	if _, err := os.Stat(repo.LocalPath()); os.IsNotExist(err) {
		return repo.Get()
	}
	return repo.Update()
}

// Path is where the rig will be installed into.
func (i *VCSInstaller) Path() string {
	if i.Source == "" {
		return ""
	}
	u, _ := url.Parse(i.Source)
	return filepath.Join(i.Home.Repositories(), u.Host, u.Path)
}
