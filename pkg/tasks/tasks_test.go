package tasks

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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

	results, err := taskFile.Run(DefaultRunner, PreUp, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("Expected one pre-up command to be run, got %v", len(results))
	}

	results, _ = taskFile.Run(DefaultRunner, PostDeploy, "testpod")
	if len(results) != 1 {
		t.Errorf("Expected one post deploy command to be run, got %v", len(results))
	}

	results, _ = taskFile.Run(DefaultRunner, PostDelete, "")
	if len(results) != 1 {
		t.Errorf("Expected one cleanup command to be run, got %v", len(results))
	}
}

func TestRun_Interpolation(t *testing.T) {
	os.Setenv("DRAFT_HELLO", "hello")
	defer os.Unsetenv("DRAFT_HELLO")

	cases := []struct {
		description string
		tasks       *Tasks
		kind        string
		podName     string
		expectedCmd []string
	}{
		{
			description: "PreUp with environment variable",
			tasks: &Tasks{
				PreUp: map[string]string{
					"echo": "echo $DRAFT_HELLO",
				},
			},
			kind:        PreUp,
			expectedCmd: []string{"echo", "hello"},
		},
		{
			description: "PostDeploy with environment variable",
			tasks: &Tasks{
				PostDeploy: map[string]string{
					"echo": "echo $DRAFT_HELLO",
				},
			},
			kind:        PostDeploy,
			podName:     "pod-1234",
			expectedCmd: []string{"kubectl", "exec", "pod-1234", "--", "echo", "hello"},
		},
		{
			description: "PostDelete with environment variable",
			tasks: &Tasks{
				PostDelete: map[string]string{
					"echo": "echo $DRAFT_HELLO",
				},
			},
			kind:        PostDelete,
			expectedCmd: []string{"echo", "hello"},
		},
		{
			description: "PreUp with complicated interpolation",
			tasks: &Tasks{
				PreUp: map[string]string{
					"echo": "echo $DRAFT_HELLO/$DRAFT_HELLO",
				},
			},
			kind:        PreUp,
			expectedCmd: []string{"echo", "hello/hello"},
		},
		{
			description: "PreUp with escaped variables",
			tasks: &Tasks{
				PreUp: map[string]string{
					"echo": "echo $DRAFT_HELLO/$$DRAFT_HELLO/\\$DRAFT_HELLO",
				},
			},
			kind:        PreUp,
			expectedCmd: []string{"echo", "hello/$DRAFT_HELLO/$DRAFT_HELLO"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var got []string
			runner := func(cmd *exec.Cmd) error {
				got = cmd.Args
				return nil
			}

			_, err := tc.tasks.Run(runner, tc.kind, tc.podName)
			if err != nil {
				t.Fatal(err)
			} else if !reflect.DeepEqual(got, tc.expectedCmd) {
				t.Errorf("got cmd: %v, want: %v", got, tc.expectedCmd)
			}
		})
	}
}
