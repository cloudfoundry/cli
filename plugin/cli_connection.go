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

func (cliConnection *cliConnection) GetCurrentOrg() (plugin_models.OrganizationSummary, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return plugin_models.OrganizationSummary{}, err
	}

	var result plugin_models.OrganizationSummary

	err = client.Call("CliRpcCmd.GetCurrentOrg", "", &result)
	return result, err
}

func (cliConnection *cliConnection) GetCurrentSpace() (plugin_models.SpaceSummary, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return plugin_models.SpaceSummary{}, err
	}

	var result plugin_models.SpaceSummary

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

func (cliConnection *cliConnection) GetApp(appName string) (plugin_models.Application, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return plugin_models.Application{}, err
	}

	var result plugin_models.Application

	err = client.Call("CliRpcCmd.GetApp", appName, &result)
	return result, err
}

func (cliConnection *cliConnection) GetApps() ([]plugin_models.ApplicationSummary, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return []plugin_models.ApplicationSummary{}, err
	}

	var result []plugin_models.ApplicationSummary

	err = client.Call("CliRpcCmd.GetApps", "", &result)
	return result, err
}

func (cliConnection *cliConnection) GetOrgs() ([]plugin_models.OrganizationSummary, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return []plugin_models.OrganizationSummary{}, err
	}

	var result []plugin_models.OrganizationSummary

	err = client.Call("CliRpcCmd.GetOrgs", "", &result)
	return result, err
}

func (cliConnection *cliConnection) GetSpaces() ([]plugin_models.SpaceSummary, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return []plugin_models.SpaceSummary{}, err
	}

	var result []plugin_models.SpaceSummary

	err = client.Call("CliRpcCmd.GetSpaces", "", &result)
	return result, err
}

func (cliConnection *cliConnection) GetServices() ([]plugin_models.ServiceInstance, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return []plugin_models.ServiceInstance{}, err
	}

	var result []plugin_models.ServiceInstance

	err = client.Call("CliRpcCmd.GetServices", "", &result)
	return result, err
}

func (cliConnection *cliConnection) GetOrgUsers(orgName string, args ...string) ([]plugin_models.User, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return []plugin_models.User{}, err
	}

	var result []plugin_models.User

	cmdArgs := append([]string{orgName}, args...)

	err = client.Call("CliRpcCmd.GetOrgUsers", cmdArgs, &result)
	return result, err
}

func (cliConnection *cliConnection) GetSpaceUsers(orgName string, spaceName string) ([]plugin_models.User, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return []plugin_models.User{}, err
	}

	var result []plugin_models.User

	cmdArgs := []string{orgName, spaceName}

	err = client.Call("CliRpcCmd.GetSpaceUsers", cmdArgs, &result)
	return result, err
}

func (cliConnection *cliConnection) GetOrg(orgName string) (plugin_models.Organization, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return plugin_models.Organization{}, err
	}

	var result plugin_models.Organization

	err = client.Call("CliRpcCmd.GetOrg", orgName, &result)
	return result, err
}

func (cliConnection *cliConnection) GetSpace(spaceName string) (plugin_models.SpaceDetails, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+cliConnection.cliServerPort)
	if err != nil {
		return plugin_models.SpaceDetails{}, err
	}

	var result plugin_models.SpaceDetails

	err = client.Call("CliRpcCmd.GetSpace", spaceName, &result)
	return result, err
}
