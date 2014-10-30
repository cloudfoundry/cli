package rpc

import (
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"strconv"

	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

func RunMethodIfExists(coreCommandRunner *cli.App, args []string, outputCapture terminal.OutputCapture) (bool, error) {
	pluginsConfig := plugin_config.NewPluginConfig(func(err error) { panic(err) })
	pluginList := pluginsConfig.Plugins()
	for _, metadata := range pluginList {
		for _, command := range metadata.Commands {
			if command.Name == args[0] {
				cliServer, err := startCliServer(coreCommandRunner, outputCapture)
				if err != nil {
					return false, err
				}

				defer cliServer.Stop()
				pluginArgs := append([]string{"empty arg", cliServer.Port()}, args...)
				cmd := exec.Command(metadata.Location, pluginArgs...)
				cmd.Stdout = os.Stdout
				cmd.Stdin = os.Stdin

				err = cmd.Run()
				if err != nil {
					return false, err
				}

				defer stopPlugin(cmd)

				return true, err
			}
		}
	}
	return false, nil
}

func dialServer(port string) (*rpc.Client, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+port)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func startCliServer(coreCommandRunner *cli.App, outputCapture terminal.OutputCapture) (*CliRpcService, error) {
	cliServer, err := NewRpcService(coreCommandRunner, outputCapture)
	if err != nil {
		return nil, err
	}

	err = cliServer.Start()
	if err != nil {
		return nil, err
	}

	return cliServer, nil
}

func runPlugin(location string, pluginPort string, servicePort string) (*exec.Cmd, error) {
	cmd := exec.Command(location, pluginPort, servicePort)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func stopPlugin(plugin *exec.Cmd) {
	plugin.Process.Kill()
	plugin.Wait()
}

func obtainPort() string {
	//assign 0 to port to get a random open port
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	l.Close()
	return port
}
