// +build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/plugin"
)

// runTask will execute a plugin task.
func runTask(p *plugin.Plugin, event string) error {
	tasks, ok := p.Metadata.PlatformTasks["bash"]
	if !ok {
		return nil
	}
	task := tasks.Get(event)
	if task == "" {
		return nil
	}

	prog := exec.Command("sh", "-c", task)

	debug("running %s task: %s %v", event, prog.Path, prog.Args)

	home := draftpath.Home(homePath())
	setupPluginEnv(p.Metadata.Name, p.Metadata.Version, p.Dir, home.Plugins(), home)
	prog.Stdout, prog.Stderr = os.Stdout, os.Stderr
	if err := prog.Run(); err != nil {
		if eerr, ok := err.(*exec.ExitError); ok {
			os.Stderr.Write(eerr.Stderr)
			return fmt.Errorf("plugin %s task for %q exited with error", event, p.Metadata.Name)
		}
		return err
	}
	return nil
}
