package plugin

import (
	"errors"
	"net/rpc"
	"os"
	"os/exec"
	"time"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/pluginconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	rpcService "github.com/cloudfoundry/cli/plugin/rpc"
)

type PluginUninstall struct {
	ui         terminal.UI
	config     pluginconfig.PluginConfiguration
	rpcService *rpcService.CliRpcService
}

func init() {
	commandregistry.Register(&PluginUninstall{})
}

func (cmd *PluginUninstall) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "uninstall-plugin",
		Description: T("Uninstall the plugin defined in command argument"),
		Usage: []string{
			T("CF_NAME uninstall-plugin PLUGIN-NAME"),
		},
	}
}

func (cmd *PluginUninstall) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("uninstall-plugin"))
	}

	reqs := []requirements.Requirement{}
	return reqs
}

func (cmd *PluginUninstall) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.PluginConfig

	//reset rpc registration in case there is other running instance,
	//each service can only be registered once
	server := rpc.NewServer()

	RPCService, err := rpcService.NewRpcService(deps.TeePrinter, deps.TeePrinter, deps.Config, deps.RepoLocator, rpcService.NewCommandRunner(), deps.Logger, cmd.ui.Writer(), server)
	if err != nil {
		cmd.ui.Failed("Error initializing RPC service: " + err.Error())
	}

	cmd.rpcService = RPCService

	return cmd
}

func (cmd *PluginUninstall) Execute(c flags.FlagContext) error {
	pluginName := c.Args()[0]
	pluginNameMap := map[string]interface{}{"PluginName": pluginName}

	cmd.ui.Say(T("Uninstalling plugin {{.PluginName}}...", pluginNameMap))

	plugins := cmd.config.Plugins()

	if _, ok := plugins[pluginName]; !ok {
		return errors.New(T("Plugin name {{.PluginName}} does not exist", pluginNameMap))
	}

	pluginMetadata := plugins[pluginName]

	warn, err := cmd.notifyPluginUninstalling(pluginMetadata)
	if err != nil {
		return err
	}
	if warn != nil {
		cmd.ui.Say("Error invoking plugin: " + warn.Error() + ". Process to uninstall ...")
	}

	time.Sleep(500 * time.Millisecond) //prevent 'process being used' error in Windows

	err = os.Remove(pluginMetadata.Location)
	if err != nil {
		cmd.ui.Warn("Error removing plugin binary: " + err.Error())
	}

	cmd.config.RemovePlugin(pluginName)

	cmd.ui.Ok()
	cmd.ui.Say(T("Plugin {{.PluginName}} successfully uninstalled.", pluginNameMap))
	return nil
}

func (cmd *PluginUninstall) notifyPluginUninstalling(meta pluginconfig.PluginMetadata) (error, error) {
	err := cmd.rpcService.Start()
	if err != nil {
		return nil, err
	}
	defer cmd.rpcService.Stop()

	pluginInvocation := exec.Command(meta.Location, cmd.rpcService.Port(), "CLI-MESSAGE-UNINSTALL")
	pluginInvocation.Stdout = os.Stdout

	return pluginInvocation.Run(), nil
}
