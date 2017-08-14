package pack

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/helm/pkg/chartutil"
)

const packName = "foo"

func fooPackFiles() []*File {
	return []*File{
		{
			// Chart.yaml
			Path:    filepath.Join(packName, ChartDir, ChartfileName),
			Content: []byte("name: foo\n"),
			Perm:    0644,
		},
		{
			// values.yaml
			Path:    filepath.Join(packName, ChartDir, ValuesfileName),
			Content: nil,
			Perm:    0644,
		},
		{
			// .helmignore
			Path:    filepath.Join(packName, ChartDir, IgnorefileName),
			Content: nil,
			Perm:    0644,
		},
		{
			// deployment.yaml
			Path:    filepath.Join(packName, ChartDir, TemplatesDir, DeploymentName),
			Content: nil,
			Perm:    0644,
		},
		{
			// service.yaml
			Path:    filepath.Join(packName, ChartDir, TemplatesDir, ServiceName),
			Content: nil,
			Perm:    0644,
		},
		{
			// ingress.yaml
			Path:    filepath.Join(packName, ChartDir, TemplatesDir, IngressName),
			Content: nil,
			Perm:    0644,
		},
		{
			// NOTES.txt
			Path:    filepath.Join(packName, ChartDir, TemplatesDir, NotesName),
			Content: nil,
			Perm:    0644,
		},
		{
			// _helpers.tpl
			Path:    filepath.Join(packName, ChartDir, TemplatesDir, HelpersName),
			Content: nil,
			Perm:    0644,
		},
		{
			// Dockerfile
			Path:    filepath.Join(packName, DockerfileName),
			Content: nil,
			Perm:    0644,
		},
	}
}

func TestCreate(t *testing.T) {
	tdir, err := ioutil.TempDir("", "pack-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tdir)

	// first test: run Create() on a path that's actually a file
	filePath := filepath.Join(tdir, "foo.txt")
	if err = ioutil.WriteFile(filePath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Create("foo.txt", tdir, fooPackFiles()); err == nil {
		t.Error("expected error supplying path to a file")
	} else {
		expectedErr := fmt.Sprintf("file %s already exists and is not a directory", filePath)
		if err.Error() != expectedErr {
			t.Errorf("expected '%s',  got '%s'", expectedErr, err.Error())
		}
	}

	// second test: run Create() on a path that doesn't exist
	if _, err := Create("", filepath.Join(tdir, "bar"), fooPackFiles()); err == nil {
		t.Error("expected error supplying path to a non-existent dir")
	} else {
		expectedErr := fmt.Sprintf("stat %s: no such file or directory", filepath.Join(tdir, "bar"))
		if err.Error() != expectedErr {
			t.Errorf("expected '%s',  got '%s'", expectedErr, err.Error())
		}
	}

	// third test: run Create() on a base path that is a file
	if _, err := Create("", filepath.Join(tdir, "foo.txt"), fooPackFiles()); err == nil {
		t.Error("expected error supplying path to a file")
	} else {
		expectedErr := fmt.Sprintf("%s is not a directory", filepath.Join(tdir, "foo.txt"))
		if err.Error() != expectedErr {
			t.Errorf("expected '%s',  got '%s'", expectedErr, err.Error())
		}
	}

	// fourth test: run Create() on a valid path with bad write permissions
	badPermsDir, err := ioutil.TempDir(tdir, "badpack-")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(badPermsDir, 0000); err != nil {
		t.Fatal(err)
	}
	if _, err := Create(packName, badPermsDir, fooPackFiles()); err == nil {
		t.Error("expected error when creating pack in dir with bad write permissions")
	} else {
		expectedErr := fmt.Sprintf("mkdir %s: permission denied", filepath.Join(badPermsDir, packName))
		if err.Error() != expectedErr {
			t.Errorf("expected '%s',  got '%s'", expectedErr, err.Error())
		}
	}

	// now actually create a valid pack and perform further tests
	c, err := Create(packName, tdir, fooPackFiles())
	if err != nil {
		t.Error(err)
	}

	// re-run Create(), expecting to fail with ErrPackExists
	if _, err := Create(packName, tdir, fooPackFiles()); err != ErrPackExists {
		t.Errorf("expected ErrPackExists, got '%v'", err)
	}

	dir := filepath.Join(tdir, packName)

	// even though we tack on some things like the detect script and the Dockerfile, the chart
	// directory should still be load-able using Helm's libs.
	mychart, err := chartutil.LoadDir(filepath.Join(c, ChartDir))
	if err != nil {
		t.Fatalf("Failed to load newly created chart %q: %s", c, err)
	}

	if mychart.Metadata.Name != packName {
		t.Errorf("Expected name to be 'foo', got %q", mychart.Metadata.Name)
	}

	for _, d := range []string{TemplatesDir} {
		if fi, err := os.Stat(filepath.Join(dir, ChartDir, d)); err != nil {
			t.Errorf("Expected %s dir: %s", d, err)
		} else if !fi.IsDir() {
			t.Errorf("Expected %s to be a directory.", d)
		}
	}

	for _, f := range []string{ChartfileName, ValuesfileName, IgnorefileName} {
		if fi, err := os.Stat(filepath.Join(dir, ChartDir, f)); err != nil {
			t.Errorf("Expected %s file: %s", f, err)
		} else if fi.IsDir() {
			t.Errorf("Expected %s to be a file.", f)
		}
	}

	for _, f := range []string{NotesName, DeploymentName, ServiceName, IngressName, HelpersName} {
		if fi, err := os.Stat(filepath.Join(dir, ChartDir, TemplatesDir, f)); err != nil {
			t.Errorf("Expected %s file: %s", f, err)
		} else if fi.IsDir() {
			t.Errorf("Expected %s to be a file.", f)
		}
	}

	if fi, err := os.Stat(filepath.Join(dir, DockerfileName)); err != nil {
		t.Errorf("Expected %s file: %s", DockerfileName, err)
	} else if fi.IsDir() {
		t.Errorf("Expected %s to be a file.", DockerfileName)
	}

}

func TestCreateFrom(t *testing.T) {
	tdir, err := ioutil.TempDir("", "pack-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tdir)

	if err := CreateFrom(tdir, "testdata/pack-python"); err != nil {
		t.Errorf("expected err to be nil, got %v", err)
	}

	if err := CreateFrom(tdir, "testdata/pack-does-not-exist"); err == nil {
		t.Error("expected err to be non-nil with an invalid source pack")
	}
}

func TestBuiltins(t *testing.T) {
	b, err := Builtins()
	if err != nil {
		t.Fatal(err)
	}

	if c := len(b); c <= 0 {
		t.Errorf("Expected at least one pack, got %d", c)
	}

	gopack, ok := b["go"]
	if !ok {
		t.Fatal("Go pack not found")
	}

	if c := len(gopack); c != 11 {
		t.Errorf("Expected 11 files in pack, got %d", c)
		for _, f := range gopack {
			t.Log(f.Path)
		}
	}
}
