package plugin

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"

	"code.cloudfoundry.org/cli/plugin/models"
)

type cliConnection struct {
	cliServerPort string
}

func NewCliConnection(cliServerPort string) *cliConnection {
	return &cliConnection{
		cliServerPort: cliServerPort,
	}
}

func (c *cliConnection) withClientDo(f func(client *rpc.Client) error) error {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+c.cliServerPort)
	if err != nil {
		return err
	}
	defer client.Close()

	return f(client)
}

func (c *cliConnection) sendPluginMetadataToCliServer(metadata PluginMetadata) {
	var success bool

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.SetPluginMetadata", metadata, &success)
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !success {
		os.Exit(1)
	}

	os.Exit(0)
}

func (c *cliConnection) isMinCliVersion(version string) bool {
	var result bool

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.IsMinCliVersion", version, &result)
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return result
}

func (c *cliConnection) CliCommandWithoutTerminalOutput(args ...string) ([]string, error) {
	return c.callCliCommand(true, args...)
}

func (c *cliConnection) CliCommand(args ...string) ([]string, error) {
	return c.callCliCommand(false, args...)
}

func (c *cliConnection) callCliCommand(silently bool, args ...string) ([]string, error) {
	var (
		success                  bool
		cmdOutput                []string
		callCoreCommandErr       error
		getOutputAndResetErr     error
		disableTerminalOutputErr error
	)

	c.withClientDo(func(client *rpc.Client) error {
		disableTerminalOutputErr = client.Call("CliRpcCmd.DisableTerminalOutput", silently, &success)
		callCoreCommandErr = client.Call("CliRpcCmd.CallCoreCommand", args, &success)
		getOutputAndResetErr = client.Call("CliRpcCmd.GetOutputAndReset", success, &cmdOutput)

		return nil
	})

	if disableTerminalOutputErr != nil {
		return []string{}, disableTerminalOutputErr
	}

	if callCoreCommandErr != nil {
		return []string{}, callCoreCommandErr
	}

	if !success {
		return []string{}, errors.New("Error executing cli core command")
	}

	if getOutputAndResetErr != nil {
		return []string{}, errors.New("something completely unexpected happened")
	}

	return cmdOutput, nil
}

func (c *cliConnection) pingCLI() {
	//call back to cf saying we have been setup
	var connErr error
	var conn net.Conn
	for i := 0; i < 5; i++ {
		conn, connErr = net.Dial("tcp", "127.0.0.1:"+c.cliServerPort)
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

func (c *cliConnection) GetCurrentOrg() (plugin_models.Organization, error) {
	var result plugin_models.Organization

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetCurrentOrg", "", &result)
	})

	return result, err
}

func (c *cliConnection) GetCurrentSpace() (plugin_models.Space, error) {
	var result plugin_models.Space

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetCurrentSpace", "", &result)
	})

	return result, err
}

func (c *cliConnection) Username() (string, error) {
	var result string

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.Username", "", &result)
	})

	return result, err
}

func (c *cliConnection) UserGuid() (string, error) {
	var result string

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.UserGuid", "", &result)
	})

	return result, err
}

func (c *cliConnection) UserEmail() (string, error) {
	var result string

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.UserEmail", "", &result)
	})

	return result, err
}

func (c *cliConnection) IsSSLDisabled() (bool, error) {
	var result bool

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.IsSSLDisabled", "", &result)
	})

	return result, err
}

func (c *cliConnection) IsLoggedIn() (bool, error) {
	var result bool

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.IsLoggedIn", "", &result)
	})

	return result, err
}

func (c *cliConnection) HasOrganization() (bool, error) {
	var result bool

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.HasOrganization", "", &result)
	})

	return result, err
}

func (c *cliConnection) HasSpace() (bool, error) {
	var result bool

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.HasSpace", "", &result)
	})

	return result, err
}

func (c *cliConnection) ApiEndpoint() (string, error) {
	var result string

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.ApiEndpoint", "", &result)
	})

	return result, err
}

func (c *cliConnection) HasAPIEndpoint() (bool, error) {
	var result bool

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.HasAPIEndpoint", "", &result)
	})

	return result, err
}

func (c *cliConnection) ApiVersion() (string, error) {
	var result string

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.ApiVersion", "", &result)
	})

	return result, err
}

func (c *cliConnection) LoggregatorEndpoint() (string, error) {
	var result string

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.LoggregatorEndpoint", "", &result)
	})

	return result, err
}

func (c *cliConnection) DopplerEndpoint() (string, error) {
	var result string

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.DopplerEndpoint", "", &result)
	})

	return result, err
}

func (c *cliConnection) AccessToken() (string, error) {
	var result string

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.AccessToken", "", &result)
	})

	return result, err
}

func (c *cliConnection) GetApp(appName string) (plugin_models.GetAppModel, error) {
	var result plugin_models.GetAppModel

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetApp", appName, &result)
	})

	return result, err
}

func (c *cliConnection) GetApps() ([]plugin_models.GetAppsModel, error) {
	var result []plugin_models.GetAppsModel

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetApps", "", &result)
	})

	return result, err
}

func (c *cliConnection) GetOrgs() ([]plugin_models.GetOrgs_Model, error) {
	var result []plugin_models.GetOrgs_Model

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetOrgs", "", &result)
	})

	return result, err
}

func (c *cliConnection) GetSpaces() ([]plugin_models.GetSpaces_Model, error) {
	var result []plugin_models.GetSpaces_Model

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetSpaces", "", &result)
	})

	return result, err
}

func (c *cliConnection) GetServices() ([]plugin_models.GetServices_Model, error) {
	var result []plugin_models.GetServices_Model

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetServices", "", &result)
	})

	return result, err
}

func (c *cliConnection) GetOrgUsers(orgName string, args ...string) ([]plugin_models.GetOrgUsers_Model, error) {
	var result []plugin_models.GetOrgUsers_Model

	cmdArgs := append([]string{orgName}, args...)

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetOrgUsers", cmdArgs, &result)
	})

	return result, err
}

func (c *cliConnection) GetSpaceUsers(orgName string, spaceName string) ([]plugin_models.GetSpaceUsers_Model, error) {
	var result []plugin_models.GetSpaceUsers_Model

	cmdArgs := []string{orgName, spaceName}

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetSpaceUsers", cmdArgs, &result)
	})

	return result, err
}

func (c *cliConnection) GetOrg(orgName string) (plugin_models.GetOrg_Model, error) {
	var result plugin_models.GetOrg_Model

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetOrg", orgName, &result)
	})

	return result, err
}

func (c *cliConnection) GetSpace(spaceName string) (plugin_models.GetSpace_Model, error) {
	var result plugin_models.GetSpace_Model

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetSpace", spaceName, &result)
	})

	return result, err
}

func (c *cliConnection) GetService(serviceInstance string) (plugin_models.GetService_Model, error) {
	var result plugin_models.GetService_Model

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetService", serviceInstance, &result)
	})

	return result, err
}
