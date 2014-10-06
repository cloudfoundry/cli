package rpc

import (
	"net/rpc"
	"os"
	"os/exec"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	"github.com/cloudfoundry/cli/plugin"
)

func RunListCmd(location string) ([]plugin.Command, error) {
	port := "20001"
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
	port := "20001"
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
	port := "20001"
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
	port := "20001"
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

func runPluginServer(location string, port string) (*exec.Cmd, error) {
	cmd := exec.Command(location, port)
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	time.Sleep(300 * time.Millisecond)
	return cmd, nil
}

func stopPluginServer(plugin *exec.Cmd) {
	plugin.Process.Kill()
	plugin.Wait()
}
