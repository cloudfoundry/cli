package plugin

import (
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"time"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	rpcService "github.com/cloudfoundry/cli/plugin/rpc"
)

type PluginUninstall struct {
	ui         terminal.UI
	config     plugin_config.PluginConfiguration
	rpcService *rpcService.CliRpcService
}

func init() {
	command_registry.Register(&PluginUninstall{})
}

func (cmd *PluginUninstall) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "uninstall-plugin",
		Description: T("Uninstall the plugin defined in command argument"),
		Usage:       T("CF_NAME uninstall-plugin PLUGIN-NAME"),
	}
}

func (cmd *PluginUninstall) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("uninstall-plugin"))
	}

	return
}

func (cmd *PluginUninstall) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.PluginConfig

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

func (cmd *PluginUninstall) Execute(c flags.FlagContext) {
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

	time.Sleep(500 * time.Millisecond) //prevent 'process being used' error in Windows

	err = os.Remove(pluginMetadata.Location)
	if err != nil {
		cmd.ui.Warn("Error removing plugin binary: " + err.Error())
	}

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
