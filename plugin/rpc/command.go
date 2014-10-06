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
	cmd, err := runPluginServer(location)
	if err != nil {
		return []plugin.Command{}, err
	}
	defer stopPluginServer(cmd)

	rpcClient, err := dialClient()
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
	cmd, err := runPluginServer(location)
	if err != nil {
		return false, err
	}
	defer stopPluginServer(cmd)

	rpcClient, err := dialClient()
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

	for _, location := range pluginList {
		cmd, err := runPluginServer(location)
		if err != nil {
			continue
		}

		exists, _ = runClientCmd("CmdExists", cmdName)

		if exists {
			_, err = runClientCmd("Run", cmdName)
			stopPluginServer(cmd)
			return true, err
		}
		stopPluginServer(cmd)
	}
	return false, nil
}

func GetAllPluginCommands() []plugin.Command {
	pluginsConfig := plugin_config.NewPluginConfig(func(err error) { panic(err) })
	pluginList := pluginsConfig.Plugins()

	var combinedCmdList, cmdList []plugin.Command

	for _, location := range pluginList {
		cmd, err := runPluginServer(location)
		if err != nil {
			continue
		}

		cmdList, err = getPluginCmds()

		if err == nil {
			combinedCmdList = append(combinedCmdList, cmdList...)
		}
		stopPluginServer(cmd)
	}
	return combinedCmdList
}

func runClientCmd(cmd string, method string) (bool, error) {
	client, err := dialClient()
	var reply bool
	err = client.Call("CliPlugin."+cmd, method, &reply)
	if err != nil {
		return false, err
	}
	return reply, nil
}

func getPluginCmds() ([]plugin.Command, error) {
	client, err := dialClient()
	var cmds []plugin.Command
	err = client.Call("CliPlugin.ListCmds", "", &cmds)
	if err != nil {
		return nil, err
	}
	return cmds, nil
}

func dialClient() (*rpc.Client, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:20001")
	if err != nil {
		return nil, err
	}
	return client, nil
}

func runPluginServer(location string) (*exec.Cmd, error) {
	cmd := exec.Command(location)
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
