package installer

import (
	"reflect"
	"testing"

	"github.com/ghodss/yaml"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
	testcore "k8s.io/kubernetes/pkg/client/testing/core"
	"k8s.io/kubernetes/pkg/runtime"

	"github.com/deis/prow/pkg/version"
)

func TestDeploymentManifest(t *testing.T) {

	tests := []struct {
		name   string
		image  string
		canary bool
		expect string
	}{
		{"default", "", false, "quay.io/deis/prowd:" + version.Release},
		{"canary", "example.com/tiller", true, "quay.io/deis/prowd:canary"},
		{"custom", "example.com/prowd:latest", false, "example.com/prowd:latest"},
	}

	for _, tt := range tests {

		o, err := DeploymentManifest(api.NamespaceDefault, tt.image, tt.canary)
		if err != nil {
			t.Fatalf("%s: error %q", tt.name, err)
		}
		var dep extensions.Deployment
		if err := yaml.Unmarshal([]byte(o), &dep); err != nil {
			t.Fatalf("%s: error %q", tt.name, err)
		}

		if got := dep.Spec.Template.Spec.Containers[0].Image; got != tt.expect {
			t.Errorf("%s: expected image %q, got %q", tt.name, tt.expect, got)
		}

		if got := dep.Spec.Template.Spec.Containers[0].Env[0].Value; got != api.NamespaceDefault {
			t.Errorf("%s: expected namespace %q, got %q", tt.name, api.NamespaceDefault, got)
		}
	}
}

func TestServiceManifest(t *testing.T) {
	o, err := ServiceManifest(api.NamespaceDefault)
	if err != nil {
		t.Fatalf("error %q", err)
	}
	var svc api.Service
	if err := yaml.Unmarshal([]byte(o), &svc); err != nil {
		t.Fatalf("error %q", err)
	}

	if got := svc.ObjectMeta.Namespace; got != api.NamespaceDefault {
		t.Errorf("expected namespace %s, got %s", api.NamespaceDefault, got)
	}
}

func TestInstall(t *testing.T) {
	image := "quay.io/deis/prowd:v1.0.0"

	fc := &fake.Clientset{}
	fc.AddReactor("create", "deployments", func(action testcore.Action) (bool, runtime.Object, error) {
		obj := action.(testcore.CreateAction).GetObject().(*extensions.Deployment)
		l := obj.GetLabels()
		if reflect.DeepEqual(l, map[string]string{"app": "prow"}) {
			t.Errorf("expected labels = '', got '%s'", l)
		}
		i := obj.Spec.Template.Spec.Containers[0].Image
		if i != image {
			t.Errorf("expected image = '%s', got '%s'", image, i)
		}
		return true, obj, nil
	})
	fc.AddReactor("create", "services", func(action testcore.Action) (bool, runtime.Object, error) {
		obj := action.(testcore.CreateAction).GetObject().(*api.Service)
		l := obj.GetLabels()
		if reflect.DeepEqual(l, map[string]string{"app": "prow"}) {
			t.Errorf("expected labels = '', got '%s'", l)
		}
		n := obj.ObjectMeta.Namespace
		if n != api.NamespaceDefault {
			t.Errorf("expected namespace = '%s', got '%s'", api.NamespaceDefault, n)
		}
		return true, obj, nil
	})

	if err := Install(fc, api.NamespaceDefault, image, false, false); err != nil {
		t.Errorf("unexpected error: %#+v", err)
	}

	if actions := fc.Actions(); len(actions) != 2 {
		t.Errorf("unexpected actions: %v, expected 2 actions got %d", actions, len(actions))
	}
}

func TestInstall_canary(t *testing.T) {
	fc := &fake.Clientset{}
	fc.AddReactor("create", "deployments", func(action testcore.Action) (bool, runtime.Object, error) {
		obj := action.(testcore.CreateAction).GetObject().(*extensions.Deployment)
		i := obj.Spec.Template.Spec.Containers[0].Image
		if i != "quay.io/deis/prowd:canary" {
			t.Errorf("expected canary image, got '%s'", i)
		}
		return true, obj, nil
	})
	fc.AddReactor("create", "services", func(action testcore.Action) (bool, runtime.Object, error) {
		obj := action.(testcore.CreateAction).GetObject().(*api.Service)
		return true, obj, nil
	})

	if err := Install(fc, api.NamespaceDefault, "", true, false); err != nil {
		t.Errorf("unexpected error: %#+v", err)
	}

	if actions := fc.Actions(); len(actions) != 2 {
		t.Errorf("unexpected actions: %v, expected 2 actions got %d", actions, len(actions))
	}
}

func TestUpgrade(t *testing.T) {
	image := "quay.io/deis/prowd:v1.0.0"

	existingDeployment := deployment(api.NamespaceDefault, "imageToReplace", false)
	existingService := service(api.NamespaceDefault)

	fc := &fake.Clientset{}
	fc.AddReactor("get", "deployments", func(action testcore.Action) (bool, runtime.Object, error) {
		return true, existingDeployment, nil
	})
	fc.AddReactor("update", "deployments", func(action testcore.Action) (bool, runtime.Object, error) {
		obj := action.(testcore.UpdateAction).GetObject().(*extensions.Deployment)
		i := obj.Spec.Template.Spec.Containers[0].Image
		if i != image {
			t.Errorf("expected image = '%s', got '%s'", image, i)
		}
		return true, obj, nil
	})
	fc.AddReactor("get", "services", func(action testcore.Action) (bool, runtime.Object, error) {
		return true, existingService, nil
	})

	if err := Upgrade(fc, api.NamespaceDefault, image, false); err != nil {
		t.Errorf("unexpected error: %#+v", err)
	}

	if actions := fc.Actions(); len(actions) != 3 {
		t.Errorf("unexpected actions: %v, expected 3 actions got %d", actions, len(actions))
	}
}

func TestUpgrade_serviceNotFound(t *testing.T) {
	image := "quay.io/deis/prowd:v1.0.0"

	existingDeployment := deployment(api.NamespaceDefault, "imageToReplace", false)

	fc := &fake.Clientset{}
	fc.AddReactor("get", "deployments", func(action testcore.Action) (bool, runtime.Object, error) {
		return true, existingDeployment, nil
	})
	fc.AddReactor("update", "deployments", func(action testcore.Action) (bool, runtime.Object, error) {
		obj := action.(testcore.UpdateAction).GetObject().(*extensions.Deployment)
		i := obj.Spec.Template.Spec.Containers[0].Image
		if i != image {
			t.Errorf("expected image = '%s', got '%s'", image, i)
		}
		return true, obj, nil
	})
	fc.AddReactor("get", "services", func(action testcore.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(api.Resource("services"), "1")
	})
	fc.AddReactor("create", "services", func(action testcore.Action) (bool, runtime.Object, error) {
		obj := action.(testcore.CreateAction).GetObject().(*api.Service)
		n := obj.ObjectMeta.Namespace
		if n != api.NamespaceDefault {
			t.Errorf("expected namespace = '%s', got '%s'", api.NamespaceDefault, n)
		}
		return true, obj, nil
	})

	if err := Upgrade(fc, api.NamespaceDefault, image, false); err != nil {
		t.Errorf("unexpected error: %#+v", err)
	}

	if actions := fc.Actions(); len(actions) != 4 {
		t.Errorf("unexpected actions: %v, expected 4 actions got %d", actions, len(actions))
	}
}
