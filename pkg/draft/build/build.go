package build

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/oklog/ulid"
)

var (
	LogRoot             = "/var/run/draft/builds"
	BuildImgLogFilename = "docker-build.log"
	PushImgLogFilename  = "docker-push.log"
	ReleaseLogFilename  = "helm.log"
)

type Build struct {
	ID           string
	BuildImgLogs *bytes.Buffer
	PushImgLogs  *bytes.Buffer
	ReleaseLogs  *bytes.Buffer
}

func New() *Build {
	return &Build{
		ID:           getulid(),
		BuildImgLogs: new(bytes.Buffer),
		PushImgLogs:  new(bytes.Buffer),
		ReleaseLogs:  new(bytes.Buffer),
	}
}

func (b *Build) FlushToDisk() error {
	buildPath := filepath.Join(LogRoot, b.ID)
	logMap := map[string]*bytes.Buffer{
		filepath.Join(buildPath, BuildImgLogFilename): b.BuildImgLogs,
		filepath.Join(buildPath, PushImgLogFilename):  b.PushImgLogs,
		filepath.Join(buildPath, ReleaseLogFilename):  b.ReleaseLogs,
	}
	if err := os.MkdirAll(buildPath, 0755); err != nil {
		return err
	}
	for p, buf := range logMap {
		if err := ioutil.WriteFile(p, buf.Bytes(), 0644); err != nil {
			return err
		}
	}
	return nil
}

func (b *Build) LoadFromDisk() error {
	buildPath := filepath.Join(LogRoot, b.ID)
	logMap := map[string]*bytes.Buffer{
		filepath.Join(buildPath, BuildImgLogFilename): b.BuildImgLogs,
		filepath.Join(buildPath, PushImgLogFilename):  b.PushImgLogs,
		filepath.Join(buildPath, ReleaseLogFilename):  b.ReleaseLogs,
	}

	for p, buf := range logMap {
		b, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}
		io.Copy(buf, bytes.NewBuffer(b))
	}
	return nil
}

func getulid() string { return <-ulidc }

// A channel which returns build ulids.
var ulidc = make(chan string)

func init() {
	rnd := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	go func() {
		for {
			ulidc <- ulid.MustNew(ulid.Timestamp(time.Now().UTC()), rnd).String()
		}
	}()
}
