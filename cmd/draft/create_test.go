package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Azure/draft/pkg/draft/draftpath"
)

const gitkeepfile = ".gitkeep"

var update = flag.Bool("update", false, "update generated reference files")

func TestCreate(t *testing.T) {
	var generatedpath = "testdata/create/generated"

	testCases := []struct {
		src string

		wantErr bool
	}{
		{"testdata/create/src/simple-go", false},
		{"testdata/create/src/simple-go-with-draftignore", false},
		{"testdata/create/src/simple-go-with-chart", false},
		{"testdata/create/src/empty", false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("create %s", tc.src), func(t *testing.T) {
			pDir, teardown := tempDir(t)
			defer teardown()

			destcompare := filepath.Join(generatedpath, path.Base(tc.src))
			// On update, override and clean destcompare
			if *update {
				// Skip the ones where we expect an error on
				if tc.wantErr {
					return
				}
				pDir = destcompare
			}
			copyTree(t, tc.src, pDir)

			// Test
			create := &createCmd{
				appName: "myapp",
				out:     os.Stdout,
				home:    draftpath.Home("../../"),
				dest:    pDir,
			}
			err := create.run()

			// Error checking
			if err != nil != tc.wantErr {
				t.Errorf("draft create error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// append .gitkeep file on empty directories
			if *update && !tc.wantErr {
				addGitKeep(t, pDir)
			}

			// Compare directories to ensure they are identical
			assertIdentical(t, pDir, destcompare)
		})
	}
}

// tempDir create and clean a temporary directory to work in our tests
func tempDir(t *testing.T) (string, func()) {
	path, err := ioutil.TempDir("", "draft-create")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	return path, func() {
		if err := os.RemoveAll(path); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
}

// copyTree copies src directory content tree to dest.
// If dest exists, it's deleted.
// We don't handle symlinks (not needed in this test helper)
func copyTree(t *testing.T, src, dest string) {
	if err := os.RemoveAll(dest); err != nil {
		t.Fatalf("couldn't remove directory %s: %v", src, err)
	}

	if err := filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		dest := filepath.Join(dest, strings.TrimPrefix(p, src))
		if info.IsDir() {
			if err := os.MkdirAll(dest, info.Mode()); err != nil {
				return err
			}
		} else {
			data, err := ioutil.ReadFile(p)
			if err != nil {
				return err
			}
			if err = ioutil.WriteFile(dest, data, info.Mode()); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("couldn't copy %s: %v", src, err)
	}
}

// add .gitkeep to generated empty directories
func addGitKeep(t *testing.T, p string) {
	if err := filepath.Walk(p, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		files, err := ioutil.ReadDir(p)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			f, err := os.OpenFile(filepath.Join(p, gitkeepfile), os.O_RDONLY|os.O_CREATE, 0666)
			if err != nil {
				return err
			}
			defer f.Close()
		}
		return nil
	}); err != nil {
		t.Fatalf("couldn't stamp git keep files: %v", err)
	}
}

// assertIdentical compares recursively all original and generated file content
func assertIdentical(t *testing.T, original, generated string) {
	if err := filepath.Walk(original, func(f string, fi os.FileInfo, err error) error {
		relp := strings.TrimPrefix(f, original)
		// root path
		if relp == "" {
			return nil
		}
		relp = relp[1:]
		p := filepath.Join(generated, relp)

		// .keep files are only for keeping directory creations in remote git repo
		if filepath.Base(p) == gitkeepfile {
			return nil
		}

		fo, err := os.Stat(p)
		if err != nil {
			t.Fatalf("%s doesn't exist while %s does", p, f)
		}

		if fi.IsDir() {
			if !fo.IsDir() {
				t.Fatalf("%s is a directory and %s isn't", f, p)
			}
			// else, it's a directory as well and we are done.
			return nil
		}

		wanted, err := ioutil.ReadFile(f)
		if err != nil {
			t.Fatalf("Couldn't read %s: %v", f, err)
		}
		actual, err := ioutil.ReadFile(p)
		if err != nil {
			t.Fatalf("Couldn't read %s: %v", p, err)
		}
		if !bytes.Equal(actual, wanted) {
			t.Errorf("%s and %s content differs:\nACTUAL:\n%s\n\nWANTED:\n%s", p, f, actual, wanted)
		}
		return nil
	}); err != nil {
		t.Fatalf("err: %s", err)
	}

	// on the other side, check that all generated items are in origin
	if err := filepath.Walk(generated, func(f string, _ os.FileInfo, err error) error {
		relp := strings.TrimPrefix(f, generated)
		// root path
		if relp == "" {
			return nil
		}
		relp = relp[1:]
		p := filepath.Join(original, relp)

		// .keep files are only for keeping directory creations in remote git repo
		if filepath.Base(p) == gitkeepfile {
			return nil
		}

		if _, err := os.Stat(p); err != nil {
			t.Errorf("%s doesn't exist while %s does", p, f)
		}
		return nil
	}); err != nil {
		t.Fatalf("err: %s", err)
	}
}
