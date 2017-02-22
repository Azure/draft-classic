package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/deis/prow/pkg/prow/prowpath"
)

func TestManuallyProcessArgs(t *testing.T) {
	input := []string{
		"--debug",
		"--foo", "bar",
		"--host", "example.com",
		"--kube-context", "test1",
		"--home=/tmp",
		"command",
	}

	expectKnown := []string{
		"--debug", "--host", "example.com", "--kube-context", "test1", "--home=/tmp",
	}

	expectUnknown := []string{
		"--foo", "bar", "command",
	}

	known, unknown := manuallyProcessArgs(input)

	for i, k := range known {
		if k != expectKnown[i] {
			t.Errorf("expected known flag %d to be %q, got %q", i, expectKnown[i], k)
		}
	}
	for i, k := range unknown {
		if k != expectUnknown[i] {
			t.Errorf("expected unknown flag %d to be %q, got %q", i, expectUnknown[i], k)
		}
	}

}

func TestLoadPlugins(t *testing.T) {
	// Set prow home to point to testdata
	old := prowHome
	prowHome = "testdata/prowhome"
	defer func() {
		prowHome = old
	}()
	ph := prowpath.Home(homePath())

	out := bytes.NewBuffer(nil)
	cmd := &cobra.Command{}
	loadPlugins(cmd, ph, out)

	envs := strings.Join([]string{
		"fullenv",
		ph.Plugins() + "/fullenv",
		ph.Plugins(),
		ph.String(),
		os.Args[0],
	}, "\n")

	// Test that the YAML file was correctly converted to a command.
	tests := []struct {
		use    string
		short  string
		long   string
		expect string
		args   []string
	}{
		{"args", "echo args", "This echos args", "-a -b -c\n", []string{"-a", "-b", "-c"}},
		{"echo", "echo stuff", "This echos stuff", "hello\n", []string{}},
		{"env", "env stuff", "show the env", ph.String() + "\n", []string{}},
		{"fullenv", "show env vars", "show all env vars", envs + "\n", []string{}},
	}

	plugins := cmd.Commands()

	if len(plugins) != len(tests) {
		t.Fatalf("Expected %d plugins, got %d", len(tests), len(plugins))
	}

	for i := 0; i < len(plugins); i++ {
		out.Reset()
		tt := tests[i]
		pp := plugins[i]
		if pp.Use != tt.use {
			t.Errorf("%d: Expected Use=%q, got %q", i, tt.use, pp.Use)
		}
		if pp.Short != tt.short {
			t.Errorf("%d: Expected Use=%q, got %q", i, tt.short, pp.Short)
		}
		if pp.Long != tt.long {
			t.Errorf("%d: Expected Use=%q, got %q", i, tt.long, pp.Long)
		}

		// Currently, plugins assume a Linux subsystem. Skip the execution
		// tests until this is fixed
		if runtime.GOOS != "windows" {
			if err := pp.RunE(pp, tt.args); err != nil {
				t.Errorf("Error running %s: %s", tt.use, err)
			}
			if out.String() != tt.expect {
				t.Errorf("Expected %s to output:\n%s\ngot\n%s", tt.use, tt.expect, out.String())
			}
		}
	}
}

func TestSetupEnv(t *testing.T) {
	name := "pequod"
	ph := prowpath.Home("testdata/prowhome")
	base := filepath.Join(ph.Plugins(), name)
	plugdirs := ph.Plugins()
	flagDebug = true
	defer func() {
		flagDebug = false
	}()

	setupEnv(name, base, plugdirs, ph)
	for _, tt := range []struct {
		name   string
		expect string
	}{
		{"PROW_PLUGIN_NAME", name},
		{"PROW_PLUGIN_DIR", base},
		{"PROW_PLUGIN", ph.Plugins()},
		{"PROW_DEBUG", "1"},
		{"PROW_HOME", ph.String()},
		{"PROW_PACKS_HOME", ph.Packs()},
		{"PROW_HOST", prowHost},
	} {
		if got := os.Getenv(tt.name); got != tt.expect {
			t.Errorf("Expected $%s=%q, got %q", tt.name, tt.expect, got)
		}
	}
}
