package plugin

import (
	"fmt"
	"os"
	"path"

	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/fileutils"
	"github.com/codegangsta/cli"
)

type PluginInstall struct {
	ui     terminal.UI
	config configuration.ReadWriter
}

func NewPluginInstall(ui terminal.UI, config configuration.ReadWriter) *PluginInstall {
	return &PluginInstall{
		ui:     ui,
		config: config,
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

	_, pluginName := path.Split(pluginPath)

	plugins := cmd.config.Plugins()

	if _, ok := plugins[pluginName]; ok {
		cmd.ui.Failed(fmt.Sprintf(T("Plugin name {{.PluginName}} is already taken", map[string]interface{}{"PluginName": pluginName})))
	}

	cmd.ui.Say(fmt.Sprintf(T("Installing plugin {{.PluginName}}...", map[string]interface{}{"PluginName": pluginName})))

	homeDir := path.Join(cmd.config.UserHomePath(), ".cf", "plugin")
	err := os.MkdirAll(homeDir, 0700)
	if err != nil {
		cmd.ui.Failed(fmt.Sprintf(T("Could not create the plugin directory: \n{{.Error}}", map[string]interface{}{"Error": err.Error()})))
	}

	_, err = os.Stat(path.Join(homeDir, pluginName))
	if err == nil || os.IsExist(err) {
		cmd.ui.Failed(fmt.Sprintf(T("The file {{.PluginName}} already exists under the plugin directory.\n", map[string]interface{}{"PluginName": pluginName})))
	} else if !os.IsNotExist(err) {
		cmd.ui.Failed(fmt.Sprintf(T("Unexpected error has occurred:\n{{.Error}}", map[string]interface{}{"Error": err.Error()})))
	}

	err = fileutils.CopyFile(path.Join(homeDir, pluginName), pluginPath)
	if err != nil {
		cmd.ui.Failed(fmt.Sprintf(T("Could not copy plugin binary: \n{{.Error}}", map[string]interface{}{"Error": err.Error()})))
	}

	cmd.config.SetPlugin(pluginName, path.Join(homeDir, pluginName))

	cmd.ui.Ok()
	cmd.ui.Say(fmt.Sprintf(T("Plugin {{.PluginName}} successfully installed.", map[string]interface{}{"PluginName": pluginName})))
}
