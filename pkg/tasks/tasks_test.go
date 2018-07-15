package tasks

import (
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tasksFile, err := Load(filepath.Join("testdata", "tasks.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(tasksFile.PreUp) != 1 {
		t.Errorf("Expected 1 pre-up task, got %v", len(tasksFile.PreUp))
	}
	if len(tasksFile.PostDeploy) != 1 {
		t.Errorf("Expected 1 post-deploy task, got %v", len(tasksFile.PostDeploy))
	}
	if len(tasksFile.PostDelete) != 1 {
		t.Errorf("Expected 1 cleanup task, got %v", len(tasksFile.PostDeploy))
	}
}

func TestLoadError(t *testing.T) {
	_, err := Load(filepath.Join("testdata", "nonexistent.yaml"))
	if err == nil {
		t.Error(err)
	}

	_, err = Load(filepath.Join("testdata", "malformedTasks.yaml"))
	if err == nil {
	}
}

func TestRun(t *testing.T) {
	taskFile, err := Load(filepath.Join("testdata", "tasks.toml"))
	if err != nil {
		t.Fatal(err)
	}

	results, err := taskFile.Run(PreUp, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("Expected one pre-up command to be run, got %v", len(results))
	}

	results, _ = taskFile.Run(PostDeploy, "testpod")
	if len(results) != 1 {
		t.Errorf("Expected one post deploy command to be run, got %v", len(results))
	}

	results, _ = taskFile.Run(PostDelete, "")
	if len(results) != 1 {
		t.Errorf("Expected one cleanup command to be run, got %v", len(results))
	}
}
