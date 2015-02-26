package plugin

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin/rpc"
	"github.com/codegangsta/cli"
)

type PluginUninstall struct {
	ui         terminal.UI
	config     plugin_config.PluginConfiguration
	rpcService *rpc.CliRpcService
}

func NewPluginUninstall(ui terminal.UI, config plugin_config.PluginConfiguration, rpcService *rpc.CliRpcService) *PluginUninstall {
	return &PluginUninstall{
		ui:         ui,
		config:     config,
		rpcService: rpcService,
	}
}

func (cmd *PluginUninstall) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "uninstall-plugin",
		Description: T("Uninstall the plugin defined in command argument"),
		Usage:       T("CF_NAME uninstall-plugin PLUGIN-NAME"),
	}
}

func (cmd *PluginUninstall) GetRequirements(_ requirements.Factory, c *cli.Context) (req []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	return
}

func (cmd *PluginUninstall) Run(c *cli.Context) {
	pluginName := c.Args()[0]
	pluginNameMap := map[string]interface{}{"PluginName": pluginName}

	cmd.ui.Say(fmt.Sprintf(T("Uninstalling plugin {{.PluginName}}...", pluginNameMap)))

	plugins := cmd.config.Plugins()

	if _, ok := plugins[pluginName]; !ok {
		cmd.ui.Failed(fmt.Sprintf(T("Plugin name {{.PluginName}} does not exist", pluginNameMap)))
	}

	pluginMetadata := plugins[pluginName]

	err := cmd.notifyPluginUninstalling(pluginMetadata)
	if err != nil {
		cmd.ui.Say("Error invoking plugin: " + err.Error() + ". Process to uninstall ...")
	}

	os.Remove(pluginMetadata.Location)

	cmd.config.RemovePlugin(pluginName)

	cmd.ui.Ok()
	cmd.ui.Say(fmt.Sprintf(T("Plugin {{.PluginName}} successfully uninstalled.", pluginNameMap)))
}

func (cmd *PluginUninstall) notifyPluginUninstalling(meta plugin_config.PluginMetadata) error {
	err := cmd.rpcService.Start()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	defer cmd.rpcService.Stop()

	pluginInvocation := exec.Command(meta.Location, cmd.rpcService.Port(), "CLI-MESSAGE-UNINSTALL")
	pluginInvocation.Stdout = os.Stdout

	return pluginInvocation.Run()
}
