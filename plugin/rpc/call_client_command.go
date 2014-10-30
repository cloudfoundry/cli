package rpc

import (
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

func RunMethodIfExists(coreCommandRunner *cli.App, args []string, outputCapture terminal.OutputCapture) (bool, error) {
	pluginsConfig := plugin_config.NewPluginConfig(func(err error) { panic(err) })
	pluginList := pluginsConfig.Plugins()

	for pluginName, metadata := range pluginList {
		for _, command := range metadata.Commands {
			if command.Name == args[0] {
				cliServer, err := startCliServer(coreCommandRunner, outputCapture)
				if err != nil {
					return false, err
				}

				defer cliServer.Stop()

				pluginPort := obtainPort()
				pluginProcess, err := runPlugin(metadata.Location, pluginPort, cliServer.Port())
				if err != nil {
					continue
				}

				defer stopPlugin(pluginProcess)

				_, err = runClientCmd(pluginName, pluginPort, args)
				return true, err
			}
		}
	}
	return false, nil
}

func runClientCmd(pluginName string, pluginPort string, args []string) (bool, error) {
	client, err := dialServer(pluginPort)
	var reply bool
	err = client.Call(pluginName+".Run", args, &reply)
	if err != nil {
		return false, err
	}
	return reply, nil
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

	err = waitForPluginToRespondToRpc(pluginPort)
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

func waitForPluginToRespondToRpc(port string) error {
	var err error
	var conn net.Conn
	for i := 0; i < 5; i++ {
		conn, err = net.Dial("tcp", "127.0.0.1:"+port)
		if err != nil {
			time.Sleep(200 * time.Millisecond)
		} else {
			conn.Close()
			return err
		}
	}
	return err
}
