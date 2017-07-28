package installer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Masterminds/vcs"

	"github.com/Azure/draft/pkg/draft/draftpath"
)

var _ Installer = new(VCSInstaller)

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
	if err := os.MkdirAll(home.Plugins(), 0755); err != nil {
		t.Fatalf("Could not create %s: %s", home.Plugins(), err)
	}

	source := "https://github.com/org/draft-env"
	testRepoPath, _ := filepath.Abs("../testdata/plugdir/echo")
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
	if i.Path() != home.Path("plugins", "draft-env") {
		t.Errorf("expected path '$DRAFT_HOME/plugins/draft-env', got %q", i.Path())
	}

	// Install again to test plugin exists error
	if err := Install(i); err == nil {
		t.Error("expected error for plugin exists, got none")
	} else if err.Error() != "plugin already exists" {
		t.Errorf("expected error for plugin exists, got (%v)", err)
	}
}
