package inprocess

import (
	"context"
	"fmt"

	"github.com/Azure/draft/pkg/storage"
)

// Store is an inprocess storage engine for draft.
type Store struct {
	// builds is mapping of app name to storage objects.
	builds map[string][]*storage.Object
}

// compile-time guarantee that Store implements storage.Store
var _ storage.Store = (*Store)(nil)

// NewStore returns a new *inprocess.Store.
func NewStore() *Store {
	return &Store{builds: make(map[string][]*storage.Object)}
}

// DeleteBuilds deletes all draft builds for the application specified by appName.
func (s *Store) DeleteBuilds(ctx context.Context, appName string) ([]*storage.Object, error) {
	h, ok := s.builds[appName]
	if !ok {
		return nil, fmt.Errorf("storage history for %q not found", appName)
	}
	delete(s.builds, appName)
	return h, nil
}

// DeleteBuild deletes the draft build given by buildID for the application specified by appName.
func (s *Store) DeleteBuild(ctx context.Context, appName, buildID string) (*storage.Object, error) {
	h, ok := s.builds[appName]
	if !ok {
		return nil, fmt.Errorf("storage history for %q not found", appName)
	}
	for i, o := range h {
		if buildID == o.BuildID {
			s.builds[appName] = append(h[:i], h[i+1:]...)
			return o, nil
		}
	}
	return nil, fmt.Errorf("application %q storage object %q not found", appName, buildID)
}

// CreateBuild stores a draft.Build for the application specified by appName.
func (s *Store) CreateBuild(ctx context.Context, appName string, build *storage.Object) error {
	if _, ok := s.builds[appName]; ok {
		s.builds[appName] = append(s.builds[appName], build)
		return nil
	}
	s.builds[appName] = []*storage.Object{build}
	return nil
}

// GetBuilds returns a slice of builds for the given app name.
func (s *Store) GetBuilds(ctx context.Context, appName string) ([]*storage.Object, error) {
	h, ok := s.builds[appName]
	if !ok {
		return nil, fmt.Errorf("storage history for %q not found", appName)
	}
	return h, nil
}

// GetBuild returns the build associated with buildID for the specified app name.
func (s *Store) GetBuild(ctx context.Context, appName, buildID string) (*storage.Object, error) {
	h, ok := s.builds[appName]
	if !ok {
		return nil, fmt.Errorf("storage history for %q not found", appName)
	}
	for _, o := range h {
		if buildID == o.BuildID {
			return o, nil
		}
	}
	return nil, fmt.Errorf("application %q storage object %q not found", appName, buildID)
}
