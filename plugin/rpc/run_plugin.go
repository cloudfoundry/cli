package rpc

import (
	"os"
	"os/exec"

	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

func RunMethodIfExists(coreCommandRunner *cli.App, args []string, outputCapture terminal.OutputCapture, terminalOutputSwitch terminal.TerminalOutputSwitch) bool {
	pluginsConfig := plugin_config.NewPluginConfig(func(err error) { panic(err) })
	pluginList := pluginsConfig.Plugins()
	for _, metadata := range pluginList {
		for _, command := range metadata.Commands {
			if command.Name == args[0] {
				cliServer, err := startCliServer(coreCommandRunner, outputCapture, terminalOutputSwitch)
				if err != nil {
					os.Exit(1)
				}

				defer cliServer.Stop()
				pluginArgs := append([]string{cliServer.Port()}, args...)
				cmd := exec.Command(metadata.Location, pluginArgs...)
				cmd.Stdout = os.Stdout
				cmd.Stdin = os.Stdin

				defer stopPlugin(cmd)
				err = cmd.Run()
				if err != nil {
					os.Exit(1)
				}
				return true
			}
		}
	}
	return false
}

func startCliServer(coreCommandRunner *cli.App, outputCapture terminal.OutputCapture, terminalOutputSwitch terminal.TerminalOutputSwitch) (*CliRpcService, error) {
	cliServer, err := NewRpcService(coreCommandRunner, outputCapture, terminalOutputSwitch)
	if err != nil {
		return nil, err
	}

	err = cliServer.Start()
	if err != nil {
		return nil, err
	}

	return cliServer, nil
}

func stopPlugin(plugin *exec.Cmd) {
	plugin.Process.Kill()
	plugin.Wait()
}
