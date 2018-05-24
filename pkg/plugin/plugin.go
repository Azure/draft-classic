package plugin

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/docker/docker/pkg/archive"
)

// Plugin provides metadata to install a piece of software.
type Plugin struct {
	// The canonical name of the software.
	Name string
	// A (short) description of the software.
	Description string
	// The license identifier for the software.
	License string
	// The homepage URL for the software.
	Homepage string
	// Caveats inform the user about any specific caveats regarding the software.
	Caveats string
	// The version of the software.
	Version string
	// The list of binary distributions available for the software.
	Packages []*Package
	// UseTunnel defines whether or not this plugin requires a tunneled connection to Tiller.
	UseTunnel bool
}

// Package provides metadata to install a piece of software on a given operating system and architecture.
type Package struct {
	// the running program's operating system target. One of darwin, linux, windows, and so on.
	OS string
	// the running program's architecture target. One of 386, amd64, arm, s390x, and so on.
	Arch string
	// The URL used to download the binary distribution for this version of the software. The file must be a gzipped tarball (.tar.gz) or a zipfile (.zip) for unpacking.
	URL string
	// Additional URLs for this version of the software.
	Mirrors []string
	// To verify the cached download's integrity and security, we verify the SHA-256 hash matches what we've declared in the software.
	SHA256 string
	// Path is the path to the executable relative from the root of the unpacked archive. After it is unpacked, Path is made executable (chmod +x).
	Path string
}

// Install attempts to install the plugin, returning errors if it fails.
func (p *Plugin) Install(home Home) error {
	plugDir := filepath.Join(home.Installed(), p.Name, p.Version)
	pkg := p.GetPackage(runtime.GOOS, runtime.GOARCH)
	if pkg == nil {
		return fmt.Errorf("plugin '%s' does not support the current platform (%s/%s)", p.Name, runtime.GOOS, runtime.GOARCH)
	}
	cachedFilePath, err := downloadCachedFileToPath(home.Cache(), pkg.URL)
	if err != nil {
		return err
	}
	if err := checksumVerifyPath(cachedFilePath, pkg.SHA256); err != nil {
		return fmt.Errorf("shasum verify check failed: %v", err)
	}

	if err := os.MkdirAll(plugDir, 0755); err != nil {
		return err
	}
	unarchiveOrCopy(cachedFilePath, plugDir)

	if err := os.Chmod(filepath.Join(plugDir, pkg.Path), 0755); err != nil {
		return err
	}

	if p.Caveats != "" {
		fmt.Println(p.Caveats)
	}

	return nil
}

// Installed checks to see if this plugin is installed. This is actually just a check for if the
// directory exists and is not empty.
func (p *Plugin) Installed(home Home) bool {
	files, err := ioutil.ReadDir(filepath.Join(home.Installed(), p.Name, p.Version))
	if err != nil {
		return false
	}
	return len(files) > 0
}

// Uninstall attempts to uninstall the package, returning errors if it fails.
func (p *Plugin) Uninstall(home Home) error {
	pkg := p.GetPackage(runtime.GOOS, runtime.GOARCH)
	if pkg == nil {
		return nil
	}
	plugDir := filepath.Join(home.Installed(), p.Name, p.Version)
	return os.RemoveAll(plugDir)
}

func unarchiveOrCopy(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if archive.IsArchivePath(src) {
		return archive.Untar(in, dest, &archive.TarOptions{NoLchown: true})
	} else if isZipPath(src) {
		in.Close()
		return unzip(src, dest)
	}
	out, err := os.Create(filepath.Join(dest, filepath.Base(src)))
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// GetPackage does a lookup for a package supporting the given os/arch. If none were found, this
// returns nil.
func (p *Plugin) GetPackage(os, arch string) *Package {
	for _, pkg := range p.Packages {
		if pkg.OS == os && pkg.Arch == arch {
			return pkg
		}
	}
	return nil
}

// downloadCachedFileToPath will download a file from the given url to a directory, returning the
// path to the cached file. If it already exists, it'll skip downloading the file and just return
// the path to the cached file.
func downloadCachedFileToPath(dir string, url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(dir, path.Base(req.URL.Path))

	if _, err = os.Stat(filePath); err == nil {
		return filePath, nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return filePath, err
}

func isZipPath(path string) bool {
	_, err := zip.OpenReader(path)
	return err == nil
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, zf := range r.File {
		dst, err := os.Create(filepath.Join(dest, zf.Name))
		if err != nil {
			return err
		}
		defer dst.Close()
		src, err := zf.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		io.Copy(dst, src)
	}
	return nil
}

func checksumVerifyPath(path string, checksum string) error {
	hasher := sha256.New()
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(hasher, f); err != nil {
		return err
	}

	actualChecksum := fmt.Sprintf("%x", hasher.Sum(nil))
	if strings.Compare(actualChecksum, checksum) != 0 {
		return fmt.Errorf("checksums differ for %s: expected '%s', got '%s'", path, checksum, actualChecksum)
	}
	return nil
}
