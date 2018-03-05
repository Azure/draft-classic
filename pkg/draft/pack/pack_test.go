package pack

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/helm/pkg/proto/hapi/chart"
)

const testDockerfile = `FROM nginx:latest
`

func TestSaveDir(t *testing.T) {

	charts := []*chart.Chart{
		{
			Metadata: &chart.Metadata{
				Name: "chart-for-nigel-thornberry",
			},
		},
	}

	file := ChartRootFiles{
		Filename: "Dockerfile",
		File:     []byte(testDockerfile),
	}

	files := []*ChartRootFiles{&file}

	p := &Pack{
		Charts: charts,
		Files:  files,
	}
	dir, err := ioutil.TempDir("", "draft-pack-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := p.SaveDir(dir); err != nil {
		t.Errorf("expected there to be no error when writing to %v, got %v", dir, err)
	}
}

func TestSaveDirDockerfileExistsInAppDir(t *testing.T) {
	charts := []*chart.Chart{
		{
			Metadata: &chart.Metadata{
				Name: "chart-for-nigel-thornberry",
			},
		},
	}

	file := &ChartRootFiles{
		Filename: "Dockerfile",
		File:     []byte(testDockerfile),
	}

	files := []*ChartRootFiles{file}

	p := &Pack{
		Charts: charts,
		Files:  files,
	}
	dir, err := ioutil.TempDir("", "draft-pack-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	tmpfn := filepath.Join(dir, "Dockerfile")
	expectedDockerfile := []byte("FROM draft")
	if err := ioutil.WriteFile(tmpfn, expectedDockerfile, 0644); err != nil {
		t.Fatal(err)
	}

	if err := p.SaveDir(dir); err != nil {
		t.Errorf("expected there to be no error when writing to %v, got %v", dir, err)
	}

	savedDockerfile, err := ioutil.ReadFile(tmpfn)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(savedDockerfile, expectedDockerfile) {
		t.Errorf("expected '%s', got '%s'", string(expectedDockerfile), string(savedDockerfile))
	}
}
