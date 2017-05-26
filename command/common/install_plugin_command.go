package common

import (
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/actor/pluginaction"
	oldCmd "code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/plugin/shared"
	"code.cloudfoundry.org/cli/util"
	"code.cloudfoundry.org/cli/util/configv3"
)

//go:generate counterfeiter . InstallPluginActor

type InstallPluginActor interface {
	CreateExecutableCopy(path string, tempPluginDir string) (string, error)
	DownloadExecutableBinaryFromURL(url string, tempPluginDir string) (string, int64, error)
	FileExists(path string) bool
	GetAndValidatePlugin(metadata pluginaction.PluginMetadata, commands pluginaction.CommandList, path string) (configv3.Plugin, error)
	GetPluginInfoFromRepository(pluginName string, pluginRepo configv3.PluginRepository) (pluginaction.PluginInfo, error)
	GetPluginRepository(repositoryName string) (configv3.PluginRepository, error)
	InstallPluginFromPath(path string, plugin configv3.Plugin) error
	IsPluginInstalled(pluginName string) bool
	UninstallPlugin(uninstaller pluginaction.PluginUninstaller, name string) error
	ValidateFileChecksum(path string, checksum string) bool
}

const installConfirmationPrompt = "Do you want to install the plugin {{.Path}}?"

type InvalidChecksumError struct {
}

func (_ InvalidChecksumError) Error() string {
	return "Downloaded plugin binary's checksum does not match repo metadata.\nPlease try again or contact the plugin author."
}

type PluginSource int

const (
	PluginFromRepository PluginSource = iota
	PluginFromLocalFile
	PluginFromURL
)

type InstallPluginCommand struct {
	OptionalArgs         flag.InstallPluginArgs `positional-args:"yes"`
	Force                bool                   `short:"f" description:"Force install of plugin without confirmation"`
	SkipSSLValidation    bool                   `short:"k" hidden:"true" description:"Skip SSL certificate validation"`
	RegisteredRepository string                 `short:"r" description:"Name of a registered repository where the specified plugin is located"`
	usage                interface{}            `usage:"CF_NAME install-plugin (LOCAL-PATH/TO/PLUGIN | URL | -r REPO_NAME PLUGIN_NAME) [-f]\n\nEXAMPLES:\n   CF_NAME install-plugin ~/Downloads/plugin-foobar\n   CF_NAME install-plugin https://example.com/plugin-foobar_linux_amd64\n   CF_NAME install-plugin -r My-Repo plugin-echo"`
	relatedCommands      interface{}            `related_commands:"add-plugin-repo, list-plugin-repos, plugins"`
	UI                   command.UI
	Config               command.Config
	Actor                InstallPluginActor
}

func (cmd *InstallPluginCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.Actor = pluginaction.NewActor(config, shared.NewClient(config, ui, cmd.SkipSSLValidation))
	return nil
}

func (cmd InstallPluginCommand) Execute(_ []string) error {
	if !cmd.Config.Experimental() {
		oldCmd.Main(os.Getenv("CF_TRACE"), os.Args)
		return nil
	}

	err := os.MkdirAll(cmd.Config.PluginHome(), 0700)
	if err != nil {
		return shared.HandleError(err)
	}

	tempPluginDir, err := ioutil.TempDir(cmd.Config.PluginHome(), "temp")
	defer os.RemoveAll(tempPluginDir)

	if err != nil {
		return shared.HandleError(err)
	}

	tempPluginPath, pluginSource, err := cmd.getPluginBinaryAndSource(tempPluginDir)
	if err != nil {
		return shared.HandleError(err)
	}

	// copy twice when downloading from a URL to keep Windows specific code
	// isolated to CreateExecutableCopy
	executablePath, err := cmd.Actor.CreateExecutableCopy(tempPluginPath, tempPluginDir)
	if err != nil {
		return shared.HandleError(err)
	}

	rpcService, err := shared.NewRPCService(cmd.Config, cmd.UI)
	if err != nil {
		return shared.HandleError(err)
	}

	plugin, err := cmd.Actor.GetAndValidatePlugin(rpcService, Commands, executablePath)
	if err != nil {
		return shared.HandleError(err)
	}

	if cmd.Actor.IsPluginInstalled(plugin.Name) {
		if !cmd.Force && pluginSource != PluginFromRepository {
			return shared.PluginAlreadyInstalledError{
				BinaryName: cmd.Config.BinaryName(),
				Name:       plugin.Name,
				Version:    plugin.Version.String(),
			}
		}

		err = cmd.uninstallPlugin(plugin, rpcService)
		if err != nil {
			return shared.HandleError(err)
		}
	}

	return cmd.installPlugin(plugin, executablePath)
}

func (cmd InstallPluginCommand) installPlugin(plugin configv3.Plugin, pluginPath string) error {
	cmd.UI.DisplayTextWithFlavor("Installing plugin {{.Name}}...", map[string]interface{}{
		"Name": plugin.Name,
	})

	installErr := cmd.Actor.InstallPluginFromPath(pluginPath, plugin)
	if installErr != nil {
		return installErr
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("Plugin {{.Name}} {{.Version}} successfully installed.", map[string]interface{}{
		"Name":    plugin.Name,
		"Version": plugin.Version.String(),
	})
	return nil
}

func (cmd InstallPluginCommand) uninstallPlugin(plugin configv3.Plugin, rpcService *shared.RPCService) error {
	cmd.UI.DisplayText("Plugin {{.Name}} {{.Version}} is already installed. Uninstalling existing plugin...", map[string]interface{}{
		"Name":    plugin.Name,
		"Version": plugin.Version.String(),
	})

	uninstallErr := cmd.Actor.UninstallPlugin(rpcService, plugin.Name)
	if uninstallErr != nil {
		return uninstallErr
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("Plugin {{.Name}} successfully uninstalled.", map[string]interface{}{
		"Name": plugin.Name,
	})

	return nil
}

func (cmd InstallPluginCommand) getPluginBinaryAndSource(tempPluginDir string) (string, PluginSource, error) {
	pluginNameOrLocation := cmd.OptionalArgs.PluginNameOrLocation.String()

	switch {
	case cmd.RegisteredRepository != "":
		return cmd.getPluginFromRepository(pluginNameOrLocation, tempPluginDir)

	case cmd.Actor.FileExists(pluginNameOrLocation):
		return cmd.getPluginFromLocalFile(pluginNameOrLocation)

	case util.IsHTTPScheme(pluginNameOrLocation):
		return cmd.getPluginFromURL(pluginNameOrLocation, tempPluginDir)

	case util.IsUnsupportedURLScheme(pluginNameOrLocation):
		return "", 0, command.UnsupportedURLSchemeError{UnsupportedURL: pluginNameOrLocation}

	default:
		return "", 0, shared.FileNotFoundError{Path: pluginNameOrLocation}
	}
}

func (cmd InstallPluginCommand) getPluginFromLocalFile(pluginLocation string) (string, PluginSource, error) {
	err := cmd.installPluginPrompt(installConfirmationPrompt, map[string]interface{}{
		"Path": pluginLocation,
	})
	if err != nil {
		return "", 0, err
	}

	return pluginLocation, PluginFromLocalFile, err
}

func (cmd InstallPluginCommand) getPluginFromURL(pluginLocation string, tempPluginDir string) (string, PluginSource, error) {
	err := cmd.installPluginPrompt(installConfirmationPrompt, map[string]interface{}{
		"Path": pluginLocation,
	})
	if err != nil {
		return "", 0, err
	}

	cmd.UI.DisplayText("Starting download of plugin binary from URL...")

	var size int64
	tempPath, size, err := cmd.Actor.DownloadExecutableBinaryFromURL(pluginLocation, tempPluginDir)
	if err != nil {
		return "", 0, err
	}

	cmd.UI.DisplayText("{{.Bytes}} bytes downloaded...", map[string]interface{}{
		"Bytes": size,
	})

	return tempPath, PluginFromURL, err
}

func (cmd InstallPluginCommand) getPluginFromRepository(pluginName string, tempPluginDir string) (string, PluginSource, error) {
	var (
		pluginRepository configv3.PluginRepository
		pluginInfo       pluginaction.PluginInfo
		err              error
		tempPath         string
	)

	pluginRepository, err = cmd.Actor.GetPluginRepository(cmd.RegisteredRepository)
	if err != nil {
		return "", 0, err
	}

	cmd.UI.DisplayTextWithFlavor("Searching {{.RepositoryName}} for plugin {{.PluginName}}...", map[string]interface{}{
		"RepositoryName": cmd.RegisteredRepository,
		"PluginName":     pluginName,
	})

	pluginInfo, err = cmd.Actor.GetPluginInfoFromRepository(pluginName, pluginRepository)
	if err != nil {
		if _, ok := err.(pluginaction.PluginNotFoundInRepositoryError); ok {
			return "", 0, shared.PluginNotFoundInRepositoryError{
				BinaryName:     cmd.Config.BinaryName(),
				PluginName:     pluginName,
				RepositoryName: cmd.RegisteredRepository,
			}
		}
		return "", 0, err
	}
	cmd.UI.DisplayText("Plugin {{.PluginName}} {{.PluginVersion}} found in: {{.RepositoryName}}", map[string]interface{}{
		"PluginName":     pluginName,
		"PluginVersion":  pluginInfo.Version,
		"RepositoryName": cmd.RegisteredRepository,
	})

	installedPlugin, exist := cmd.Config.GetPlugin(pluginName)
	if exist {
		cmd.UI.DisplayText("Plugin {{.PluginName}} {{.PluginVersion}} is already installed.", map[string]interface{}{
			"PluginName":    installedPlugin.Name,
			"PluginVersion": installedPlugin.Version.String(),
		})

		err = cmd.installPluginPrompt("Do you want to uninstall the existing plugin and install {{.Path}} {{.PluginVersion}}?", map[string]interface{}{
			"Path":          pluginName,
			"PluginVersion": pluginInfo.Version,
		})
	} else {
		err = cmd.installPluginPrompt(installConfirmationPrompt, map[string]interface{}{
			"Path": pluginName,
		})
	}

	if err != nil {
		return "", 0, err
	}

	cmd.UI.DisplayText("Starting download of plugin binary from repository {{.RepositoryName}}...", map[string]interface{}{
		"RepositoryName": cmd.RegisteredRepository,
	})

	var size int64
	tempPath, size, err = cmd.Actor.DownloadExecutableBinaryFromURL(pluginInfo.URL, tempPluginDir)
	if err != nil {
		return "", 0, err
	}

	cmd.UI.DisplayText("{{.Bytes}} bytes downloaded...", map[string]interface{}{
		"Bytes": size,
	})

	if !cmd.Actor.ValidateFileChecksum(tempPath, pluginInfo.Checksum) {
		return "", 0, InvalidChecksumError{}
	}

	return tempPath, PluginFromRepository, err
}

func (cmd InstallPluginCommand) installPluginPrompt(template string, templateValues ...map[string]interface{}) error {
	cmd.UI.DisplayHeader("Attention: Plugins are binaries written by potentially untrusted authors.")
	cmd.UI.DisplayHeader("Install and use plugins at your own risk.")

	if cmd.Force {
		return nil
	}

	var (
		really    bool
		promptErr error
	)

	really, promptErr = cmd.UI.DisplayBoolPrompt(false, template, templateValues...)

	if promptErr != nil {
		return promptErr
	}

	if !really {
		cmd.UI.DisplayText("Plugin installation cancelled.")
		return shared.PluginInstallationCancelled{}
	}

	return nil
}
