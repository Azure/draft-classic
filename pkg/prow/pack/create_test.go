package pack

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

const packName = "foo"

func TestCreate(t *testing.T) {
	tdir, err := ioutil.TempDir("", "pack-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tdir)

	c, err := Create(packName, tdir)
	if err != nil {
		t.Fatal(err)
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

	for _, d := range []string{TemplatesDir, ChartsDir} {
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

	for _, f := range []string{DockerfileName, DetectName} {
		if fi, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("Expected %s file: %s", f, err)
		} else if fi.IsDir() {
			t.Errorf("Expected %s to be a file.", f)
		}
	}

}

func TestCreateFrom(t *testing.T) {
	tdir, err := ioutil.TempDir("", "pack-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tdir)

	if err := CreateFrom(&chart.Metadata{Name: "foo"}, tdir, "testdata/pack-python"); err != nil {
		t.Errorf("expected err to be nil, got %v", err)
	}

	if err := CreateFrom(&chart.Metadata{Name: "foo"}, tdir, "testdata/pack-does-not-exist"); err == nil {
		t.Error("expected err to be non-nil with an invalid source pack")
	}
}
