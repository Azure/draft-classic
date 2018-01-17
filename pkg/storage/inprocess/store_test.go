package inprocess

import (
	"context"
	"github.com/Azure/draft/pkg/storage"
	"reflect"
	"testing"
)

func TestStoreDeleteBuilds(t *testing.T) {
	var (
		store = NewStoreWithMocks()
		ctx   = context.TODO()
	)
	builds, err := store.DeleteBuilds(ctx, "app1")
	if err != nil {
		t.Fatalf("failed to delete build entries: %v", err)
	}
	if len(store.builds["app1"]) > 0 {
		t.Fatal("expected build entries to empty")
	}
	assertEqual(t, "DeleteBuilds", builds, []*storage.Object{
		&storage.Object{BuildID: "foo1", Release: "bar1", ContextID: []byte("foobar1")},
		&storage.Object{BuildID: "foo2", Release: "bar2", ContextID: []byte("foobar2")},
		&storage.Object{BuildID: "foo3", Release: "bar3", ContextID: []byte("foobar3")},
		&storage.Object{BuildID: "foo4", Release: "bar4", ContextID: []byte("foobar4")},
	})
}

func TestStoreDeleteBuild(t *testing.T) {
	var (
		store = NewStoreWithMocks()
		ctx   = context.TODO()
	)
	build, err := store.DeleteBuild(ctx, "app1", "foo1")
	if err != nil {
		t.Fatalf("failed to delete build entry: %v", err)
	}
	assertEqual(t, "DeleteBuild", build, &storage.Object{
		BuildID:   "foo1",
		Release:   "bar1",
		ContextID: []byte("foobar1"),
	})
}

func TestStoreCreateBuild(t *testing.T) {
	var (
		build = &storage.Object{BuildID: "foo", Release: "bar", ContextID: []byte("foobar")}
		store = NewStoreWithMocks()
		ctx   = context.TODO()
	)
	if err := store.CreateBuild(ctx, "app2", build); err != nil {
		t.Fatalf("failed to create storage entry: %v", err)
	}
	alt, err := store.GetBuild(ctx, "app2", build.BuildID)
	if err != nil {
		t.Fatalf("failed to get build entry: %v", err)
	}
	assertEqual(t, "CreateBuild", build, alt)
}

func TestStoreGetBuilds(t *testing.T) {
	var (
		store = NewStoreWithMocks()
		ctx   = context.TODO()
	)
	// make sure the build is returnable by appID
	ls, err := store.GetBuilds(ctx, "app1")
	if err != nil {
		t.Fatalf("could not get builds: %v", err)
	}
	assertEqual(t, "GetBuilds", ls, []*storage.Object{
		&storage.Object{BuildID: "foo1", Release: "bar1", ContextID: []byte("foobar1")},
		&storage.Object{BuildID: "foo2", Release: "bar2", ContextID: []byte("foobar2")},
		&storage.Object{BuildID: "foo3", Release: "bar3", ContextID: []byte("foobar3")},
		&storage.Object{BuildID: "foo4", Release: "bar4", ContextID: []byte("foobar4")},
	})
	// try fetching a build with an unknown appID; should fail.
	if alt, err := store.GetBuilds(ctx, "bad"); err == nil {
		t.Fatalf("want err != nil; got alt: %+v", alt)
	}
}

func TestStoreGetBuild(t *testing.T) {
	var (
		store = NewStoreWithMocks()
		ctx   = context.TODO()
	)
	// make sure the build is returnable by appID
	obj, err := store.GetBuild(ctx, "app1", "foo1")
	if err != nil {
		t.Fatalf("could not get build: %v", err)
	}
	assertEqual(t, "GetBuild", obj, &storage.Object{
		BuildID:   "foo1",
		Release:   "bar1",
		ContextID: []byte("foobar1"),
	})
	// try fetching a build with an unknown appID; should fail.
	if alt, err := store.GetBuild(ctx, "bad", ""); err == nil {
		t.Fatalf("want err != nil; got alt: %+v", alt)
	}
}

func NewStoreWithMocks() *Store {
	store := NewStore()
	store.builds["app1"] = []*storage.Object{
		&storage.Object{BuildID: "foo1", Release: "bar1", ContextID: []byte("foobar1")},
		&storage.Object{BuildID: "foo2", Release: "bar2", ContextID: []byte("foobar2")},
		&storage.Object{BuildID: "foo3", Release: "bar3", ContextID: []byte("foobar3")},
		&storage.Object{BuildID: "foo4", Release: "bar4", ContextID: []byte("foobar4")},
	}
	return store
}

func assertEqual(t *testing.T, label string, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("failed equality for %s", label)
	}
}
