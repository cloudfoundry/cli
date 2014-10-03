package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/fileutils"
	"github.com/codegangsta/cli"
)

type PluginInstall struct {
	ui terminal.UI
}

func NewPluginInstall(ui terminal.UI) *PluginInstall {
	return &PluginInstall{
		ui: ui,
	}
}

func (cmd *PluginInstall) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        T("install-plugin"),
		Description: T("install-plugin PATH/TO/PLUGIN-NAME  - Install the plugin defined in command argument"),
		Usage:       T("CF_NAME install-plugin PATH/TO/PLUGIN"),
	}
}

func (cmd *PluginInstall) GetRequirements(_ requirements.Factory, c *cli.Context) (req []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	return
}

func (cmd *PluginInstall) Run(c *cli.Context) {
	pluginPath := c.Args()[0]

	_, pluginName := filepath.Split(pluginPath)

	pluginsConfig := plugin_config.NewPluginConfig(func(err error) { cmd.ui.Failed(err.Error()) })
	plugins := pluginsConfig.Plugins()

	if _, ok := plugins[pluginName]; ok {
		cmd.ui.Failed(fmt.Sprintf(T("Plugin name {{.PluginName}} is already taken", map[string]interface{}{"PluginName": pluginName})))
	}

	cmd.ui.Say(fmt.Sprintf(T("Installing plugin {{.PluginName}}...", map[string]interface{}{"PluginName": pluginName})))

	pluginsDir := filepath.Join(config_helpers.UserHomeDir(), ".cf", "plugins")
	pluginExecutable := filepath.Join(pluginsDir, pluginName)

	_, err := os.Stat(pluginExecutable)
	if err == nil || os.IsExist(err) {
		cmd.ui.Failed(fmt.Sprintf(T("The file {{.PluginName}} already exists under the plugin directory.\n", map[string]interface{}{"PluginName": pluginName})))
	} else if !os.IsNotExist(err) {
		cmd.ui.Failed(fmt.Sprintf(T("Unexpected error has occurred:\n{{.Error}}", map[string]interface{}{"Error": err.Error()})))
	}

	err = fileutils.CopyFile(pluginExecutable, pluginPath)
	if err != nil {
		cmd.ui.Failed(fmt.Sprintf(T("Could not copy plugin binary: \n{{.Error}}", map[string]interface{}{"Error": err.Error()})))
	}

	pluginsConfig.SetPlugin(pluginName, pluginExecutable)

	cmd.ui.Ok()
	cmd.ui.Say(fmt.Sprintf(T("Plugin {{.PluginName}} successfully installed.", map[string]interface{}{"PluginName": pluginName})))
}
