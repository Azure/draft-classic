package kube

import (
	"context"
	"fmt"

	"github.com/Azure/draft/pkg/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// ConfigMaps represents a Kubernetes configmap storage engine for a storage.Object .
type ConfigMaps struct {
	impl corev1.ConfigMapInterface
}

var _ storage.Store = (*ConfigMaps)(nil)

func NewConfigMaps(impl corev1.ConfigMapInterface) *ConfigMaps {
	return &ConfigMaps{impl}
}

// DeleteBuilds deletes all draft builds for the application specified by appName.
func (this *ConfigMaps) DeleteBuilds(ctx context.Context, appName string) ([]*storage.Object, error) {
	builds, err := this.GetBuilds(ctx, appName)
	if err != nil {
		return nil, err
	}
	err = this.impl.Delete(appName, &metav1.DeleteOptions{})
	return builds, err
}

// DeleteBuild deletes the draft build given by buildID for the application specified by appName.
func (this *ConfigMaps) DeleteBuild(ctx context.Context, appName, buildID string) (obj *storage.Object, err error) {
	var cfgmap *v1.ConfigMap
	if cfgmap, err = this.impl.Get(appName, metav1.GetOptions{}); err != nil {
		return nil, err
	}
	if build, ok := cfgmap.Data[buildID]; ok {
		if obj, err = storage.DecodeString(build); err != nil {
			return nil, err
		}
		delete(cfgmap.Data, buildID)
		_, err = this.impl.Update(cfgmap)
		return obj, err
	}
	return nil, fmt.Errorf("application %q storage object %q not found", appName, buildID)
}

// CreateBuild stores a draft.Build for the application specified by appName.
func (this *ConfigMaps) CreateBuild(ctx context.Context, appName string, build *storage.Object) error {
	cfgmap, err := newConfigMap(appName, build)
	if err != nil {
		return err
	}
	_, err = this.impl.Create(cfgmap)
	return err
}

// GetBuilds returns a slice of builds for the given app name.
func (this *ConfigMaps) GetBuilds(ctx context.Context, appName string) (builds []*storage.Object, err error) {
	var cfgmap *v1.ConfigMap
	if cfgmap, err = this.impl.Get(appName, metav1.GetOptions{}); err != nil {
		return nil, err
	}
	for _, obj := range cfgmap.Data {
		build, err := storage.DecodeString(obj)
		if err != nil {
			return nil, err
		}
		builds = append(builds, build)
	}
	return builds, nil
}

// GetBuild returns the build associated with buildID for the specified app name.
func (this *ConfigMaps) GetBuild(ctx context.Context, appName, buildID string) (obj *storage.Object, err error) {
	var cfgmap *v1.ConfigMap
	if cfgmap, err = this.impl.Get(appName, metav1.GetOptions{}); err != nil {
		return nil, err
	}
	if data, ok := cfgmap.Data[buildID]; ok {
		if obj, err = storage.DecodeString(data); err != nil {
			return nil, err
		}
		return obj, nil
	}
	return nil, fmt.Errorf("application %q storage object %q not found", appName, buildID)
}

// newConfigMap constructs a kubernetes ConfigMap object to store a build.
//
// Each configmap data entry is the base64 encoded string of a *storage.Object
// binary protobuf encoding.
func newConfigMap(appName string, build *storage.Object) (*v1.ConfigMap, error) {
	content, err := storage.EncodeToString(build)
	if err != nil {
		return nil, err
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
	return cfgmap, nil
}
