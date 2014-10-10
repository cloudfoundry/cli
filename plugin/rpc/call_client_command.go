package rpc

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	"github.com/cloudfoundry/cli/plugin"
)

func RunListCmd(location string) ([]plugin.Command, error) {
	port := obtainPort()

	service, err := startCliServer()
	if err != nil {
		return []plugin.Command{}, err
	}

	defer service.Stop()

	cmd, err := runPlugin(location, port, service.Port())
	if err != nil {
		return []plugin.Command{}, err
	}
	defer stopPlugin(cmd)

	rpcClient, err := dialClient(port)
	if err != nil {
		return []plugin.Command{}, err
	}

	var cmdList []plugin.Command
	err = rpcClient.Call("CliPlugin.ListCmds", "", &cmdList)
	if err != nil {
		return []plugin.Command{}, err
	}

	return cmdList, nil
}

func RunCommandExists(methodName string, location string) (bool, error) {
	port := obtainPort()

	service, err := startCliServer()
	if err != nil {
		return false, err
	}

	defer service.Stop()

	cmd, err := runPlugin(location, port, service.Port())
	if err != nil {
		return false, err
	}
	defer stopPlugin(cmd)

	rpcClient, err := dialClient(port)
	if err != nil {
		return false, err
	}

	var exist bool
	err = rpcClient.Call("CliPlugin.CmdExists", methodName, &exist)
	if err != nil {
		return false, err
	}

	return exist, nil
}

func RunMethodIfExists(cmdName string) (bool, error) {
	var exists bool
	pluginsConfig := plugin_config.NewPluginConfig(func(err error) { panic(err) })
	pluginList := pluginsConfig.Plugins()
	port := obtainPort()

	service, err := startCliServer()
	if err != nil {
		return false, err
	}

	defer service.Stop()

	for _, location := range pluginList {
		cmd, err := runPlugin(location, port, service.Port())
		if err != nil {
			continue
		}

		exists, _ = runClientCmd("CmdExists", cmdName, port)

		if exists {
			_, err = runClientCmd("Run", cmdName, port)
			stopPlugin(cmd)
			return true, err
		}
		stopPlugin(cmd)
	}
	return false, nil
}

func GetAllPluginCommands() ([]plugin.Command, error) {
	var combinedCmdList, cmdList []plugin.Command

	pluginsConfig := plugin_config.NewPluginConfig(func(err error) { panic(err) })
	pluginList := pluginsConfig.Plugins()

	service, err := startCliServer()
	if err != nil {
		return []plugin.Command{}, err
	}

	defer service.Stop()

	port := obtainPort()
	for _, location := range pluginList {
		cmd, err := runPlugin(location, port, service.Port()) //both started
		if err != nil {
			continue
		}

		cmdList, err = getPluginCmds(port)

		if err == nil {
			combinedCmdList = append(combinedCmdList, cmdList...)
		}
		stopPlugin(cmd)
	}
	return combinedCmdList, nil
}

func runClientCmd(cmd string, method string, port string) (bool, error) {
	client, err := dialClient(port)
	var reply bool
	err = client.Call("CliPlugin."+cmd, method, &reply)
	if err != nil {
		return false, err
	}
	return reply, nil
}

func getPluginCmds(port string) ([]plugin.Command, error) {
	client, err := dialClient(port)
	if err != nil {
		fmt.Println("error dailing to plugin: ", err)
		return []plugin.Command{}, err
	}
	var cmds []plugin.Command
	err = client.Call("CliPlugin.ListCmds", "", &cmds)
	if err != nil {
		return nil, err
	}
	return cmds, nil
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
