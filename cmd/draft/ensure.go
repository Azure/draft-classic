package main

import (
	"fmt"

	"github.com/Azure/draft/pkg/plugin/repository"

	"github.com/Azure/draft/pkg/osutil"

	"github.com/Azure/draft/pkg/draft/pack/repo"
	"github.com/Azure/draft/pkg/plugin"
)

// ensureDirectories checks to see if $DRAFT_HOME exists
//
// If $DRAFT_HOME does not exist, this function will create it.
func (i *initCmd) ensureDirectories() error {
	pHome := plugin.Home(i.home.Plugins())
	configDirectories := []string{
		i.home.String(),
		i.home.Plugins(),
		i.home.Packs(),
		i.home.Logs(),
		pHome.Cache(),
		pHome.Installed(),
		pHome.Repositories(),
	}
	for _, p := range configDirectories {
		err := osutil.EnsureDirectory(p)
		if err != nil {
			return err
		}
	}

	return nil
}

// ensureConfig checks to see if $DRAFT_HOME/config.toml exists
//
// If it does not exist, this function will create it.
func (i *initCmd) ensureConfig() error {
	return osutil.EnsureFile(i.home.Config())
}

// ensurePacks checks to see if the default packs exist.
//
// If the pack does not exist, this function will create it.
func (i *initCmd) ensurePacks(repos []repo.Builtin) error {
	existingRepos := repo.FindRepositories(i.home.Packs())

	fmt.Fprintln(i.out, "Installing default pack repositories...")
	for _, builtin := range repo.Builtins() {
		if err := i.ensurePack(builtin, existingRepos); err != nil {
			return err
		}
	}
	fmt.Fprintln(i.out, "Installation of default pack repositories complete")

	if len(repos) > 0 {
		existingRepos := repo.FindRepositories(i.home.Packs())
		fmt.Fprintln(i.out, "Installing packs from config file")

		for _, r := range repos {
			if err := i.ensurePack(&r, existingRepos); err != nil {
				fmt.Println(err)
				return err
			}
		}

		fmt.Fprintln(i.out, "Installation of packs from config file complete")

	}
	return nil
}

func (i *initCmd) ensurePack(builtin *repo.Builtin, existingRepos []repo.Repository) error {

	for _, repo := range existingRepos {
		if builtin.Name == repo.Name {
			return nil
		}
	}

	addArgs := []string{
		"add",
		builtin.URL,
	}

	addFlags := []string{
		"--version",
		builtin.Version,
		"--home",
		string(i.home),
		fmt.Sprintf("--debug=%v", flagDebug),
	}

	packRepoCmd, _, err := rootCmd.Find([]string{"pack-repo"})
	if err != nil {
		return err
	}

	if err := packRepoCmd.ParseFlags(addFlags); err != nil {
		return err
	}

	if err := packRepoCmd.RunE(packRepoCmd, addArgs); err != nil {
		return err
	}
	debug("Successfully installed pack repo: %v %v", builtin.URL, builtin.Version)
	return nil
}

// ensurePluginRepositories checks to see if the default plugin repositories exist.
//
// If the plugin repo does not exist, this function will add it.
func (i *initCmd) ensurePluginRepositories(pluginRepos []repository.Builtin) error {
	fmt.Fprintln(i.out, "Installing default plugin repositories...")
	for _, builtin := range repository.Builtins() {
		if err := i.ensurePluginRepository(string(builtin)); err != nil {
			return err
		}
	}
	fmt.Fprintln(i.out, "Installation of default plugin repositories complete")

	if len(pluginRepos) > 0 {
		fmt.Fprintln(i.out, "Installing plugin repositories from config file")

		for _, plugRepo := range pluginRepos {
			if err := i.ensurePluginRepository(string(plugRepo)); err != nil {
				fmt.Println(err)
				return err
			}
		}

		fmt.Fprintln(i.out, "Installation of plugin repositories from config file complete")

	}

	return nil
}

// ensurePlugins checks to see if the default plugins exist.
//
// If the plugin does not exist, this function will add it.
func (i *initCmd) ensurePlugins(plugins []plugin.Builtin) error {
	if err := i.updatePluginRepos(); err != nil {
		return err
	}

	fmt.Fprintln(i.out, "Installing default plugins...")
	for _, builtin := range plugin.Builtins() {
		if err := i.ensurePlugin(string(builtin)); err != nil {
			return err
		}
	}
	fmt.Fprintln(i.out, "Installation of default plugins complete")

	if len(plugins) > 0 {
		fmt.Fprintln(i.out, "Installing plugins from config file")

		for _, plug := range plugins {
			if err := i.ensurePlugin(string(plug)); err != nil {
				fmt.Println(err)
				return err
			}
		}

		fmt.Fprintln(i.out, "Installation of plugins from config file complete")

	}

	return nil
}

func (i *initCmd) updatePluginRepos() error {
	flags := []string{
		"--home",
		string(i.home),
		fmt.Sprintf("--debug=%v", flagDebug),
	}

	plugUpdateCmd, _, err := rootCmd.Find([]string{"plugin", "update"})
	if err != nil {
		return err
	}

	if err := plugUpdateCmd.ParseFlags(flags); err != nil {
		return err
	}

	if err := plugUpdateCmd.PreRunE(plugUpdateCmd, []string{}); err != nil {
		return err
	}

	return plugUpdateCmd.RunE(plugUpdateCmd, []string{})
}

func (i *initCmd) ensurePlugin(builtin string) error {
	installArgs := []string{
		builtin,
	}

	installFlags := []string{
		"--home",
		string(i.home),
		fmt.Sprintf("--debug=%v", flagDebug),
	}

	plugInstallCmd, _, err := rootCmd.Find([]string{"plugin", "install"})
	if err != nil {
		return err
	}

	if err := plugInstallCmd.ParseFlags(installFlags); err != nil {
		return err
	}

	if err := plugInstallCmd.PreRunE(plugInstallCmd, installArgs); err != nil {
		return err
	}

	if err := plugInstallCmd.RunE(plugInstallCmd, installArgs); err != nil {
		return err
	}

	// reload plugins
	loadPlugins(rootCmd, i.home, i.out, i.in)

	debug("Successfully installed %v", builtin)
	return nil
}

func (i *initCmd) ensurePluginRepository(builtin string) error {
	installArgs := []string{
		builtin,
	}

	installFlags := []string{
		"--home",
		string(i.home),
		fmt.Sprintf("--debug=%v", flagDebug),
	}

	plugRepoAddCmd, _, err := rootCmd.Find([]string{"plugin", "repository", "add"})
	if err != nil {
		return err
	}

	if err := plugRepoAddCmd.ParseFlags(installFlags); err != nil {
		return err
	}

	if err := plugRepoAddCmd.RunE(plugRepoAddCmd, installArgs); err != nil {
		return err
	}

	// reload plugins
	loadPlugins(rootCmd, i.home, i.out, i.in)

	debug("Successfully installed %v", builtin)
	return nil
}
