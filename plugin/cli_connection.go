package plugin

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"

	"github.com/cloudfoundry/cli/plugin/models"
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
	defer cliServerConn.Close()

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
	defer client.Close()

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

func (cliConnection *cliConnection) GetCurrentOrg() (plugin_models.Organization, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return plugin_models.Organization{}, err
	}

	var result plugin_models.Organization

	err = client.Call("CliRpcCmd.GetCurrentOrg", "", &result)
	return result, err
}

func (cliConnection *cliConnection) GetCurrentSpace() (plugin_models.Space, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return plugin_models.Space{}, err
	}

	var result plugin_models.Space

	err = client.Call("CliRpcCmd.GetCurrentSpace", "", &result)
	return result, err
}

func (cliConnection *cliConnection) Username() (string, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return "", err
	}

	var result string

	err = client.Call("CliRpcCmd.Username", "", &result)
	return result, err
}

func (cliConnection *cliConnection) UserGuid() (string, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return "", err
	}

	var result string

	err = client.Call("CliRpcCmd.UserGuid", "", &result)
	return result, err
}

func (cliConnection *cliConnection) UserEmail() (string, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return "", err
	}

	var result string

	err = client.Call("CliRpcCmd.UserEmail", "", &result)
	return result, err
}

func (cliConnection *cliConnection) IsSSLDisabled() (bool, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return false, err
	}

	var result bool

	err = client.Call("CliRpcCmd.IsSSLDisabled", "", &result)
	return result, err
}

func (cliConnection *cliConnection) IsLoggedIn() (bool, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return false, err
	}

	var result bool

	err = client.Call("CliRpcCmd.IsLoggedIn", "", &result)
	return result, err
}

func (cliConnection *cliConnection) HasOrganization() (bool, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return false, err
	}

	var result bool

	err = client.Call("CliRpcCmd.HasOrganization", "", &result)
	return result, err
}

func (cliConnection *cliConnection) HasSpace() (bool, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return false, err
	}

	var result bool

	err = client.Call("CliRpcCmd.HasSpace", "", &result)
	return result, err
}

func (cliConnection *cliConnection) ApiEndpoint() (string, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return "", err
	}

	var result string

	err = client.Call("CliRpcCmd.ApiEndpoint", "", &result)
	return result, err
}

func (cliConnection *cliConnection) HasAPIEndpoint() (bool, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return false, err
	}

	var result bool

	err = client.Call("CliRpcCmd.HasApiEndpoint", "", &result)
	return result, err
}

func (cliConnection *cliConnection) ApiVersion() (string, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return "", err
	}

	var result string

	err = client.Call("CliRpcCmd.ApiVersion", "", &result)
	return result, err
}

func (cliConnection *cliConnection) LoggregatorEndpoint() (string, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return "", err
	}

	var result string

	err = client.Call("CliRpcCmd.LoggregatorEndpoint", "", &result)
	return result, err
}

func (cliConnection *cliConnection) DopplerEndpoint() (string, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return "", err
	}

	var result string

	err = client.Call("CliRpcCmd.DopplerEndpoint", "", &result)
	return result, err
}

func (cliConnection *cliConnection) AccessToken() (string, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return "", err
	}

	var result string

	err = client.Call("CliRpcCmd.AccessToken", "", &result)
	return result, err
}
