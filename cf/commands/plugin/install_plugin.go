package plugin

import (
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/actors/plugininstaller"
	"github.com/cloudfoundry/cli/cf/actors/pluginrepo"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/configuration/pluginconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/downloader"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/cli/utils"
	"github.com/cloudfoundry/gofileutils/fileutils"

	pluginRPCService "github.com/cloudfoundry/cli/plugin/rpc"
)

type PluginInstall struct {
	ui           terminal.UI
	config       coreconfig.Reader
	pluginConfig pluginconfig.PluginConfiguration
	pluginRepo   pluginrepo.PluginRepo
	checksum     utils.Sha1Checksum
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

func (cmd *PluginInstall) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("install-plugin"))
	}

	reqs := []requirements.Requirement{}
	return reqs
}

func (cmd *PluginInstall) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.pluginConfig = deps.PluginConfig
	cmd.pluginRepo = deps.PluginRepo
	cmd.checksum = deps.ChecksumUtil

	//reset rpc registration in case there is other running instance,
	//each service can only be registered once
	rpc.DefaultServer = rpc.NewServer()

	rpcService, err := pluginRPCService.NewRpcService(deps.TeePrinter, deps.TeePrinter, deps.Config, deps.RepoLocator, pluginRPCService.NewCommandRunner(), deps.Logger, cmd.ui.Writer())
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

	cmd.ui.Say(fmt.Sprintf(T("Installing plugin {{.PluginPath}}...", map[string]interface{}{"PluginPath": pluginExecutableName})))

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
		if pluginCmd.Name == "help" || commandregistry.Commands.CommandExists(pluginCmd.Name) {
			cmd.ui.Failed(fmt.Sprintf(T("Command `{{.Command}}` in the plugin being installed is a native CF command/alias.  Rename the `{{.Command}}` command in the plugin being installed in order to enable its installation and use.",
				map[string]interface{}{"Command": pluginCmd.Name})))
		}

		//check for alias conflicting core command/alias
		if pluginCmd.Alias == "help" || commandregistry.Commands.CommandExists(pluginCmd.Alias) {
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

	configMetadata := pluginconfig.PluginMetadata{
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
