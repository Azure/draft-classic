package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/renstrom/fuzzysearch/fuzzy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/plugin"
	"github.com/Azure/draft/pkg/plugin/repository"
	"github.com/Azure/draft/pkg/plugin/repository/installer"
)

const (
	pluginHelp   = `Manage client-side Draft plugins.`
	pluginEnvVar = `DRAFT_PLUGIN`
)

func newPluginCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "add Draft plugins",
		Long:  pluginHelp,
	}
	cmd.AddCommand(
		newPluginCreateCmd(out),
		newPluginInstallCmd(out),
		newPluginListCmd(out),
		newPluginRepositoryCmd(out),
		newPluginSearchCmd(out),
		newPluginUninstallCmd(out),
		newPluginUpdateCmd(out),
		newPluginUpgradeCmd(out),
	)
	return cmd
}

func pluginDirPath(home draftpath.Home) string {
	plugdirs := os.Getenv(pluginEnvVar)

	if plugdirs == "" {
		plugdirs = home.Plugins()
	}

	return plugdirs
}

// loadPlugins loads plugins into the command list.
//
// This follows a different pattern than the other commands because it has
// to inspect its environment and then add commands to the base command
// as it finds them.
func loadPlugins(baseCmd *cobra.Command, home draftpath.Home, out io.Writer, in io.Reader) {
	plugdir := pluginDirPath(home)
	pHome := plugin.Home(plugdir)
	// Now we create commands for all of these.
	for _, plug := range findPlugins(pHome) {
		p, _, err := getPlugin(plug, pHome)
		if err != nil {
			log.Debugf("could not load plugin %s: %v", p, err)
			continue
		}
		var commandExists bool
		for _, command := range baseCmd.Commands() {
			if strings.Compare(command.Short, p.Description) == 0 {
				commandExists = true
			}
		}
		if commandExists {
			log.Debugf("command %s exists", p.Name)
			continue
		}

		c := &cobra.Command{
			Use:   p.Name,
			Short: p.Description,
			RunE: func(cmd *cobra.Command, args []string) error {

				k, u := manuallyProcessArgs(args)
				if err := cmd.Parent().ParseFlags(k); err != nil {
					return err
				}

				// Call setupEnv before PrepareCommand because
				// PrepareCommand uses os.ExpandEnv and expects the
				// setupEnv vars.
				setupPluginEnv(plug, filepath.Join(pHome.Installed(), p.Name, p.Version), plugdir, draftpath.Home(homePath()))
				main := filepath.Join(os.Getenv("DRAFT_PLUGIN_DIR"), p.GetPackage(runtime.GOOS, runtime.GOARCH).Path)

				prog := exec.Command(main, u...)
				prog.Env = os.Environ()
				prog.Stdout = out
				prog.Stderr = os.Stderr
				prog.Stdin = in
				return prog.Run()
			},
			// This passes all the flags to the subcommand.
			DisableFlagParsing: true,
		}

		if p.UseTunnel {
			c.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
				// Parse the parent flag, but not the local flags.
				k, _ := manuallyProcessArgs(args)
				if err := cmd.Parent().ParseFlags(k); err != nil {
					return err
				}
				client, config, err := getKubeClient(kubeContext)
				if err != nil {
					return fmt.Errorf("Could not get a kube client: %s", err)
				}

				tillerTunnel, err := setupTillerConnection(client, config, tillerNamespace)
				if err != nil {
					return err
				}
				tillerHost = fmt.Sprintf("127.0.0.1:%d", tillerTunnel.Local)
				return nil
			}
		}

		baseCmd.AddCommand(c)
	}
}

// manuallyProcessArgs processes an arg array, removing special args.
//
// Returns two sets of args: known and unknown (in that order)
func manuallyProcessArgs(args []string) ([]string, []string) {
	known := []string{}
	unknown := []string{}
	kvargs := []string{"--host", "--kube-context", "--home"}
	knownArg := func(a string) bool {
		for _, pre := range kvargs {
			if strings.HasPrefix(a, pre+"=") {
				return true
			}
		}
		return false
	}
	for i := 0; i < len(args); i++ {
		switch a := args[i]; a {
		case "--debug":
			known = append(known, a)
		case "--host", "--kube-context", "--home":
			known = append(known, a, args[i+1])
			i++
		default:
			if knownArg(a) {
				known = append(known, a)
				continue
			}
			unknown = append(unknown, a)
		}
	}
	return known, unknown
}

// setupPluginEnv prepares os.Env for plugins. It operates on os.Env because
// the plugin subsystem itself needs access to the environment variables
// created here.
func setupPluginEnv(shortname, base, plugdirs string, home draftpath.Home) {
	// Set extra env vars:
	for key, val := range map[string]string{
		"DRAFT_PLUGIN_NAME": shortname,
		"DRAFT_PLUGIN_DIR":  base,
		"DRAFT_BIN":         os.Args[0],

		// Set vars that may not have been set, and save client the
		// trouble of re-parsing.
		pluginEnvVar: pluginDirPath(home),
		homeEnvVar:   home.String(),
		hostEnvVar:   tillerHost,
		// Set vars that convey common information.
		"DRAFT_PACKS_HOME": home.Packs(),
	} {
		os.Setenv(key, val)
	}

	if flagDebug {
		os.Setenv("DRAFT_DEBUG", "1")
	}
}

// getPlugin renders a plugin's lua script and returns a plugin, along with the repository
// the plugin was found under. It will return an error if there was an issue rendering the
// plugin from Lua into a Go representation or by running the Lua script to generate the plugin.
func getPlugin(pluginName string, home plugin.Home) (*plugin.Plugin, string, error) {
	var (
		name string
		repo string
	)
	pluginInfo := strings.Split(pluginName, "/")
	if len(pluginInfo) == 1 {
		name = pluginInfo[0]
		repo = repository.DefaultPluginRepository
	} else {
		name = pluginInfo[len(pluginInfo)-1]
		repo = path.Dir(pluginName)
	}
	if strings.Contains(name, "./\\") {
		return nil, "", fmt.Errorf("plugin name '%s' is invalid. Plugin names cannot include the following characters: './\\'", name)
	}
	l := lua.NewState()
	defer l.Close()
	if err := l.DoFile(filepath.Join(home.Repositories(), repo, "Plugins", fmt.Sprintf("%s.lua", name))); err != nil {
		return nil, "", err
	}
	var plugin plugin.Plugin
	if err := gluamapper.Map(l.GetGlobal(strings.ToLower("plugin")).(*lua.LTable), &plugin); err != nil {
		return nil, "", err
	}
	return &plugin, repo, nil
}

// findPlugins returns a list of installed plugins.
func findPlugins(home plugin.Home) []string {
	var plugins []string
	files, err := ioutil.ReadDir(home.Installed())
	if err != nil {
		return []string{}
	}

	for _, f := range files {
		if f.IsDir() {
			files, err := ioutil.ReadDir(filepath.Join(home.Installed(), f.Name()))
			if err != nil {
				continue
			}
			if len(files) > 0 {
				plugins = append(plugins, f.Name())
			}
		}
	}
	return plugins
}

// findPluginVersions returns a list of all installed versions of a given plugin.
func findPluginVersions(name string, home plugin.Home) []string {
	var versions []string
	files, err := ioutil.ReadDir(filepath.Join(home.Installed(), name))
	if err != nil {
		return []string{}
	}

	for _, f := range files {
		if f.IsDir() {
			versions = append(versions, f.Name())
		}
	}
	return versions
}

func updatePluginRepositories(home plugin.Home) error {
	repositories := findRepositories(home.Repositories())
	for _, repository := range repositories {
		i, err := installer.FindSource(filepath.Join(home.Repositories(), repository), home)
		if err != nil {
			return err
		}
		if err := installer.Update(i); err != nil {
			return err
		}
	}
	return nil
}

func search(keywords []string, home plugin.Home) []string {
	var pluginNames = findPluginMetadata(home)
	var foundPlugins = make(map[string]bool)
	// if no keywords are given, display all available plugins
	if len(keywords) == 0 {
		for _, found := range pluginNames {
			foundPlugins[found] = true
		}
	} else {
		for _, keyword := range keywords {
			for _, found := range fuzzy.Find(keyword, pluginNames) {
				foundPlugins[found] = true
			}
		}
	}
	names := []string{}
	for n := range foundPlugins {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func findPluginMetadata(home plugin.Home) []string {
	repoPath := home.Repositories()
	var plugins []string
	filepath.Walk(repoPath, func(p string, f os.FileInfo, err error) error {
		if err != nil {
			log.Errorln(err)
			return nil
		}
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".lua") {
			pluginName := strings.TrimSuffix(f.Name(), ".lua")
			repoName := strings.TrimPrefix(p, repoPath+string(os.PathSeparator))
			repoName = strings.TrimSuffix(repoName, string(os.PathSeparator)+filepath.Join("Plugins", f.Name()))
			// for Windows clients, we need to replace the path separator with forward slashes
			repoName = strings.Replace(repoName, "\\", "/", -1)
			name := pluginName
			// if the plugin comes from a non-default repository, we tack on the repo name
			if repoName != repository.DefaultPluginRepository {
				name = path.Join(repoName, pluginName)
			}
			plugins = append(plugins, name)
		}
		return nil
	})
	sort.Strings(plugins)
	return plugins
}
