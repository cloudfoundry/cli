package plugin

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type PluginUninstall struct {
	ui terminal.UI
}

func NewPluginUninstall(ui terminal.UI) *PluginUninstall {
	return &PluginUninstall{
		ui: ui,
	}
}

func (cmd *PluginUninstall) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        T("uninstall-plugin"),
		Description: T("PLUGIN-NAME - Uninstall the plugin defined in command argument"),
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

	pluginsConfig := plugin_config.NewPluginConfig(func(err error) {
		cmd.ui.Failed(err.Error())
	})

	plugins := pluginsConfig.Plugins()

	if _, ok := plugins[pluginName]; !ok {
		cmd.ui.Failed(fmt.Sprintf(T("Plugin name {{.PluginName}} does not exist", pluginNameMap)))
	}

	pluginPath := plugins[pluginName]
	os.Remove(pluginPath)

	pluginsConfig.RemovePlugin(pluginName)

	cmd.ui.Ok()
	cmd.ui.Say(fmt.Sprintf(T("Plugin {{.PluginName}} successfully uninstalled.", pluginNameMap)))
}
