package pack

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/ghodss/yaml"

	"k8s.io/helm/pkg/proto/hapi/chart"
)

// Archive creates an archived pack to the given directory.
//
// This takes an existing pack and a destination directory.
//
// If the directory is /foo, and the pack is named bar, this
// will generate /foo/bar.tgz.
//
// This returns the absolute path to the pack archive file.
func Archive(p *Pack, outDir string) (string, error) {

	// Create archive
	if fi, err := os.Stat(outDir); err != nil {
		return "", err
	} else if !fi.IsDir() {
		return "", fmt.Errorf("location %s is not a directory", outDir)
	}

	name := p.Metadata.Name
	if name == "" {
		return "", errors.New("no pack name specified")
	}

	filename := fmt.Sprintf("%s.tgz", name)
	filename = filepath.Join(outDir, filename)
	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}

	// Wrap in gzip writer
	zipper := gzip.NewWriter(f)
	// zipper.Head.Extra = headerBytes // Helm uses this. Why?
	zipper.Header.Comment = "Draft"

	// Wrap in tar writer
	twriter := tar.NewWriter(zipper)
	rollback := false
	defer func() {
		twriter.Close()
		zipper.Close()
		f.Close()
		if rollback {
			os.Remove(filename)
		}
	}()

	if err := writePackTarContent(twriter, p); err != nil {
		rollback = true
	}

	return filename, err
}

func writePackTarContent(out *tar.Writer, p *Pack) error {

	base := p.Metadata.Name

	// write the metadata
	pData, err := yaml.Marshal(p.Metadata)
	if err != nil {
		return err
	}

	if err := writeToTar(out, path.Join(base, MetadataName), pData); err != nil {
		return err
	}

	// write the content of the chart
	chartName := p.Chart.Metadata.Name
	// HACK(rodcloutier): writeTarContents uses the chart name as the dirname. Because we want to
	// write it to chart/, we name the chart 'chart'
	p.Chart.Metadata.Name = ChartDir
	if err := writeTarContents(out, p.Chart, base); err != nil {
		return err
	}
	p.Chart.Metadata.Name = chartName

	// write the dockerfile
	if err = writeToTar(out, path.Join(base, DockerfileName), p.Dockerfile); err != nil {
		return err
	}
	// write the detect script
	if len(p.DetectScript) > 0 {
		if err = writeToTar(out, path.Join(base, DetectName), p.DetectScript); err != nil {
			return err
		}
	}

	return nil
}

// --------------------------------------------------------------------------
// Note: This comes directly from Helm, we should be able to
// share the content
// helm/pkg/chartutil/save.go:writeTarContents
func writeTarContents(out *tar.Writer, c *chart.Chart, prefix string) error {
	base := filepath.Join(prefix, c.Metadata.Name)

	// Save Chart.yaml
	cdata, err := yaml.Marshal(c.Metadata)
	if err != nil {
		return err
	}
	if err := writeToTar(out, base+"/Chart.yaml", cdata); err != nil {
		return err
	}

	// Save values.yaml
	if c.Values != nil && len(c.Values.Raw) > 0 {
		if err := writeToTar(out, base+"/values.yaml", []byte(c.Values.Raw)); err != nil {
			return err
		}
	}

	// Save templates
	for _, f := range c.Templates {
		n := filepath.Join(base, f.Name)
		if err := writeToTar(out, n, f.Data); err != nil {
			return err
		}
	}

	// Save files
	for _, f := range c.Files {
		n := filepath.Join(base, f.TypeUrl)
		if err := writeToTar(out, n, f.Value); err != nil {
			return err
		}
	}

	// Save dependencies
	for _, dep := range c.Dependencies {
		if err := writeTarContents(out, dep, base+"/charts"); err != nil {
			return err
		}
	}
	return nil
}

// writeToTar writes a single file to a tar archive.
func writeToTar(out *tar.Writer, name string, body []byte) error {
	// TODO: Do we need to create dummy parent directory names if none exist?
	h := &tar.Header{
		Name: name,
		Mode: 0755,
		Size: int64(len(body)),
	}
	if err := out.WriteHeader(h); err != nil {
		return err
	}
	if _, err := out.Write(body); err != nil {
		return err
	}
	return nil
}
