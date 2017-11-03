package plugin

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/plugin/shared"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

//go:generate counterfeiter . UninstallPluginActor

type UninstallPluginActor interface {
	UninstallPlugin(uninstaller pluginaction.PluginUninstaller, name string) error
}

type UninstallPluginCommand struct {
	RequiredArgs    flag.PluginName `positional-args:"yes"`
	usage           interface{}     `usage:"CF_NAME uninstall-plugin PLUGIN-NAME"`
	relatedCommands interface{}     `related_commands:"plugins"`

	Config command.Config
	UI     command.UI
	Actor  UninstallPluginActor
}

func (cmd *UninstallPluginCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.Actor = pluginaction.NewActor(config, nil)
	return nil
}

func (cmd UninstallPluginCommand) Execute(args []string) error {
	pluginName := cmd.RequiredArgs.PluginName
	plugin, exist := cmd.Config.GetPluginCaseInsensitive(pluginName)
	if !exist {
		return translatableerror.PluginNotFoundError{PluginName: pluginName}
	}

	cmd.UI.DisplayTextWithFlavor("Uninstalling plugin {{.PluginName}}...",
		map[string]interface{}{
			"PluginName": plugin.Name,
		})

	rpcService, err := shared.NewRPCService(cmd.Config, cmd.UI)
	if err != nil {
		return err
	}

	err = cmd.Actor.UninstallPlugin(rpcService, plugin.Name)
	if err != nil {
		switch e := err.(type) {
		case actionerror.PluginBinaryRemoveFailedError:
			return translatableerror.PluginBinaryRemoveFailedError{
				Err: e.Err,
			}
		case actionerror.PluginExecuteError:
			return translatableerror.PluginBinaryUninstallError{
				Err: e.Err,
			}
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("Plugin {{.PluginName}} {{.PluginVersion}} successfully uninstalled.",
		map[string]interface{}{
			"PluginName":    plugin.Name,
			"PluginVersion": plugin.Version,
		})

	return nil
}
