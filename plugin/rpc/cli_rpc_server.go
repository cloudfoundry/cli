package rpc

import (
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/cli/plugin/models"
	"github.com/codegangsta/cli"

	"fmt"
	"net"
	"net/rpc"
	"strconv"
)

type CliRpcService struct {
	listener net.Listener
	stopCh   chan struct{}
	Pinged   bool
	RpcCmd   *CliRpcCmd
}

type CliRpcCmd struct {
	PluginMetadata       *plugin.PluginMetadata
	coreCommandRunner    *cli.App
	outputCapture        terminal.OutputCapture
	terminalOutputSwitch terminal.TerminalOutputSwitch
	cliConfig            core_config.Repository
}

func NewRpcService(commandRunner *cli.App, outputCapture terminal.OutputCapture, terminalOutputSwitch terminal.TerminalOutputSwitch, cliConfig core_config.Repository) (*CliRpcService, error) {
	rpcService := &CliRpcService{
		RpcCmd: &CliRpcCmd{
			PluginMetadata:       &plugin.PluginMetadata{},
			coreCommandRunner:    commandRunner,
			outputCapture:        outputCapture,
			terminalOutputSwitch: terminalOutputSwitch,
			cliConfig:            cliConfig,
		},
	}

	err := rpc.Register(rpcService.RpcCmd)
	if err != nil {
		return nil, err
	}

	return rpcService, nil
}

func (cli *CliRpcService) Stop() {
	close(cli.stopCh)
	cli.listener.Close()
}

func (cli *CliRpcService) Port() string {
	return strconv.Itoa(cli.listener.Addr().(*net.TCPAddr).Port)
}

func (cli *CliRpcService) Start() error {
	var err error

	cli.stopCh = make(chan struct{})

	cli.listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := cli.listener.Accept()
			if err != nil {
				select {
				case <-cli.stopCh:
					return
				default:
					fmt.Println(err)
				}
			} else {
				go rpc.ServeConn(conn)
			}
		}
	}()

	return nil
}

func (cmd *CliRpcService) SetTheApp(app *cli.App) {
	cmd.RpcCmd.coreCommandRunner = app
}

func (cmd *CliRpcCmd) SetPluginMetadata(pluginMetadata plugin.PluginMetadata, retVal *bool) error {
	cmd.PluginMetadata = &pluginMetadata
	*retVal = true
	return nil
}

func (cmd *CliRpcCmd) DisableTerminalOutput(disable bool, retVal *bool) error {
	cmd.terminalOutputSwitch.DisableTerminalOutput(disable)
	*retVal = true
	return nil
}

func (cmd *CliRpcCmd) CallCoreCommand(args []string, retVal *bool) error {
	defer func() {
		recover()
	}()

	err := cmd.coreCommandRunner.Run(append([]string{"CF_NAME"}, args...))

	if err != nil {
		*retVal = false
		return err
	}

	*retVal = true
	return nil
}

func (cmd *CliRpcCmd) GetOutputAndReset(args bool, retVal *[]string) error {
	*retVal = cmd.outputCapture.GetOutputAndReset()
	return nil
}

func (cmd *CliRpcCmd) GetCurrentOrg(args string, retVal *plugin_models.Organization) error {
	retVal.Name = cmd.cliConfig.OrganizationFields().Name
	retVal.Guid = cmd.cliConfig.OrganizationFields().Guid
	retVal.QuotaDefinition.Guid = cmd.cliConfig.OrganizationFields().QuotaDefinition.Guid
	retVal.QuotaDefinition.Name = cmd.cliConfig.OrganizationFields().QuotaDefinition.Name
	retVal.QuotaDefinition.MemoryLimit = cmd.cliConfig.OrganizationFields().QuotaDefinition.MemoryLimit
	retVal.QuotaDefinition.InstanceMemoryLimit = cmd.cliConfig.OrganizationFields().QuotaDefinition.InstanceMemoryLimit
	retVal.QuotaDefinition.RoutesLimit = cmd.cliConfig.OrganizationFields().QuotaDefinition.RoutesLimit
	retVal.QuotaDefinition.ServicesLimit = cmd.cliConfig.OrganizationFields().QuotaDefinition.ServicesLimit
	retVal.QuotaDefinition.NonBasicServicesAllowed = cmd.cliConfig.OrganizationFields().QuotaDefinition.NonBasicServicesAllowed

	return nil
}

func (cmd *CliRpcCmd) GetCurrentSpace(args string, retVal *plugin_models.Space) error {
	retVal.Name = cmd.cliConfig.SpaceFields().Name
	retVal.Guid = cmd.cliConfig.SpaceFields().Guid

	return nil
}

func (cmd *CliRpcCmd) Username(args string, retVal *string) error {
	*retVal = cmd.cliConfig.Username()

	return nil
}

func (cmd *CliRpcCmd) UserGuid(args string, retVal *string) error {
	*retVal = cmd.cliConfig.UserGuid()

	return nil
}

func (cmd *CliRpcCmd) UserEmail(args string, retVal *string) error {
	*retVal = cmd.cliConfig.UserEmail()

	return nil
}

func (cmd *CliRpcCmd) IsLoggedIn(args string, retVal *bool) error {
	*retVal = cmd.cliConfig.IsLoggedIn()

	return nil
}

func (cmd *CliRpcCmd) IsSSLDisabled(args string, retVal *bool) error {
	*retVal = cmd.cliConfig.IsSSLDisabled()

	return nil
}

func (cmd *CliRpcCmd) HasOrganization(args string, retVal *bool) error {
	*retVal = cmd.cliConfig.HasOrganization()

	return nil
}

func (cmd *CliRpcCmd) HasSpace(args string, retVal *bool) error {
	*retVal = cmd.cliConfig.HasSpace()

	return nil
}

func (cmd *CliRpcCmd) ApiEndpoint(args string, retVal *string) error {
	*retVal = cmd.cliConfig.ApiEndpoint()

	return nil
}

func (cmd *CliRpcCmd) HasAPIEndpoint(args string, retVal *bool) error {
	*retVal = cmd.cliConfig.HasAPIEndpoint()

	return nil
}

func (cmd *CliRpcCmd) ApiVersion(args string, retVal *string) error {
	*retVal = cmd.cliConfig.ApiVersion()

	return nil
}

func (cmd *CliRpcCmd) LoggregatorEndpoint(args string, retVal *string) error {
	*retVal = cmd.cliConfig.LoggregatorEndpoint()

	return nil
}

func (cmd *CliRpcCmd) DopplerEndpoint(args string, retVal *string) error {
	*retVal = cmd.cliConfig.DopplerEndpoint()

	return nil
}

func (cmd *CliRpcCmd) AccessToken(args string, retVal *string) error {
	*retVal = cmd.cliConfig.AccessToken()

	return nil
}
