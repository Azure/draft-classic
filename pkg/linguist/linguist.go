package linguist

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Azure/draft/pkg/osutil"
	log "github.com/Sirupsen/logrus"
	"github.com/generaltso/linguist"
)

var isIgnored func(string) bool

// used for displaying results
type (
	// Language is the programming langage and the percentage on how sure linguist feels about its
	// decision.
	Language struct {
		Language string  `json:"language"`
		Percent  float64 `json:"percent"`
		// Color represents the color associated with the language in HTML hex notation.
		Color string `json:"color"`
	}
)

type sortableResult []*Language

func (s sortableResult) Len() int {
	return len(s)
}

func (s sortableResult) Less(i, j int) bool {
	return s[i].Percent < s[j].Percent
}

func (s sortableResult) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func initGitIgnore(dir string) error {
	gitDirExists, err := osutil.Exists(filepath.Join(dir, ".git"))
	if err != nil {
		return err
	}
	gitignoreExists, err := osutil.Exists(filepath.Join(dir, ".gitignore"))
	if err != nil {
		return err
	}
	if gitDirExists && gitignoreExists {
		log.Debugln("found .git directory and .gitignore")

		f, err := os.Open(filepath.Join(dir, ".gitignore"))
		if err != nil {
			return err
		}

		pathlist, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		ignore := []string{}
		except := []string{}
		for _, path := range strings.Split(string(pathlist), "\n") {
			path = strings.TrimSpace(path)
			if len(path) == 0 || string(path[0]) == "#" {
				continue
			}
			isExcept := false
			if string(path[0]) == "!" {
				isExcept = true
				path = path[1:]
			}
			fields := strings.Split(path, " ")
			p := fields[len(fields)-1:][0]
			p = strings.Trim(p, string(filepath.Separator))
			if isExcept {
				except = append(except, p)
			} else {
				ignore = append(ignore, p)
			}
		}
		isIgnored = func(filename string) bool {
			for _, p := range ignore {
				if m, _ := filepath.Match(p, filename); m {
					for _, e := range except {
						if m, _ := filepath.Match(e, filename); m {
							return false
						}
					}
					return true
				}
			}
			return false
		}
	} else {
		log.Debugln("no .gitignore found")
		isIgnored = func(filename string) bool {
			return false
		}
	}
	return nil
}

// shoutouts to php
func fileGetContents(filename string) ([]byte, error) {
	log.Debugln("reading contents of", filename)

	// read only first 512 bytes of files
	contents := make([]byte, 512)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	_, err = f.Read(contents)
	f.Close()
	if err != io.EOF {
		if err != nil {
			return nil, err
		}
	}
	return contents, nil
}

// ProcessDir walks through a directory and returns a list of sorted languages within that directory.
func ProcessDir(dirname string) ([]*Language, error) {
	var (
		langs     = make(map[string]int)
		totalSize int
	)
	if err := initGitIgnore(dirname); err != nil {
		return nil, err
	}
	exists, err := osutil.Exists(dirname)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, os.ErrNotExist
	}
	filepath.Walk(dirname, func(path string, file os.FileInfo, err error) error {
		size := int(file.Size())
		log.Debugln("with file: ", path)
		log.Debugln(path, "is", size, "bytes")
		if size == 0 {
			log.Debugln(path, "is empty file, skipping")
			return nil
		}
		if isIgnored(path) {
			log.Debugln(path, "is ignored, skipping")
			if file.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if file.IsDir() {
			if file.Name() == ".git" {
				log.Debugln(".git directory, skipping")
				return filepath.SkipDir
			}
		} else if (file.Mode() & os.ModeSymlink) == 0 {
			if linguist.ShouldIgnoreFilename(path) {
				log.Debugln(path, ": filename should be ignored, skipping")
				return nil
			}

			byName := linguist.LanguageByFilename(path)
			if byName != "" {
				log.Debugln(path, "got result by name: ", byName)
				langs[byName] += size
				totalSize += size
				return nil
			}

			contents, err := fileGetContents(path)
			if err != nil {
				return err
			}

			if linguist.ShouldIgnoreContents(contents) {
				log.Debugln(path, ": contents should be ignored, skipping")
				return nil
			}

			hints := linguist.LanguageHints(path)
			log.Debugf("%s got language hints: %#v\n", path, hints)
			byData := linguist.LanguageByContents(contents, hints)

			if byData != "" {
				log.Debugln(path, "got result by data: ", byData)
				langs[byData] += size
				totalSize += size
				return nil
			}

			log.Debugln(path, "got no result!!")
			langs["(unknown)"] += size
			totalSize += size
		}
		return nil
	})

	results := []*Language{}
	for lang, size := range langs {
		l := &Language{
			Language: lang,
			Percent:  (float64(size) / float64(totalSize)) * 100.0,
			Color:    linguist.LanguageColor(lang),
		}
		results = append(results, l)
		log.Debugf("language: %s percent: %f color: %s", l.Language, l.Percent, l.Color)
	}
	sort.Sort(sort.Reverse(sortableResult(results)))
	return results, nil
}
