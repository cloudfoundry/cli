package plugin

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

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

	cmd.ui.Say(fmt.Sprintf(T("Installing plugin {{.PluginPath}}...", map[string]interface{}{"PluginPath": pluginPath})))

	cmd.validateCandidatePluginPath(pluginPath)

	_, pluginExecutableName := filepath.Split(pluginPath)

	plugins := cmd.config.Plugins()
	pluginExecutable := filepath.Join(cmd.config.GetPluginPath(), pluginExecutableName)

	cmd.ensurePluginDoesNotExist(pluginExecutable, pluginExecutableName)

	rpcService, err := rpc.NewRpcService()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	err = rpcService.Start()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	defer rpcService.Stop()

	cmd.runPluginBinary(pluginPath, rpcService.Port())

	pluginMetadata := rpcService.RpcCmd.ReturnData.(plugin.PluginMetadata)
	if pluginMetadata.Name == "" {
		cmd.ui.Failed(fmt.Sprintf("Unable to obtain plugin name for executable %s", pluginPath))
	}

	if _, ok := plugins[pluginMetadata.Name]; ok {
		cmd.ui.Failed(fmt.Sprintf(T("Plugin name {{.PluginName}} is already taken", map[string]interface{}{"PluginName": pluginMetadata.Name})))
	}

	if pluginMetadata.Commands == nil {
		cmd.ui.Failed(fmt.Sprintf("Error getting command list from plugin %s", pluginPath))
	}

	for k, pluginCmd := range pluginMetadata.Commands {
		println(k, " . ", pluginCmd.Name)
		if _, exists := cmd.coreCmds[pluginCmd.Name]; exists {
			cmd.ui.Failed(fmt.Sprintf("Plugin '%s' cannot be installed from '%s' at this time because the command 'cf %s' already exists.", pluginExecutable, pluginPath, pluginCmd.Name))
		}
	}

	err = fileutils.CopyFile(pluginExecutable, pluginPath)
	if err != nil {
		cmd.ui.Failed(fmt.Sprintf(T("Could not copy plugin binary: \n{{.Error}}", map[string]interface{}{"Error": err.Error()})))
	}

	cmd.config.SetPlugin(pluginMetadata.Name, pluginExecutable)
	cmd.ui.Ok()
	cmd.ui.Say(fmt.Sprintf(T("Plugin {{.PluginName}} successfully installed.", map[string]interface{}{"PluginName": pluginMetadata.Name})))
}

func (cmd *PluginInstall) ensurePluginDoesNotExist(pluginExecutable, pluginExecutableName string) {
	_, err := os.Stat(pluginExecutable)
	if err == nil || os.IsExist(err) {
		cmd.ui.Failed(fmt.Sprintf(T("The file {{.PluginExecutableName}} already exists under the plugin directory.\n",
			map[string]interface{}{
				"PluginExecutableName": pluginExecutableName,
			})))
	} else if !os.IsNotExist(err) {
		cmd.ui.Failed(fmt.Sprintf(T("Unexpected error has occurred:\n{{.Error}}", map[string]interface{}{"Error": err.Error()})))
	}
}

func (cmd *PluginInstall) validateCandidatePluginPath(pluginPath string) {
	_, err := os.Stat(pluginPath)
	if err != nil && os.IsNotExist(err) {
		cmd.ui.Failed(fmt.Sprintf("Binary file '%s' not found", pluginPath))
	}
}

func (cmd *PluginInstall) runPluginBinary(location string, servicePort string) {
	pluginInvocation := exec.Command(location, obtainPort(), servicePort, "install-plugin")

	err := pluginInvocation.Run()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
}

func obtainPort() string {
	//assign 0 to port to get a random open port
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	l.Close()
	return port
}
