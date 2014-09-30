package plugin

import (
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"time"

	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type Plugins struct {
	ui     terminal.UI
	config configuration.ReadWriter
}

func NewPlugins(ui terminal.UI, config configuration.ReadWriter) *Plugins {
	return &Plugins{
		ui:     ui,
		config: config,
	}
}

func (cmd *Plugins) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "plugins",
		Description: "list all available plugin commands",
		Usage:       "CF_NAME plugins",
	}
}

func (cmd *Plugins) GetRequirements(_ requirements.Factory, _ *cli.Context) (req []requirements.Requirement, err error) {
	return
}

func (cmd *Plugins) Run(c *cli.Context) {
	cmd.ui.Say("Listing Installed Plugins...")
	plugins := cmd.config.Plugins()

	table := terminal.NewTable(cmd.ui, []string{"Plugin name", "Command name"})

	for pluginName, location := range plugins {
		process := cmd.runPluginServer(location)
		cmdList := cmd.runClientCmd("ListCmds")
		stopPluginServer(process)

		for _, commandName := range cmdList {
			table.Add(pluginName, commandName)
		}
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table.Print()
}

func (cmd *Plugins) runClientCmd(processName string) []string {
	client, err := rpc.Dial("tcp", "127.0.0.1:20001")
	if err != nil {
		cmd.ui.Failed(fmt.Sprintf("Error dialing to plugin %s: %s\n", processName, err.Error()))
		os.Exit(1)
	}

	var cmdList []string
	err = client.Call("CliPlugin."+processName, "", &cmdList)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	return cmdList
}

func (cmd *Plugins) runPluginServer(location string) *exec.Cmd {
	process := exec.Command(location)
	process.Stdout = os.Stdout
	err := process.Start()
	if err != nil {
		cmd.ui.Failed("Error starting plugin: ", err.Error())
	}

	time.Sleep(300 * time.Millisecond)
	return process
}

func stopPluginServer(plugin *exec.Cmd) {
	plugin.Process.Kill()
	plugin.Wait()
}
