package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/command"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/fileutils"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/cli/plugin/rpc"
	"github.com/codegangsta/cli"
)

type PluginInstall struct {
	ui       terminal.UI
	config   plugin_config.PluginConfiguration
	coreCmds map[string]command.Command
}

func NewPluginInstall(ui terminal.UI, config plugin_config.PluginConfiguration, coreCmds map[string]command.Command) *PluginInstall {
	return &PluginInstall{
		ui:       ui,
		config:   config,
		coreCmds: coreCmds,
	}
}

func (cmd *PluginInstall) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "install-plugin",
		Description: T("Install the plugin defined in command argument"),
		Usage: T(`CF_NAME install-plugin URL or LOCAL-PATH/TO/PLUGIN

EXAMPLE:
   cf install-plugin https://github.com/cf-experimental/plugin-foobar
   cf install-plugin ~/Downloads/plugin-foobar
`),
	}
}

func (cmd *PluginInstall) GetRequirements(_ requirements.Factory, c *cli.Context) (req []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	return
}

func (cmd *PluginInstall) Run(c *cli.Context) {
	var downloader fileutils.Downloader

	pluginSourceFilepath := c.Args()[0]

	if filepath.Dir(pluginSourceFilepath) == "." {
		pluginSourceFilepath = "./" + filepath.Clean(pluginSourceFilepath)
	}

	cmd.ui.Say(fmt.Sprintf(T("Installing plugin {{.PluginPath}}...", map[string]interface{}{"PluginPath": pluginSourceFilepath})))

	if !cmd.ensureCandidatePluginBinaryExistsAtGivenPath(pluginSourceFilepath) {
		cmd.ui.Say("")
		cmd.ui.Say(T("File not found locally, attempting to download binary file from internet ..."))
		pluginSourceFilepath = cmd.tryDownloadPluginBinaryfromGivenPath(pluginSourceFilepath, downloader)
	}

	_, pluginExecutableName := filepath.Split(pluginSourceFilepath)

	pluginDestinationFilepath := filepath.Join(cmd.config.GetPluginPath(), pluginExecutableName)

	cmd.ensurePluginBinaryWithSameFileNameDoesNotAlreadyExist(pluginDestinationFilepath, pluginExecutableName)

	pluginMetadata := cmd.runBinaryAndObtainPluginMetadata(pluginSourceFilepath)

	cmd.ensurePluginIsSafeForInstallation(pluginMetadata, pluginDestinationFilepath, pluginSourceFilepath)

	cmd.installPlugin(pluginMetadata, pluginDestinationFilepath, pluginSourceFilepath)

	if downloader != nil {
		err := downloader.RemoveFile()
		if err != nil {
			cmd.ui.Say(T("Problem removing downloaded binary in temp directory: ") + err.Error())
		}
	}

	cmd.ui.Ok()
	cmd.ui.Say(fmt.Sprintf(T("Plugin {{.PluginName}} v{{.Version}} successfully installed.", map[string]interface{}{"PluginName": pluginMetadata.Name, "Version": fmt.Sprintf("%d.%d.%d", pluginMetadata.Version.Major, pluginMetadata.Version.Minor, pluginMetadata.Version.Build)})))
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
	plugins := cmd.config.Plugins()
	if pluginMetadata.Name == "" {
		cmd.ui.Failed(fmt.Sprintf(T("Unable to obtain plugin name for executable {{.Executable}}", map[string]interface{}{"Executable": pluginSourceFilepath})))
	}

	if _, ok := plugins[pluginMetadata.Name]; ok {
		cmd.ui.Failed(fmt.Sprintf(T("Plugin name {{.PluginName}} is already taken", map[string]interface{}{"PluginName": pluginMetadata.Name})))
	}

	if pluginMetadata.Commands == nil {
		cmd.ui.Failed(fmt.Sprintf(T("Error getting command list from plugin {{.FilePath}}", map[string]interface{}{"FilePath": pluginSourceFilepath})))
	}

	shortNames := cmd.getShortNames()

	for _, pluginCmd := range pluginMetadata.Commands {
		//check for command conflicting core commands/alias
		if _, exists := cmd.coreCmds[pluginCmd.Name]; exists || shortNames[pluginCmd.Name] || pluginCmd.Name == "help" {
			cmd.ui.Failed(fmt.Sprintf(T("Command `{{.Command}}` in the plugin being installed is a native CF command/alias.  Rename the `{{.Command}}` command in the plugin being installed in order to enable its installation and use.",
				map[string]interface{}{"Command": pluginCmd.Name})))
		}

		//check for alias conflicting core command/alias
		if _, exists := cmd.coreCmds[pluginCmd.Alias]; exists || shortNames[pluginCmd.Alias] || pluginCmd.Alias == "help" {
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
	err := fileutils.CopyFile(pluginDestinationFilepath, pluginSourceFilepath)
	if err != nil {
		cmd.ui.Failed(fmt.Sprintf(T("Could not copy plugin binary: \n{{.Error}}", map[string]interface{}{"Error": err.Error()})))
	}

	configMetadata := plugin_config.PluginMetadata{
		Location: pluginDestinationFilepath,
		Version:  pluginMetadata.Version,
		Commands: pluginMetadata.Commands,
	}

	cmd.config.SetPlugin(pluginMetadata.Name, configMetadata)
}

func (cmd *PluginInstall) runBinaryAndObtainPluginMetadata(pluginSourceFilepath string) *plugin.PluginMetadata {
	rpcService, err := rpc.NewRpcService(nil, nil, nil)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	err = rpcService.Start()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	defer rpcService.Stop()

	cmd.runPluginBinary(pluginSourceFilepath, rpcService.Port())

	return rpcService.RpcCmd.PluginMetadata
}

func (cmd *PluginInstall) ensureCandidatePluginBinaryExistsAtGivenPath(pluginSourceFilepath string) bool {
	_, err := os.Stat(pluginSourceFilepath)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func (cmd *PluginInstall) tryDownloadPluginBinaryfromGivenPath(pluginSourceFilepath string, downloader fileutils.Downloader) string {

	savePath := os.TempDir()
	downloader = fileutils.NewDownloader(savePath)
	size, filename, err := downloader.DownloadFile(pluginSourceFilepath)

	if err != nil {
		cmd.ui.Failed(fmt.Sprintf(T("Download attempt failed: {{.Error}}\n\nUnable to install, plugin is not available from local/internet.", map[string]interface{}{"Error": err.Error()})))
	}

	cmd.ui.Say(fmt.Sprintf("%d "+T("bytes downloaded")+"...", size))

	executablePath := filepath.Join(savePath, filename)
	os.Chmod(executablePath, 0700)

	return executablePath
}

func (cmd *PluginInstall) getShortNames() map[string]bool {
	shortNames := make(map[string]bool)
	for _, singleCmd := range cmd.coreCmds {
		metaData := singleCmd.Metadata()
		if metaData.ShortName != "" {
			shortNames[metaData.ShortName] = true
		}
	}
	return shortNames
}

func (cmd *PluginInstall) runPluginBinary(location string, servicePort string) {
	pluginInvocation := exec.Command(location, servicePort, "SendMetadata")

	err := pluginInvocation.Run()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
}
