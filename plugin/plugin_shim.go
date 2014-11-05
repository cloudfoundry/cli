package plugin

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"
)

var CliServicePort string

/**
	* This function is called by the plugin to setup their server. This allows us to call Run on the plugin
	* os.Args[1] port CF_CLI rpc server is running on
	* os.Args[2] **OPTIONAL**
		* SendMetadata - used to fetch the plugin metadata
**/
func Start(cmd Plugin) {
	pingCLI()
	if isMetadataRequest() {
		sendPluginMetadataToCliServer(cmd)
	} else {
		cmd.Run(os.Args[2:])
	}
}

func isMetadataRequest() bool {
	return len(os.Args) == 3 && os.Args[2] == "SendMetadata"
}

func sendPluginMetadataToCliServer(cmd Plugin) {
	cliServerConn, err := rpc.Dial("tcp", "127.0.0.1:"+os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var success bool

	err = cliServerConn.Call("CliRpcCmd.SetPluginMetadata", cmd.GetMetadata(), &success)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !success {
		os.Exit(1)
	}

	os.Exit(0)

}

func CliCommandWithoutTerminalOutput(args ...string) ([]string, error) {
	return callCoreCommand(true, args...)
}

func CliCommand(args ...string) ([]string, error) {
	return callCoreCommand(false, args...)
}

func callCoreCommand(silently bool, args ...string) ([]string, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+CliServicePort)
	if err != nil {
		return []string{}, err
	}

	var success bool

	client.Call("CliRpcCmd.DisableTerminalOutput", silently, &success)
	err = client.Call("CliRpcCmd.CallCoreCommand", args, &success)

	var cmdOutput []string
	outputErr := client.Call("CliRpcCmd.GetOutputAndReset", success, &cmdOutput)

	if err != nil {
		return cmdOutput, err
	} else if !success {
		return cmdOutput, errors.New("Error executing cli core command")
	}

	if outputErr != nil {
		return cmdOutput, errors.New("something completely unexpected happened")
	}
	return cmdOutput, nil
}

func pingCLI() {
	//call back to cf saying we have been setup
	var connErr error
	var conn net.Conn
	for i := 0; i < 5; i++ {
		CliServicePort = os.Args[1]
		conn, connErr = net.Dial("tcp", "127.0.0.1:"+CliServicePort)
		if connErr != nil {
			time.Sleep(200 * time.Millisecond)
		} else {
			conn.Close()
			break
		}
	}
	if connErr != nil {
		fmt.Println(connErr)
		os.Exit(1)
	}
}
