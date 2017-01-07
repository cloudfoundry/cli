package plugin

import (
	"errors"
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/cli/cf/actors/plugininstaller"
	"code.cloudfoundry.org/cli/cf/actors/pluginrepo"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/configuration/pluginconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/util"
	"code.cloudfoundry.org/cli/util/downloader"
	"code.cloudfoundry.org/gofileutils/fileutils"

	pluginRPCService "code.cloudfoundry.org/cli/plugin/rpc"
)

type PluginInstall struct {
	ui           terminal.UI
	config       coreconfig.Reader
	pluginConfig pluginconfig.PluginConfiguration
	pluginRepo   pluginrepo.PluginRepo
	checksum     util.Sha1Checksum
	rpcService   *pluginRPCService.CliRpcService
}

func init() {
	commandregistry.Register(&PluginInstall{})
}

func (cmd *PluginInstall) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["r"] = &flags.StringFlag{ShortName: "r", Usage: T("Name of a registered repository where the specified plugin is located")}
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force install of plugin without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "install-plugin",
		Description: T("Install CLI plugin"),
		Usage: []string{
			T(`CF_NAME install-plugin (LOCAL-PATH/TO/PLUGIN | URL | -r REPO_NAME PLUGIN_NAME) [-f]

   Prompts for confirmation unless '-f' is provided.`),
		},
		Examples: []string{
			"CF_NAME install-plugin ~/Downloads/plugin-foobar",
			"CF_NAME install-plugin https://example.com/plugin-foobar_linux_amd64",
			"CF_NAME install-plugin -r My-Repo plugin-echo",
		},
		Flags:     fs,
		TotalArgs: 1,
	}
}

func (cmd *PluginInstall) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("install-plugin"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	reqs := []requirements.Requirement{}
	return reqs, nil
}

func (cmd *PluginInstall) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.pluginConfig = deps.PluginConfig
	cmd.pluginRepo = deps.PluginRepo
	cmd.checksum = deps.ChecksumUtil

	//reset rpc registration in case there is other running instance,
	//each service can only be registered once
	server := rpc.NewServer()

	rpcService, err := pluginRPCService.NewRpcService(deps.TeePrinter, deps.TeePrinter, deps.Config, deps.RepoLocator, pluginRPCService.NewCommandRunner(), deps.Logger, cmd.ui.Writer(), server)
	if err != nil {
		cmd.ui.Failed("Error initializing RPC service: " + err.Error())
	}

	cmd.rpcService = rpcService

	return cmd
}

func (cmd *PluginInstall) Execute(c flags.FlagContext) error {
	if !cmd.confirmWithUser(
		c,
		T("**Attention: Plugins are binaries written by potentially untrusted authors. Install and use plugins at your own risk.**\n\nDo you want to install the plugin {{.Plugin}}?",
			map[string]interface{}{
				"Plugin": c.Args()[0],
			}),
	) {
		return errors.New(T("Plugin installation cancelled"))
	}

	fileDownloader := downloader.NewDownloader(os.TempDir())

	removeTmpFile := func() {
		err := fileDownloader.RemoveFile()
		if err != nil {
			cmd.ui.Say(T("Problem removing downloaded binary in temp directory: ") + err.Error())
		}
	}
	defer removeTmpFile()

	deps := &plugininstaller.Context{
		Checksummer:    cmd.checksum,
		GetPluginRepos: cmd.config.PluginRepos,
		FileDownloader: fileDownloader,
		PluginRepo:     cmd.pluginRepo,
		RepoName:       c.String("r"),
		UI:             cmd.ui,
	}
	installer := plugininstaller.NewPluginInstaller(deps)
	pluginSourceFilepath := installer.Install(c.Args()[0])

	_, pluginExecutableName := filepath.Split(pluginSourceFilepath)

	cmd.ui.Say(T(
		"Installing plugin {{.PluginPath}}...",
		map[string]interface{}{
			"PluginPath": pluginExecutableName,
		}),
	)

	pluginDestinationFilepath := filepath.Join(cmd.pluginConfig.GetPluginPath(), pluginExecutableName)

	err := cmd.ensurePluginBinaryWithSameFileNameDoesNotAlreadyExist(pluginDestinationFilepath, pluginExecutableName)
	if err != nil {
		return err
	}

	pluginMetadata, err := cmd.runBinaryAndObtainPluginMetadata(pluginSourceFilepath)
	if err != nil {
		return err
	}

	err = cmd.ensurePluginIsSafeForInstallation(pluginMetadata, pluginDestinationFilepath, pluginSourceFilepath)
	if err != nil {
		return err
	}

	err = cmd.installPlugin(pluginMetadata, pluginDestinationFilepath, pluginSourceFilepath)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say(T(
		"Plugin {{.PluginName}} v{{.Version}} successfully installed.",
		map[string]interface{}{
			"PluginName": pluginMetadata.Name,
			"Version":    fmt.Sprintf("%d.%d.%d", pluginMetadata.Version.Major, pluginMetadata.Version.Minor, pluginMetadata.Version.Build),
		}),
	)
	return nil
}

func (cmd *PluginInstall) confirmWithUser(c flags.FlagContext, prompt string) bool {
	return c.Bool("f") || cmd.ui.Confirm(prompt)
}

func (cmd *PluginInstall) ensurePluginBinaryWithSameFileNameDoesNotAlreadyExist(pluginDestinationFilepath, pluginExecutableName string) error {
	_, err := os.Stat(pluginDestinationFilepath)
	if err == nil || os.IsExist(err) {
		return errors.New(T(
			"The file {{.PluginExecutableName}} already exists under the plugin directory.\n",
			map[string]interface{}{
				"PluginExecutableName": pluginExecutableName,
			}),
		)
	} else if !os.IsNotExist(err) {
		return errors.New(T(
			"Unexpected error has occurred:\n{{.Error}}",
			map[string]interface{}{
				"Error": err.Error(),
			}),
		)
	}
	return nil
}

func (cmd *PluginInstall) ensurePluginIsSafeForInstallation(pluginMetadata *plugin.PluginMetadata, pluginDestinationFilepath string, pluginSourceFilepath string) error {
	plugins := cmd.pluginConfig.Plugins()
	if pluginMetadata.Name == "" {
		return errors.New(T(
			"Unable to obtain plugin name for executable {{.Executable}}",
			map[string]interface{}{
				"Executable": pluginSourceFilepath,
			}),
		)
	}

	if _, ok := plugins[pluginMetadata.Name]; ok {
		return errors.New(T(
			"Plugin name {{.PluginName}} is already taken",
			map[string]interface{}{
				"PluginName": pluginMetadata.Name,
			}),
		)
	}

	if pluginMetadata.Commands == nil {
		return errors.New(T(
			"Error getting command list from plugin {{.FilePath}}",
			map[string]interface{}{
				"FilePath": pluginSourceFilepath,
			}),
		)
	}

	for _, pluginCmd := range pluginMetadata.Commands {
		//check for command conflicting core commands/alias
		if pluginCmd.Name == "help" || commandregistry.Commands.CommandExists(pluginCmd.Name) {
			return errors.New(T(
				"Command `{{.Command}}` in the plugin being installed is a native CF command/alias.  Rename the `{{.Command}}` command in the plugin being installed in order to enable its installation and use.",
				map[string]interface{}{
					"Command": pluginCmd.Name,
				}),
			)
		}

		//check for alias conflicting core command/alias
		if pluginCmd.Alias == "help" || commandregistry.Commands.CommandExists(pluginCmd.Alias) {
			return errors.New(T(
				"Alias `{{.Command}}` in the plugin being installed is a native CF command/alias.  Rename the `{{.Command}}` command in the plugin being installed in order to enable its installation and use.",
				map[string]interface{}{
					"Command": pluginCmd.Alias,
				}),
			)
		}

		for installedPluginName, installedPlugin := range plugins {
			for _, installedPluginCmd := range installedPlugin.Commands {

				//check for command conflicting other plugin commands/alias
				if installedPluginCmd.Name == pluginCmd.Name || installedPluginCmd.Alias == pluginCmd.Name {
					return errors.New(T(
						"Command `{{.Command}}` is a command/alias in plugin '{{.PluginName}}'.  You could try uninstalling plugin '{{.PluginName}}' and then install this plugin in order to invoke the `{{.Command}}` command.  However, you should first fully understand the impact of uninstalling the existing '{{.PluginName}}' plugin.",
						map[string]interface{}{
							"Command":    pluginCmd.Name,
							"PluginName": installedPluginName,
						}),
					)
				}

				//check for alias conflicting other plugin commands/alias
				if pluginCmd.Alias != "" && (installedPluginCmd.Name == pluginCmd.Alias || installedPluginCmd.Alias == pluginCmd.Alias) {
					return errors.New(T(
						"Alias `{{.Command}}` is a command/alias in plugin '{{.PluginName}}'.  You could try uninstalling plugin '{{.PluginName}}' and then install this plugin in order to invoke the `{{.Command}}` command.  However, you should first fully understand the impact of uninstalling the existing '{{.PluginName}}' plugin.",
						map[string]interface{}{
							"Command":    pluginCmd.Alias,
							"PluginName": installedPluginName,
						}),
					)
				}
			}
		}
	}
	return nil
}

func (cmd *PluginInstall) installPlugin(pluginMetadata *plugin.PluginMetadata, pluginDestinationFilepath, pluginSourceFilepath string) error {
	err := fileutils.CopyPathToPath(pluginSourceFilepath, pluginDestinationFilepath)
	if err != nil {
		return errors.New(T(
			"Could not copy plugin binary: \n{{.Error}}",
			map[string]interface{}{
				"Error": err.Error(),
			}),
		)
	}

	configMetadata := pluginconfig.PluginMetadata{
		Location: pluginDestinationFilepath,
		Version:  pluginMetadata.Version,
		Commands: pluginMetadata.Commands,
	}

	cmd.pluginConfig.SetPlugin(pluginMetadata.Name, configMetadata)
	return nil
}

func (cmd *PluginInstall) runBinaryAndObtainPluginMetadata(pluginSourceFilepath string) (*plugin.PluginMetadata, error) {
	err := cmd.rpcService.Start()
	if err != nil {
		return nil, err
	}
	defer cmd.rpcService.Stop()

	err = cmd.runPluginBinary(pluginSourceFilepath, cmd.rpcService.Port())
	if err != nil {
		return nil, err
	}

	c := cmd.rpcService.RpcCmd
	c.MetadataMutex.RLock()
	defer c.MetadataMutex.RUnlock()
	return c.PluginMetadata, nil
}

func (cmd *PluginInstall) runPluginBinary(location string, servicePort string) error {
	pluginInvocation := exec.Command(location, servicePort, "SendMetadata")

	err := pluginInvocation.Run()
	if err != nil {
		return err
	}
	return nil
}
