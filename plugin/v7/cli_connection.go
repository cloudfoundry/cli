// +build V7

package v7

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"

	plugin_models "code.cloudfoundry.org/cli/plugin/v7/models"
)

type cliConnection struct {
	cliServerPort string
}

func NewCliConnection(cliServerPort string) *cliConnection {
	return &cliConnection{
		cliServerPort: cliServerPort,
	}
}

func (c *cliConnection) GetApp(appName string) (plugin_models.Application, error) {
	var result plugin_models.Application

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetApp", appName, &result)
	})

	return result, err
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

func (c *cliConnection) withClientDo(f func(client *rpc.Client) error) error {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+c.cliServerPort)
	if err != nil {
		return err
	}
	defer client.Close()

	return f(client)
}
