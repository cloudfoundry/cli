package rpc

import (
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"time"

	"github.com/cloudfoundry/cli/plugin"
)

func RunListCmd(location string) []plugin.Command {
	cmd := runPluginServer(location)
	rpcClient := dialClient()
	var cmdList []plugin.Command
	err := rpcClient.Call("CliPlugin.ListCmds", "", &cmdList)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	stopPluginServer(cmd)

	return cmdList
}

func RunCommandExists(methodName string, location string) bool {
	cmd := runPluginServer(location)
	rpcClient := dialClient()
	var exist bool
	err := rpcClient.Call("CliPlugin.CmdExists", methodName, &exist)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	stopPluginServer(cmd)

	return exist
}

func RunMethodIfExists(cmdName string, pluginList map[string]string) bool {
	var exists bool
	for _, location := range pluginList {
		cmd := runPluginServer(location)
		exists = runClientCmd("CmdExists", cmdName)
		if exists {
			runClientCmd("Run", cmdName)
			stopPluginServer(cmd)
			return true
		}
		stopPluginServer(cmd)
	}
	return false
}

func runClientCmd(cmd string, method string) bool {
	client := dialClient()
	var reply bool
	err := client.Call("CliPlugin."+cmd, method, &reply)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	return reply
}

func dialClient() *rpc.Client {
	client, err := rpc.Dial("tcp", "127.0.0.1:20001")
	if err != nil {
		os.Exit(1)
	}
	return client
}

func runPluginServer(location string) *exec.Cmd {
	cmd := exec.Command(location)
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		fmt.Println("Error running plugin command: ", err)
		os.Exit(1)
	}

	time.Sleep(300 * time.Millisecond)
	return cmd
}

func stopPluginServer(plugin *exec.Cmd) {
	plugin.Process.Kill()
	plugin.Wait()
}
