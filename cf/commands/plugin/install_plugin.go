package plugin

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/fileutils"
	"github.com/cloudfoundry/cli/plugin/rpc"
	"github.com/codegangsta/cli"
)

type PluginInstall struct {
	ui     terminal.UI
	config plugin_config.PluginConfiguration
}

func NewPluginInstall(ui terminal.UI, config plugin_config.PluginConfiguration) *PluginInstall {
	return &PluginInstall{
		ui:     ui,
		config: config,
	}
}

func (cmd *PluginInstall) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "install-plugin",
		Description: T("PATH/TO/PLUGIN-NAME  - Install the plugin defined in command argument"),
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
	_, pluginExecutableName := filepath.Split(pluginPath)

	cmd.ui.Say(fmt.Sprintf(T("Installing plugin {{.PluginPath}}...", map[string]interface{}{"PluginPath": pluginPath})))

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

	runPluginBinary(pluginPath, rpcService.Port())

	pluginName := rpcService.RpcCmd.ReturnData.(string)
	if pluginName == "" {
		cmd.ui.Failed(fmt.Sprintf("Unable to obtain plugin name for executable %s", pluginPath))
	}

	if _, ok := plugins[pluginName]; ok {
		cmd.ui.Failed(fmt.Sprintf(T("Plugin name {{.PluginName}} is already taken", map[string]interface{}{"PluginName": pluginName})))
	}

	err = fileutils.CopyFile(pluginExecutable, pluginPath)
	if err != nil {
		cmd.ui.Failed(fmt.Sprintf(T("Could not copy plugin binary: \n{{.Error}}", map[string]interface{}{"Error": err.Error()})))
	}

	cmd.config.SetPlugin(pluginName, pluginExecutable)
	cmd.ui.Ok()
	cmd.ui.Say(fmt.Sprintf(T("Plugin {{.PluginName}} successfully installed.", map[string]interface{}{"PluginName": pluginName})))
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

func runPluginBinary(location string, servicePort string) error {
	cmd := exec.Command(location, obtainPort(), servicePort, "install-plugin")
	err := cmd.Run()
	if err != nil {
		panic(err.Error())
	}
	cmd.Wait()
	return err
}

func obtainPort() string {
	//assign 0 to port to get a random open port
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	l.Close()
	return port
}
