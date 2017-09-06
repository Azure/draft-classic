package installer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Masterminds/vcs"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/plugin/installer"
)

var _ installer.Installer = new(VCSInstaller)

type testRepo struct {
	local, remote, current string
	tags, branches         []string
	err                    error
	vcs.Repo
}

func (r *testRepo) LocalPath() string           { return r.local }
func (r *testRepo) Remote() string              { return r.remote }
func (r *testRepo) Update() error               { return r.err }
func (r *testRepo) Get() error                  { return r.err }
func (r *testRepo) IsReference(string) bool     { return false }
func (r *testRepo) Tags() ([]string, error)     { return r.tags, r.err }
func (r *testRepo) Branches() ([]string, error) { return r.branches, r.err }
func (r *testRepo) UpdateVersion(version string) error {
	r.current = version
	return r.err
}

func TestVCSInstallerSuccess(t *testing.T) {
	dh, err := ioutil.TempDir("", "draft-home-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dh)

	home := draftpath.Home(dh)
	if err := os.MkdirAll(home.Packs(), 0755); err != nil {
		t.Fatalf("Could not create %s: %s", home.Packs(), err)
	}

	source := "https://github.com/org/defaultpacks"
	testRepoPath, _ := filepath.Abs("../testdata/packdir/defaultpacks")
	repo := &testRepo{
		local: testRepoPath,
		tags:  []string{"0.1.0", "0.1.1"},
	}

	i, err := New(source, "~0.1.0", home)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// ensure a VCSInstaller was returned
	vcsInstaller, ok := i.(*VCSInstaller)
	if !ok {
		t.Error("expected a VCSInstaller")
	}

	// set the testRepo in the VCSInstaller
	vcsInstaller.Repo = repo

	if err := Install(i); err != nil {
		t.Error(err)
	}
	if repo.current != "0.1.1" {
		t.Errorf("expected version '0.1.1', got %q", repo.current)
	}
	if i.Path() != home.Path("packs", "defaultpacks") {
		t.Errorf("expected path '$DRAFT_HOME/packs/defaultpacks', got %q", i.Path())
	}

	// Install again to test pack repo exists error
	if err := Install(i); err == nil {
		t.Error("expected error for pack repo exists, got none")
	} else if err.Error() != "pack repo already exists" {
		t.Errorf("expected error for pack repo exists, got (%v)", err)
	}
}

func TestVCSInstallerUpdate(t *testing.T) {

	dh, err := ioutil.TempDir("", "draft-home-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dh)

	home := draftpath.Home(dh)
	if err := os.MkdirAll(home.Packs(), 0755); err != nil {
		t.Fatalf("Could not create %s: %s", home.Packs(), err)
	}

	// Draft can install itself. Pretty neat eh?
	source := "https://github.com/Azure/draft"
	i, err := New(source, "0.6.0", home)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// ensure a VCSInstaller was returned
	_, ok := i.(*VCSInstaller)
	if !ok {
		t.Error("expected a VCSInstaller")
	}

	if err := Update(i); err == nil {
		t.Error("expected error for pack repo does not exist, got none")
	} else if err.Error() != "pack repo does not exist" {
		t.Errorf("expected error for pack repo does not exist, got (%v)", err)
	}

	// Install pack repo before update
	if err := Install(i); err != nil {
		t.Error(err)
	}

	// Test FindSource method for positive result
	packInfo, err := FindSource(i.Path(), home)
	if err != nil {
		t.Error(err)
	}

	repoRemote := packInfo.(*VCSInstaller).Repo.Remote()
	if repoRemote != source {
		t.Errorf("invalid source found, expected %q got %q", source, repoRemote)
	}

	// Update pack repo
	if err := Update(i); err != nil {
		t.Error(err)
	}

	// Test update failure
	os.Remove(filepath.Join(i.Path(), "README.md"))
	// Testing update for error
	if err := Update(i); err == nil {
		t.Error("expected error for pack repo modified, got none")
	} else if err.Error() != "pack repo was modified" {
		t.Errorf("expected error for pack repo modified, got (%v)", err)
	}

}
