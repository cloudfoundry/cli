package common

import (
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/api/plugin"
	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/plugin/shared"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util"
	"code.cloudfoundry.org/cli/util/configv3"
	log "github.com/sirupsen/logrus"
)

//go:generate counterfeiter . InstallPluginActor

type InstallPluginActor interface {
	CreateExecutableCopy(path string, tempPluginDir string) (string, error)
	DownloadExecutableBinaryFromURL(url string, tempPluginDir string, proxyReader plugin.ProxyReader) (string, error)
	FileExists(path string) bool
	GetAndValidatePlugin(metadata pluginaction.PluginMetadata, commands pluginaction.CommandList, path string) (configv3.Plugin, error)
	GetPlatformString(runtimeGOOS string, runtimeGOARCH string) string
	GetPluginInfoFromRepositoriesForPlatform(pluginName string, pluginRepos []configv3.PluginRepository, platform string) (pluginaction.PluginInfo, []string, error)
	GetPluginRepository(repositoryName string) (configv3.PluginRepository, error)
	InstallPluginFromPath(path string, plugin configv3.Plugin) error
	UninstallPlugin(uninstaller pluginaction.PluginUninstaller, name string) error
	ValidateFileChecksum(path string, checksum string) bool
}

const installConfirmationPrompt = "Do you want to install the plugin {{.Path}}?"

type cancelInstall struct {
}

func (cancelInstall) Error() string {
	return "Nobody should see this error. If you do, report it!"
}

type PluginSource int

const (
	PluginFromRepository PluginSource = iota
	PluginFromLocalFile
	PluginFromURL
)

type InstallPluginCommand struct {
	OptionalArgs         flag.InstallPluginArgs `positional-args:"yes"`
	SkipSSLValidation    bool                   `short:"k" hidden:"true" description:"Skip SSL certificate validation"`
	Force                bool                   `short:"f" description:"Force install of plugin without confirmation"`
	RegisteredRepository string                 `short:"r" description:"Restrict search for plugin to this registered repository"`
	usage                interface{}            `usage:"CF_NAME install-plugin PLUGIN_NAME [-r REPO_NAME] [-f]\nCF_NAME install-plugin LOCAL-PATH/TO/PLUGIN | URL [-f]\n\nWARNING:\n   Plugins are binaries written by potentially untrusted authors.\n   Install and use plugins at your own risk.\n\nEXAMPLES:\n   CF_NAME install-plugin ~/Downloads/plugin-foobar\n   CF_NAME install-plugin https://example.com/plugin-foobar_linux_amd64\n   CF_NAME install-plugin -r My-Repo plugin-echo"`
	relatedCommands      interface{}            `related_commands:"add-plugin-repo, list-plugin-repos, plugins"`
	UI                   command.UI
	Config               command.Config
	Actor                InstallPluginActor
	ProgressBar          plugin.ProxyReader
}

func (cmd *InstallPluginCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.Actor = pluginaction.NewActor(config, shared.NewClient(config, ui, cmd.SkipSSLValidation))

	cmd.ProgressBar = shared.NewProgressBarProxyReader(cmd.UI.Writer())

	return nil
}

func (cmd InstallPluginCommand) Execute([]string) error {
	log.WithField("PluginHome", cmd.Config.PluginHome()).Info("making plugin dir")
	err := os.MkdirAll(cmd.Config.PluginHome(), 0700)
	if err != nil {
		return err
	}

	tempPluginDir, err := ioutil.TempDir(cmd.Config.PluginHome(), "temp")
	log.WithField("tempPluginDir", tempPluginDir).Debug("making tempPluginDir dir")
	defer os.RemoveAll(tempPluginDir)

	if err != nil {
		return err
	}

	tempPluginPath, pluginSource, err := cmd.getPluginBinaryAndSource(tempPluginDir)
	if _, ok := err.(cancelInstall); ok {
		cmd.UI.DisplayText("Plugin installation cancelled.")
		return nil
	} else if err != nil {
		return err
	}
	log.WithFields(log.Fields{"tempPluginPath": tempPluginPath, "pluginSource": pluginSource}).Debug("getPluginBinaryAndSource")

	// copy twice when downloading from a URL to keep Windows specific code
	// isolated to CreateExecutableCopy
	executablePath, err := cmd.Actor.CreateExecutableCopy(tempPluginPath, tempPluginDir)
	if err != nil {
		return err
	}
	log.WithField("executablePath", executablePath).Debug("created executable copy")

	rpcService, err := shared.NewRPCService(cmd.Config, cmd.UI)
	if err != nil {
		return err
	}
	log.Info("started RPC server")

	plugin, err := cmd.Actor.GetAndValidatePlugin(rpcService, Commands, executablePath)
	if err != nil {
		return err
	}
	log.Info("validated plugin")

	if installedPlugin, installed := cmd.Config.GetPluginCaseInsensitive(plugin.Name); installed {
		log.WithField("version", installedPlugin.Version).Debug("uninstall plugin")

		if !cmd.Force && pluginSource != PluginFromRepository {
			return translatableerror.PluginAlreadyInstalledError{
				BinaryName: cmd.Config.BinaryName(),
				Name:       plugin.Name,
				Version:    plugin.Version.String(),
			}
		}

		err = cmd.uninstallPlugin(installedPlugin, rpcService)
		if err != nil {
			return err
		}
	}

	log.Info("install plugin")
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
		log.WithField("RegisteredRepository", cmd.RegisteredRepository).Info("installing from specified repository")
		pluginRepository, err := cmd.Actor.GetPluginRepository(cmd.RegisteredRepository)
		if err != nil {
			return "", 0, err
		}
		path, pluginSource, err := cmd.getPluginFromRepositories(pluginNameOrLocation, []configv3.PluginRepository{pluginRepository}, tempPluginDir)

		if err != nil {
			switch pluginErr := err.(type) {
			case actionerror.PluginNotFoundInAnyRepositoryError:
				return "", 0, translatableerror.PluginNotFoundInRepositoryError{
					BinaryName:     cmd.Config.BinaryName(),
					PluginName:     pluginNameOrLocation,
					RepositoryName: cmd.RegisteredRepository,
				}

			case actionerror.FetchingPluginInfoFromRepositoryError:
				// The error wrapped inside pluginErr is handled differently in the case of
				// a specified repo from that of searching through all repos.  pluginErr.Err
				// is then processed by shared.HandleError by this function's caller.
				return "", 0, pluginErr.Err

			default:
				return "", 0, err
			}
		}
		return path, pluginSource, nil

	case cmd.Actor.FileExists(pluginNameOrLocation):
		log.WithField("pluginNameOrLocation", pluginNameOrLocation).Info("installing from specified file")
		return cmd.getPluginFromLocalFile(pluginNameOrLocation)

	case util.IsHTTPScheme(pluginNameOrLocation):
		log.WithField("pluginNameOrLocation", pluginNameOrLocation).Info("installing from specified URL")
		return cmd.getPluginFromURL(pluginNameOrLocation, tempPluginDir)

	case util.IsUnsupportedURLScheme(pluginNameOrLocation):
		log.WithField("pluginNameOrLocation", pluginNameOrLocation).Error("Unsupported URL")
		return "", 0, translatableerror.UnsupportedURLSchemeError{UnsupportedURL: pluginNameOrLocation}

	default:
		log.Info("installing from first repository with plugin")
		repos := cmd.Config.PluginRepositories()
		if len(repos) == 0 {
			return "", 0, translatableerror.PluginNotFoundOnDiskOrInAnyRepositoryError{PluginName: pluginNameOrLocation, BinaryName: cmd.Config.BinaryName()}
		}

		path, pluginSource, err := cmd.getPluginFromRepositories(pluginNameOrLocation, repos, tempPluginDir)
		if err != nil {
			switch pluginErr := err.(type) {
			case actionerror.PluginNotFoundInAnyRepositoryError:
				return "", 0, translatableerror.PluginNotFoundOnDiskOrInAnyRepositoryError{PluginName: pluginNameOrLocation, BinaryName: cmd.Config.BinaryName()}

			case actionerror.FetchingPluginInfoFromRepositoryError:
				return "", 0, cmd.handleFetchingPluginInfoFromRepositoriesError(pluginErr)

			default:
				return "", 0, err
			}
		}
		return path, pluginSource, nil
	}
}

// These are specific errors that we output to the user in the context of
// installing from any repository.
func (InstallPluginCommand) handleFetchingPluginInfoFromRepositoriesError(fetchErr actionerror.FetchingPluginInfoFromRepositoryError) error {
	switch clientErr := fetchErr.Err.(type) {
	case pluginerror.RawHTTPStatusError:
		return translatableerror.FetchingPluginInfoFromRepositoriesError{
			Message:        clientErr.Status,
			RepositoryName: fetchErr.RepositoryName,
		}

	case pluginerror.SSLValidationHostnameError:
		return translatableerror.FetchingPluginInfoFromRepositoriesError{
			Message:        clientErr.Error(),
			RepositoryName: fetchErr.RepositoryName,
		}

	case pluginerror.UnverifiedServerError:
		return translatableerror.FetchingPluginInfoFromRepositoriesError{
			Message:        clientErr.Error(),
			RepositoryName: fetchErr.RepositoryName,
		}

	default:
		return clientErr
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

	tempPath, err := cmd.Actor.DownloadExecutableBinaryFromURL(pluginLocation, tempPluginDir, cmd.ProgressBar)
	if err != nil {
		return "", 0, err
	}

	return tempPath, PluginFromURL, err
}

func (cmd InstallPluginCommand) getPluginFromRepositories(pluginName string, repos []configv3.PluginRepository, tempPluginDir string) (string, PluginSource, error) {
	var repoNames []string
	for _, repo := range repos {
		repoNames = append(repoNames, repo.Name)
	}

	cmd.UI.DisplayTextWithFlavor("Searching {{.RepositoryName}} for plugin {{.PluginName}}...", map[string]interface{}{
		"RepositoryName": strings.Join(repoNames, ", "),
		"PluginName":     pluginName,
	})

	currentPlatform := cmd.Actor.GetPlatformString(runtime.GOOS, runtime.GOARCH)
	pluginInfo, repoList, err := cmd.Actor.GetPluginInfoFromRepositoriesForPlatform(pluginName, repos, currentPlatform)
	if err != nil {
		return "", 0, err
	}

	cmd.UI.DisplayText("Plugin {{.PluginName}} {{.PluginVersion}} found in: {{.RepositoryName}}", map[string]interface{}{
		"PluginName":     pluginName,
		"PluginVersion":  pluginInfo.Version,
		"RepositoryName": strings.Join(repoList, ", "),
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
		"RepositoryName": repoList[0],
	})

	tempPath, err := cmd.Actor.DownloadExecutableBinaryFromURL(pluginInfo.URL, tempPluginDir, cmd.ProgressBar)
	if err != nil {
		return "", 0, err
	}

	if !cmd.Actor.ValidateFileChecksum(tempPath, pluginInfo.Checksum) {
		return "", 0, translatableerror.InvalidChecksumError{}
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
		log.Debug("plugin confirmation - 'no' inputed")
		return cancelInstall{}
	}

	return nil
}
