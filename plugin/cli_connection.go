package plugin

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"
)

type cliConnection struct {
	cliServerPort string
}

func NewCliConnection(cliServerPort string) *cliConnection {
	return &cliConnection{
		cliServerPort: cliServerPort,
	}
}

func (cliConnection *cliConnection) sendPluginMetadataToCliServer(metadata PluginMetadata) {
	cliServerConn, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var success bool

	err = cliServerConn.Call("CliRpcCmd.SetPluginMetadata", metadata, &success)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !success {
		os.Exit(1)
	}

	os.Exit(0)
}

func (cliConnection *cliConnection) CliCommandWithoutTerminalOutput(args ...string) ([]string, error) {
	return cliConnection.callCliCommand(true, args...)
}

func (cliConnection *cliConnection) CliCommand(args ...string) ([]string, error) {
	return cliConnection.callCliCommand(false, args...)
}

func (cliConnection *cliConnection) callCliCommand(silently bool, args ...string) ([]string, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
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

func (cliConnection *cliConnection) pingCLI() {
	//call back to cf saying we have been setup
	var connErr error
	var conn net.Conn
	for i := 0; i < 5; i++ {
		conn, connErr = net.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
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
