package dockerutil

import (
	"io"
	"os"
)

// Create creates a new Dockerfile at dest with the given data. It returns an
// error if dest already exists.
func Create(dest string, data io.Reader) error {
	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, data)
	return err
}

// CreateFrom creates a new Dockerfile, but scaffolds it from the src Dockerfile.
func CreateFrom(dest, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	return Create(dest, f)
}
