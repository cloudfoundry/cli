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
	cmd, err := runPluginServer(location, port)
	if err != nil {
		return []plugin.Command{}, err
	}
	defer stopPluginServer(cmd)

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
	cmd, err := runPluginServer(location, port)
	if err != nil {
		return false, err
	}
	defer stopPluginServer(cmd)

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
	for _, location := range pluginList {
		cmd, err := runPluginServer(location, port)
		if err != nil {
			continue
		}

		exists, _ = runClientCmd("CmdExists", cmdName, port)

		if exists {
			_, err = runClientCmd("Run", cmdName, port)
			stopPluginServer(cmd)
			return true, err
		}
		stopPluginServer(cmd)
	}
	return false, nil
}

func GetAllPluginCommands() []plugin.Command {
	var combinedCmdList, cmdList []plugin.Command

	pluginsConfig := plugin_config.NewPluginConfig(func(err error) { panic(err) })
	pluginList := pluginsConfig.Plugins()
	port := obtainPort()
	for _, location := range pluginList {
		cmd, err := runPluginServer(location, port)
		if err != nil {
			continue
		}

		cmdList, err = getPluginCmds(port)

		if err == nil {
			combinedCmdList = append(combinedCmdList, cmdList...)
		}
		stopPluginServer(cmd)
	}
	return combinedCmdList
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

func runPluginServer(location string, pluginPort string) (*exec.Cmd, error) {
	//wait for plugin to respond it is active
	address, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cmd := exec.Command(location, pluginPort, strconv.Itoa(listener.Addr().(*net.TCPAddr).Port))
	cmd.Stdout = os.Stdout
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	listener.SetDeadline(time.Now().Add(1 * time.Second))
	_, err = listener.Accept()
	listener.Close()

	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func stopPluginServer(plugin *exec.Cmd) {
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
