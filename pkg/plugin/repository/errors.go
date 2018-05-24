package repository

import "errors"

var (
	// ErrMissingMetadata indicates that Plugins/ is missing.
	ErrMissingMetadata = errors.New("repository does not have a Plugins/ directory")
	// ErrExists indicates that a repository already exists
	ErrExists = errors.New("repository already exists")
	// ErrDoesNotExist indicates that a repository does not exist
	ErrDoesNotExist = errors.New("repository does not exist")
	// ErrHomeMissing indicates that the directory expected to contain repositorys does not exist
	ErrHomeMissing = errors.New(`repository home "$(draft home)/repositories" does not exist`)
	// ErrMissingSource indicates that information about the source of the repository was not found
	ErrMissingSource = errors.New("cannot get information about the source of this repository")
	// ErrRepoDirty indicates that the repository repo was modified
	ErrRepoDirty = errors.New("repository repo is in a dirty git tree state so we cannot update. Try removing and adding this repository back")
	// ErrVersionDoesNotExist indicates that the request version does not exist
	ErrVersionDoesNotExist = errors.New("requested version does not exist")
)
