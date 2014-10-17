package rpc

import (
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
)

func RunMethodIfExists(cmdName string) (bool, error) {
	pluginsConfig := plugin_config.NewPluginConfig(func(err error) { panic(err) })
	pluginList := pluginsConfig.Plugins()
	port := obtainPort()

	service, err := startCliServer()
	if err != nil {
		return false, err
	}

	defer service.Stop()

	for pluginName, metadata := range pluginList {
		for _, command := range metadata.Commands {
			if command.Name == cmdName {
				cmd, err := runPlugin(metadata.Location, port, service.Port())
				if err != nil {
					continue
				}

				_, err = runClientCmd(pluginName+".Run", cmdName, port)
				stopPlugin(cmd)
				return true, err
			}
		}
	}
	return false, nil
}

func runClientCmd(cmd string, method string, port string) (bool, error) {
	client, err := dialClient(port)
	var reply bool
	err = client.Call(cmd, method, &reply)
	if err != nil {
		return false, err
	}
	return reply, nil
}

func dialClient(port string) (*rpc.Client, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+port)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func startCliServer() (*CliRpcService, error) {
	service, err := NewRpcService()
	if err != nil {
		return nil, err
	}

	err = service.Start()
	if err != nil {
		return nil, err
	}

	return service, nil
}

func runPlugin(location string, pluginPort string, servicePort string) (*exec.Cmd, error) {
	cmd := exec.Command(location, pluginPort, servicePort)
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	err = pingPlugin(pluginPort)
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

func pingPlugin(port string) error {
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
