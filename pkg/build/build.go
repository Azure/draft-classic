package build

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/Sirupsen/logrus"

	// docker deps
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/cli/command/image/build"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"

	// helm deps
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/strvals"

	"github.com/Azure/draft/pkg/draft/manifest"
	"github.com/Azure/draft/pkg/draft/pack"
)

// Context contains information about the application and the environment
//  that will be pushed up to the server
type Context struct {
	Env     *manifest.Environment
	EnvName string
	AppDir  string
	Chart   *chart.Chart
	Values  *chart.Config
	SrcName string
	Archive []byte
}

// LoadWithEnv takes the directory of the application and the environment the application
//  will be pushed to and returns a Context object with a merge of environment and app
//  information
func LoadWithEnv(appdir, whichenv string) (*Context, error) {
	ctx := &Context{AppDir: appdir, EnvName: whichenv}
	// read draft.toml from appdir.
	b, err := ioutil.ReadFile(filepath.Join(appdir, "draft.toml"))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("draft.toml does not exist")
		}
		return nil, fmt.Errorf("failed to read draft.toml from %q: %v", appdir, err)
	}
	// unmarshal draft.toml into new manifest.
	mfst := manifest.New()
	if err := toml.Unmarshal(b, mfst); err != nil {
		return nil, fmt.Errorf("failed to unmarshal draft.toml from %q: %v", appdir, err)
	}
	// if environment does not exist return error.
	var ok bool
	if ctx.Env, ok = mfst.Environments[whichenv]; !ok {
		return nil, fmt.Errorf("no environment named %q in draft.toml", whichenv)
	}
	// load the chart and the build archive; if a chart directory is present
	// this will be given priority over the chart archive specified by the
	// `chart-tar` field in the draft.toml. If this is the case, then build-tar
	// is built from scratch. If no chart directory exists but a chart-tar and
	// build-tar exist, then these will be used for values extraction.
	if err := loadArchive(ctx); err != nil {
		return nil, fmt.Errorf("failed to load chart: %v", err)
	}
	// load values from chart and merge with env.Values.
	if err := loadValues(ctx); err != nil {
		return nil, fmt.Errorf("failed to parse chart values: %v", err)
	}
	return ctx, nil
}

// loadArchive loads the chart package and build archive.
// Precedence is given to the `build-tar` and `chart-tar`
// indicated in the `draft.toml` if present. Otherwise,
// loadArchive loads the chart directory and archives the
// app directory to send to the draft server.
func loadArchive(ctx *Context) (err error) {
	if ctx.Env.BuildTarPath != "" && ctx.Env.ChartTarPath != "" {
		b, err := ioutil.ReadFile(ctx.Env.BuildTarPath)
		if err != nil {
			return fmt.Errorf("failed to load build archive %q: %v", ctx.Env.BuildTarPath, err)
		}
		ctx.SrcName = filepath.Base(ctx.Env.BuildTarPath)
		ctx.Archive = b

		ar, err := os.Open(ctx.Env.ChartTarPath)
		if err != nil {
			return err
		}
		if ctx.Chart, err = chartutil.LoadArchive(ar); err != nil {
			return fmt.Errorf("failed to load chart archive %q: %v", ctx.Env.ChartTarPath, err)
		}
		return nil
	}
	if err = archiveSrc(ctx); err != nil {
		return err
	}
	// find the first directory in chart/ and assume that is the chart we want to deploy.
	chartDir := filepath.Join(ctx.AppDir, pack.ChartsDir)
	files, err := ioutil.ReadDir(chartDir)
	if err != nil {
		return err
	}
	var found bool
	for _, file := range files {
		if file.IsDir() {
			found = true
			if ctx.Chart, err = chartutil.Load(filepath.Join(chartDir, file.Name())); err != nil {
				return err
			}
		}
	}
	if !found {
		return ErrChartNotExist
	}
	return nil
}

func loadValues(ctx *Context) error {
	var vals = make(chartutil.Values)
	for _, val := range ctx.Env.Values {
		if err := strvals.ParseInto(val, vals.AsMap()); err != nil {
			return fmt.Errorf("failed to parse %q from draft.toml: %v", val, err)
		}
	}
	s, err := vals.YAML()
	if err != nil {
		return fmt.Errorf("failed to encode values: %v", err)
	}
	ctx.Values = &chart.Config{Raw: s}
	return nil
}

func archiveSrc(ctx *Context) error {
	contextDir, relDockerfile, err := build.GetContextFromLocalDir(ctx.AppDir, "")
	if err != nil {
		return fmt.Errorf("unable to prepare docker context: %s", err)
	}
	// canonicalize dockerfile name to a platform-independent one
	relDockerfile, err = archive.CanonicalTarNameForPath(relDockerfile)
	if err != nil {
		return fmt.Errorf("cannot canonicalize dockerfile path %s: %v", relDockerfile, err)
	}
	f, err := os.Open(filepath.Join(contextDir, ".dockerignore"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	defer f.Close()

	var excludes []string
	if err == nil {
		excludes, err = dockerignore.ReadAll(f)
		if err != nil {
			return err
		}
	}

	// do not include the chart directory. That will be packaged separately.
	excludes = append(excludes, filepath.Join(contextDir, "chart"))
	if err := build.ValidateContextDirectory(contextDir, excludes); err != nil {
		return fmt.Errorf("error checking docker context: '%s'", err)
	}

	// If .dockerignore mentions .dockerignore or the Dockerfile
	// then make sure we send both files over to the daemon
	// because Dockerfile is, obviously, needed no matter what, and
	// .dockerignore is needed to know if either one needs to be
	// removed. The daemon will remove them for us, if needed, after it
	// parses the Dockerfile. Ignore errors here, as they will have been
	// caught by validateContextDirectory above.
	var includes = []string{"."}
	keepThem1, _ := fileutils.Matches(".dockerignore", excludes)
	keepThem2, _ := fileutils.Matches(relDockerfile, excludes)
	if keepThem1 || keepThem2 {
		includes = append(includes, ".dockerignore", relDockerfile)
	}

	logrus.Debugf("INCLUDES: %v", includes)
	logrus.Debugf("EXCLUDES: %v", excludes)
	rc, err := archive.TarWithOptions(contextDir, &archive.TarOptions{
		Compression:     archive.Gzip,
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
	})
	if err != nil {
		return err
	}
	defer rc.Close()

	var b bytes.Buffer
	if _, err := io.Copy(&b, rc); err != nil {
		return err
	}
	ctx.SrcName = "build.tar.gz"
	ctx.Archive = b.Bytes()
	return nil
}
