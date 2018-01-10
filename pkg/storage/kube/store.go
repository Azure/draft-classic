package kube

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"github.com/Azure/draft/pkg/storage"
	"github.com/golang/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

// Store represents a Kubernetes configmap storage engine for a storage.Object .
type Store struct {
	client    k8s.Interface
	namespace string
}

var _ storage.Store = (*Store)(nil)

func NewStore(c k8s.Interface, namespace string) *Store {
	return &Store{c, namespace}
}

// DeleteBuilds deletes all draft builds for the application specified by appName.
func (s *Store) DeleteBuilds(ctx context.Context, appName string) ([]*storage.Object, error) {
	builds, err := s.GetBuilds(ctx, appName)
	if err != nil {
		return nil, err
	}
	err = s.client.CoreV1().ConfigMaps(s.namespace).Delete(appName, &metav1.DeleteOptions{})
	return builds, err
}

// DeleteBuild deletes the draft build given by buildID for the application specified by appName.
func (s *Store) DeleteBuild(ctx context.Context, appName, buildID string) (obj *storage.Object, err error) {
	var cfgmap *v1.ConfigMap
	if cfgmap, err = s.client.CoreV1().ConfigMaps(s.namespace).Get(appName, metav1.GetOptions{}); err != nil {
		return nil, err
	}
	if build, ok := cfgmap.Data[buildID]; ok {
		if obj, err = decode(build); err != nil {
			return nil, err
		}
		delete(cfgmap.Data, buildID)
		_, err = s.client.CoreV1().ConfigMaps(s.namespace).Update(cfgmap)
		return obj, err
	}
	return nil, fmt.Errorf("application %q storage object %q not found", appName, buildID)
}

// CreateBuild stores a draft.Build for the application specified by appName.
func (s *Store) CreateBuild(ctx context.Context, appName string, build *storage.Object) error {
	content, err := encode(build)
	if err != nil {
		return err
	}
	cfgmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"heritage": "draft",
				"appname":  appName,
			},
		},
		Data: map[string]string{build.BuildID: content},
	}
	_, err = s.client.CoreV1().ConfigMaps(s.namespace).Create(cfgmap)
	return err
}

// GetBuilds returns a slice of builds for the given app name.
func (s *Store) GetBuilds(ctx context.Context, appName string) (builds []*storage.Object, err error) {
	var cfgmap *v1.ConfigMap
	if cfgmap, err = s.client.CoreV1().ConfigMaps(s.namespace).Get(appName, metav1.GetOptions{}); err != nil {
		return nil, err
	}
	for _, obj := range cfgmap.Data {
		build, err := decode(obj)
		if err != nil {
			return nil, err
		}
		builds = append(builds, build)
	}
	return builds, nil
}

// GetBuild returns the build associated with buildID for the specified app name.
func (s *Store) GetBuild(ctx context.Context, appName, buildID string) (obj *storage.Object, err error) {
	var cfgmap *v1.ConfigMap
	if cfgmap, err = s.client.CoreV1().ConfigMaps(s.namespace).Get(appName, metav1.GetOptions{}); err != nil {
		return nil, err
	}
	if data, ok := cfgmap.Data[buildID]; ok {
		if obj, err = decode(data); err != nil {
			return nil, err
		}
		return obj, nil
	}
	return nil, fmt.Errorf("application %q storage object %q not found", appName, buildID)
}

// encode returns the base64 encoding of protobuf encoded storage.Object.
//
// err != nil if the protobuf marshaling fails; otherwise nil.
func encode(obj *storage.Object) (string, error) {
	b, err := proto.Marshal(obj)
	if err != nil {
		return "", err
	}
	return b64.StdEncoding.EncodeToString(b), nil
}

// decode returns the storage.Object from its base64 encoded protobuf.
//
// err != nil if decoding fails.
func decode(str string) (*storage.Object, error) {
	b, err := b64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	var obj storage.Object
	if err := proto.Unmarshal(b, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}
