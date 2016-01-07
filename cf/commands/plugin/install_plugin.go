package plugin

import (
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/actors/plugin_installer"
	"github.com/cloudfoundry/cli/cf/actors/plugin_repo"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/downloader"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/cli/utils"
	"github.com/cloudfoundry/gofileutils/fileutils"

	rpcService "github.com/cloudfoundry/cli/plugin/rpc"
)

type PluginInstall struct {
	ui           terminal.UI
	config       core_config.Reader
	pluginConfig plugin_config.PluginConfiguration
	pluginRepo   plugin_repo.PluginRepo
	checksum     utils.Sha1Checksum
	rpcService   *rpcService.CliRpcService
}

func init() {
	command_registry.Register(&PluginInstall{})
}

func (cmd *PluginInstall) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["r"] = &cliFlags.StringFlag{ShortName: "r", Usage: T("repo name where the plugin binary is located")}
	fs["f"] = &cliFlags.BoolFlag{ShortName: "f", Usage: T("Force install of plugin without prompt")}

	return command_registry.CommandMetadata{
		Name:        "install-plugin",
		Description: T("Install the plugin defined in command argument"),
		Usage: T(`CF_NAME install-plugin URL or LOCAL-PATH/TO/PLUGIN [-r REPO_NAME] [-f]

The command will download the plugin binary from repository if '-r' is provided
Prompts for confirmation unless '-f' is provided

EXAMPLE:
   cf install-plugin https://github.com/cf-experimental/plugin-foobar
   cf install-plugin ~/Downloads/plugin-foobar
   cf install-plugin plugin-echo -r My-Repo 
`),
		Flags:     fs,
		TotalArgs: 1,
	}
}

func (cmd *PluginInstall) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("install-plugin"))
	}

	return
}

func (cmd *PluginInstall) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.pluginConfig = deps.PluginConfig
	cmd.pluginRepo = deps.PluginRepo
	cmd.checksum = deps.ChecksumUtil

	//reset rpc registration in case there is other running instance,
	//each service can only be registered once
	rpc.DefaultServer = rpc.NewServer()

	rpcService, err := rpcService.NewRpcService(deps.TeePrinter, deps.TeePrinter, deps.Config, deps.RepoLocator, rpcService.NewNonCodegangstaRunner())
	if err != nil {
		cmd.ui.Failed("Error initializing RPC service: " + err.Error())
	}

	cmd.rpcService = rpcService

	return cmd
}

func (cmd *PluginInstall) Execute(c flags.FlagContext) {
	if !cmd.confirmWithUser(
		c,
		T("**Attention: Plugins are binaries written by potentially untrusted authors. Install and use plugins at your own risk.**\n\nDo you want to install the plugin {{.Plugin}}? (y or n)", map[string]interface{}{"Plugin": c.Args()[0]}),
	) {
		cmd.ui.Failed(T("Plugin installation cancelled"))
	}

	fileDownloader := downloader.NewDownloader(os.TempDir())

	removeTmpFile := func() {
		err := fileDownloader.RemoveFile()
		if err != nil {
			cmd.ui.Say(T("Problem removing downloaded binary in temp directory: ") + err.Error())
		}
	}
	defer removeTmpFile()

	deps := &plugin_installer.PluginInstallerContext{
		Checksummer:    cmd.checksum,
		GetPluginRepos: cmd.config.PluginRepos,
		FileDownloader: fileDownloader,
		PluginRepo:     cmd.pluginRepo,
		RepoName:       c.String("r"),
		Ui:             cmd.ui,
	}
	installer := plugin_installer.NewPluginInstaller(deps)
	pluginSourceFilepath := installer.Install(c.Args()[0])

	cmd.ui.Say(fmt.Sprintf(T("Installing plugin {{.PluginPath}}...", map[string]interface{}{"PluginPath": pluginSourceFilepath})))

	_, pluginExecutableName := filepath.Split(pluginSourceFilepath)

	pluginDestinationFilepath := filepath.Join(cmd.pluginConfig.GetPluginPath(), pluginExecutableName)

	cmd.ensurePluginBinaryWithSameFileNameDoesNotAlreadyExist(pluginDestinationFilepath, pluginExecutableName)

	pluginMetadata := cmd.runBinaryAndObtainPluginMetadata(pluginSourceFilepath)

	cmd.ensurePluginIsSafeForInstallation(pluginMetadata, pluginDestinationFilepath, pluginSourceFilepath)

	cmd.installPlugin(pluginMetadata, pluginDestinationFilepath, pluginSourceFilepath)

	cmd.ui.Ok()
	cmd.ui.Say(fmt.Sprintf(T("Plugin {{.PluginName}} v{{.Version}} successfully installed.", map[string]interface{}{"PluginName": pluginMetadata.Name, "Version": fmt.Sprintf("%d.%d.%d", pluginMetadata.Version.Major, pluginMetadata.Version.Minor, pluginMetadata.Version.Build)})))
}

func (cmd *PluginInstall) confirmWithUser(c flags.FlagContext, prompt string) bool {
	return c.Bool("f") || cmd.ui.Confirm(prompt)
}

func (cmd *PluginInstall) ensurePluginBinaryWithSameFileNameDoesNotAlreadyExist(pluginDestinationFilepath, pluginExecutableName string) {
	_, err := os.Stat(pluginDestinationFilepath)
	if err == nil || os.IsExist(err) {
		cmd.ui.Failed(fmt.Sprintf(T("The file {{.PluginExecutableName}} already exists under the plugin directory.\n",
			map[string]interface{}{
				"PluginExecutableName": pluginExecutableName,
			})))
	} else if !os.IsNotExist(err) {
		cmd.ui.Failed(fmt.Sprintf(T("Unexpected error has occurred:\n{{.Error}}", map[string]interface{}{"Error": err.Error()})))
	}
}

func (cmd *PluginInstall) ensurePluginIsSafeForInstallation(pluginMetadata *plugin.PluginMetadata, pluginDestinationFilepath string, pluginSourceFilepath string) {
	plugins := cmd.pluginConfig.Plugins()
	if pluginMetadata.Name == "" {
		cmd.ui.Failed(fmt.Sprintf(T("Unable to obtain plugin name for executable {{.Executable}}", map[string]interface{}{"Executable": pluginSourceFilepath})))
	}

	if _, ok := plugins[pluginMetadata.Name]; ok {
		cmd.ui.Failed(fmt.Sprintf(T("Plugin name {{.PluginName}} is already taken", map[string]interface{}{"PluginName": pluginMetadata.Name})))
	}

	if pluginMetadata.Commands == nil {
		cmd.ui.Failed(fmt.Sprintf(T("Error getting command list from plugin {{.FilePath}}", map[string]interface{}{"FilePath": pluginSourceFilepath})))
	}

	for _, pluginCmd := range pluginMetadata.Commands {

		//check for command conflicting core commands/alias
		if pluginCmd.Name == "help" || command_registry.Commands.CommandExists(pluginCmd.Name) {
			cmd.ui.Failed(fmt.Sprintf(T("Command `{{.Command}}` in the plugin being installed is a native CF command/alias.  Rename the `{{.Command}}` command in the plugin being installed in order to enable its installation and use.",
				map[string]interface{}{"Command": pluginCmd.Name})))
		}

		//check for alias conflicting core command/alias
		if pluginCmd.Alias == "help" || command_registry.Commands.CommandExists(pluginCmd.Alias) {
			cmd.ui.Failed(fmt.Sprintf(T("Alias `{{.Command}}` in the plugin being installed is a native CF command/alias.  Rename the `{{.Command}}` command in the plugin being installed in order to enable its installation and use.",
				map[string]interface{}{"Command": pluginCmd.Alias})))
		}

		for installedPluginName, installedPlugin := range plugins {
			for _, installedPluginCmd := range installedPlugin.Commands {

				//check for command conflicting other plugin commands/alias
				if installedPluginCmd.Name == pluginCmd.Name || installedPluginCmd.Alias == pluginCmd.Name {
					cmd.ui.Failed(fmt.Sprintf(T("Command `{{.Command}}` is a command/alias in plugin '{{.PluginName}}'.  You could try uninstalling plugin '{{.PluginName}}' and then install this plugin in order to invoke the `{{.Command}}` command.  However, you should first fully understand the impact of uninstalling the existing '{{.PluginName}}' plugin.",
						map[string]interface{}{"Command": pluginCmd.Name, "PluginName": installedPluginName})))
				}

				//check for alias conflicting other plugin commands/alias
				if pluginCmd.Alias != "" && (installedPluginCmd.Name == pluginCmd.Alias || installedPluginCmd.Alias == pluginCmd.Alias) {
					cmd.ui.Failed(fmt.Sprintf(T("Alias `{{.Command}}` is a command/alias in plugin '{{.PluginName}}'.  You could try uninstalling plugin '{{.PluginName}}' and then install this plugin in order to invoke the `{{.Command}}` command.  However, you should first fully understand the impact of uninstalling the existing '{{.PluginName}}' plugin.",
						map[string]interface{}{"Command": pluginCmd.Alias, "PluginName": installedPluginName})))
				}
			}
		}
	}

}

func (cmd *PluginInstall) installPlugin(pluginMetadata *plugin.PluginMetadata, pluginDestinationFilepath, pluginSourceFilepath string) {
	err := fileutils.CopyPathToPath(pluginSourceFilepath, pluginDestinationFilepath)
	if err != nil {
		cmd.ui.Failed(fmt.Sprintf(T("Could not copy plugin binary: \n{{.Error}}", map[string]interface{}{"Error": err.Error()})))
	}

	configMetadata := plugin_config.PluginMetadata{
		Location: pluginDestinationFilepath,
		Version:  pluginMetadata.Version,
		Commands: pluginMetadata.Commands,
	}

	cmd.pluginConfig.SetPlugin(pluginMetadata.Name, configMetadata)
}

func (cmd *PluginInstall) runBinaryAndObtainPluginMetadata(pluginSourceFilepath string) *plugin.PluginMetadata {
	err := cmd.rpcService.Start()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	defer cmd.rpcService.Stop()

	cmd.runPluginBinary(pluginSourceFilepath, cmd.rpcService.Port())

	return cmd.rpcService.RpcCmd.PluginMetadata
}

func (cmd *PluginInstall) runPluginBinary(location string, servicePort string) {
	pluginInvocation := exec.Command(location, servicePort, "SendMetadata")

	err := pluginInvocation.Run()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
}
