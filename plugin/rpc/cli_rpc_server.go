package rpc

import (
	"os"
	"strings"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/version"
	"github.com/blang/semver"

	"fmt"
	"net"
	"net/rpc"
	"strconv"

	"bytes"
	"io"

	"sync"

	"code.cloudfoundry.org/cli/cf/trace"
)

var dialTimeout = os.Getenv("CF_DIAL_TIMEOUT")

type CliRpcService struct {
	listener net.Listener
	stopCh   chan struct{}
	Pinged   bool
	RpcCmd   *CliRpcCmd
	Server   *rpc.Server
}

type CliRpcCmd struct {
	PluginMetadata       *plugin.PluginMetadata
	MetadataMutex        *sync.RWMutex
	outputCapture        OutputCapture
	terminalOutputSwitch TerminalOutputSwitch
	cliConfig            coreconfig.Repository
	repoLocator          api.RepositoryLocator
	newCmdRunner         CommandRunner
	outputBucket         *bytes.Buffer
	logger               trace.Printer
	stdout               io.Writer
}

//go:generate counterfeiter . TerminalOutputSwitch

type TerminalOutputSwitch interface {
	DisableTerminalOutput(bool)
}

//go:generate counterfeiter . OutputCapture

type OutputCapture interface {
	SetOutputBucket(io.Writer)
}

func NewRpcService(
	outputCapture OutputCapture,
	terminalOutputSwitch TerminalOutputSwitch,
	cliConfig coreconfig.Repository,
	repoLocator api.RepositoryLocator,
	newCmdRunner CommandRunner,
	logger trace.Printer,
	w io.Writer,
	rpcServer *rpc.Server,
) (*CliRpcService, error) {
	rpcService := &CliRpcService{
		Server: rpcServer,
		RpcCmd: &CliRpcCmd{
			PluginMetadata:       &plugin.PluginMetadata{},
			MetadataMutex:        &sync.RWMutex{},
			outputCapture:        outputCapture,
			terminalOutputSwitch: terminalOutputSwitch,
			cliConfig:            cliConfig,
			repoLocator:          repoLocator,
			newCmdRunner:         newCmdRunner,
			logger:               logger,
			outputBucket:         &bytes.Buffer{},
			stdout:               w,
		},
	}

	err := rpcService.Server.Register(rpcService.RpcCmd)
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
				go cli.Server.ServeConn(conn)
			}
		}
	}()

	return nil
}

func (cmd *CliRpcCmd) IsMinCliVersion(passedVersion string, retVal *bool) error {
	if version.BinaryVersion == "BUILT_FROM_SOURCE" {
		*retVal = true
		return nil
	}

	actualVersion, err := semver.Make(version.BinaryVersion)
	if err != nil {
		return err
	}

	requiredVersion, err := semver.Make(passedVersion)
	if err != nil {
		return err
	}

	*retVal = actualVersion.GTE(requiredVersion)

	return nil
}

func (cmd *CliRpcCmd) SetPluginMetadata(pluginMetadata plugin.PluginMetadata, retVal *bool) error {
	cmd.MetadataMutex.Lock()
	defer cmd.MetadataMutex.Unlock()

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
	var err error
	cmdRegistry := commandregistry.Commands

	cmd.outputBucket = &bytes.Buffer{}
	cmd.outputCapture.SetOutputBucket(cmd.outputBucket)

	if cmdRegistry.CommandExists(args[0]) {
		deps := commandregistry.NewDependency(cmd.stdout, cmd.logger, dialTimeout)

		//set deps objs to be the one used by all other commands
		//once all commands are converted, we can make fresh deps for each command run
		deps.Config = cmd.cliConfig
		deps.RepoLocator = cmd.repoLocator

		//set command ui's TeePrinter to be the one used by RpcService, for output to be captured
		deps.UI = terminal.NewUI(os.Stdin, cmd.stdout, cmd.outputCapture.(*terminal.TeePrinter), cmd.logger)

		err = cmd.newCmdRunner.Command(args, deps, false)
	} else {
		*retVal = false
		return nil
	}

	if err != nil {
		*retVal = false
		return err
	}

	*retVal = true
	return nil
}

func (cmd *CliRpcCmd) GetOutputAndReset(args bool, retVal *[]string) error {
	v := strings.TrimSuffix(cmd.outputBucket.String(), "\n")
	*retVal = strings.Split(v, "\n")
	return nil
}

func (cmd *CliRpcCmd) GetCurrentOrg(args string, retVal *plugin_models.Organization) error {
	retVal.Name = cmd.cliConfig.OrganizationFields().Name
	retVal.Guid = cmd.cliConfig.OrganizationFields().GUID
	return nil
}

func (cmd *CliRpcCmd) GetCurrentSpace(args string, retVal *plugin_models.Space) error {
	retVal.Name = cmd.cliConfig.SpaceFields().Name
	retVal.Guid = cmd.cliConfig.SpaceFields().GUID

	return nil
}

func (cmd *CliRpcCmd) Username(args string, retVal *string) error {
	*retVal = cmd.cliConfig.Username()

	return nil
}

func (cmd *CliRpcCmd) UserGuid(args string, retVal *string) error {
	*retVal = cmd.cliConfig.UserGUID()

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
	*retVal = cmd.cliConfig.APIEndpoint()

	return nil
}

func (cmd *CliRpcCmd) HasAPIEndpoint(args string, retVal *bool) error {
	*retVal = cmd.cliConfig.HasAPIEndpoint()

	return nil
}

func (cmd *CliRpcCmd) ApiVersion(args string, retVal *string) error {
	*retVal = cmd.cliConfig.APIVersion()

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
	token, err := cmd.repoLocator.GetAuthenticationRepository().RefreshAuthToken()
	if err != nil {
		return err
	}

	*retVal = token

	return nil
}

func (cmd *CliRpcCmd) GetApp(appName string, retVal *plugin_models.GetAppModel) error {
	deps := commandregistry.NewDependency(cmd.stdout, cmd.logger, dialTimeout)

	//set deps objs to be the one used by all other commands
	//once all commands are converted, we can make fresh deps for each command run
	deps.Config = cmd.cliConfig
	deps.RepoLocator = cmd.repoLocator
	deps.PluginModels.Application = retVal
	cmd.terminalOutputSwitch.DisableTerminalOutput(true)
	deps.UI = terminal.NewUI(os.Stdin, cmd.stdout, cmd.terminalOutputSwitch.(*terminal.TeePrinter), cmd.logger)

	return cmd.newCmdRunner.Command([]string{"app", appName}, deps, true)
}

func (cmd *CliRpcCmd) GetApps(_ string, retVal *[]plugin_models.GetAppsModel) error {
	deps := commandregistry.NewDependency(cmd.stdout, cmd.logger, dialTimeout)

	//set deps objs to be the one used by all other commands
	//once all commands are converted, we can make fresh deps for each command run
	deps.Config = cmd.cliConfig
	deps.RepoLocator = cmd.repoLocator
	deps.PluginModels.AppsSummary = retVal
	cmd.terminalOutputSwitch.DisableTerminalOutput(true)
	deps.UI = terminal.NewUI(os.Stdin, cmd.stdout, cmd.terminalOutputSwitch.(*terminal.TeePrinter), cmd.logger)

	return cmd.newCmdRunner.Command([]string{"apps"}, deps, true)
}

func (cmd *CliRpcCmd) GetOrgs(_ string, retVal *[]plugin_models.GetOrgs_Model) error {
	deps := commandregistry.NewDependency(cmd.stdout, cmd.logger, dialTimeout)

	//set deps objs to be the one used by all other commands
	//once all commands are converted, we can make fresh deps for each command run
	deps.Config = cmd.cliConfig
	deps.RepoLocator = cmd.repoLocator
	deps.PluginModels.Organizations = retVal
	cmd.terminalOutputSwitch.DisableTerminalOutput(true)
	deps.UI = terminal.NewUI(os.Stdin, cmd.stdout, cmd.terminalOutputSwitch.(*terminal.TeePrinter), cmd.logger)

	return cmd.newCmdRunner.Command([]string{"orgs"}, deps, true)
}

func (cmd *CliRpcCmd) GetSpaces(_ string, retVal *[]plugin_models.GetSpaces_Model) error {
	deps := commandregistry.NewDependency(cmd.stdout, cmd.logger, dialTimeout)

	//set deps objs to be the one used by all other commands
	//once all commands are converted, we can make fresh deps for each command run
	deps.Config = cmd.cliConfig
	deps.RepoLocator = cmd.repoLocator
	deps.PluginModels.Spaces = retVal
	cmd.terminalOutputSwitch.DisableTerminalOutput(true)
	deps.UI = terminal.NewUI(os.Stdin, cmd.stdout, cmd.terminalOutputSwitch.(*terminal.TeePrinter), cmd.logger)

	return cmd.newCmdRunner.Command([]string{"spaces"}, deps, true)
}

func (cmd *CliRpcCmd) GetServices(_ string, retVal *[]plugin_models.GetServices_Model) error {
	deps := commandregistry.NewDependency(cmd.stdout, cmd.logger, dialTimeout)

	//set deps objs to be the one used by all other commands
	//once all commands are converted, we can make fresh deps for each command run
	//once all commands are converted, we can make fresh deps for each command run
	deps.Config = cmd.cliConfig
	deps.RepoLocator = cmd.repoLocator
	deps.PluginModels.Services = retVal
	cmd.terminalOutputSwitch.DisableTerminalOutput(true)
	deps.UI = terminal.NewUI(os.Stdin, cmd.stdout, cmd.terminalOutputSwitch.(*terminal.TeePrinter), cmd.logger)

	return cmd.newCmdRunner.Command([]string{"services"}, deps, true)
}

func (cmd *CliRpcCmd) GetOrgUsers(args []string, retVal *[]plugin_models.GetOrgUsers_Model) error {
	deps := commandregistry.NewDependency(cmd.stdout, cmd.logger, dialTimeout)

	//set deps objs to be the one used by all other commands
	//once all commands are converted, we can make fresh deps for each command run
	deps.Config = cmd.cliConfig
	deps.RepoLocator = cmd.repoLocator
	deps.PluginModels.OrgUsers = retVal
	cmd.terminalOutputSwitch.DisableTerminalOutput(true)
	deps.UI = terminal.NewUI(os.Stdin, cmd.stdout, cmd.terminalOutputSwitch.(*terminal.TeePrinter), cmd.logger)

	return cmd.newCmdRunner.Command(append([]string{"org-users"}, args...), deps, true)
}

func (cmd *CliRpcCmd) GetSpaceUsers(args []string, retVal *[]plugin_models.GetSpaceUsers_Model) error {
	deps := commandregistry.NewDependency(cmd.stdout, cmd.logger, dialTimeout)

	//set deps objs to be the one used by all other commands
	//once all commands are converted, we can make fresh deps for each command run
	deps.Config = cmd.cliConfig
	deps.RepoLocator = cmd.repoLocator
	deps.PluginModels.SpaceUsers = retVal
	cmd.terminalOutputSwitch.DisableTerminalOutput(true)
	deps.UI = terminal.NewUI(os.Stdin, cmd.stdout, cmd.terminalOutputSwitch.(*terminal.TeePrinter), cmd.logger)

	return cmd.newCmdRunner.Command(append([]string{"space-users"}, args...), deps, true)
}

func (cmd *CliRpcCmd) GetOrg(orgName string, retVal *plugin_models.GetOrg_Model) error {
	deps := commandregistry.NewDependency(cmd.stdout, cmd.logger, dialTimeout)

	//set deps objs to be the one used by all other commands
	//once all commands are converted, we can make fresh deps for each command run
	deps.Config = cmd.cliConfig
	deps.RepoLocator = cmd.repoLocator
	deps.PluginModels.Organization = retVal
	cmd.terminalOutputSwitch.DisableTerminalOutput(true)
	deps.UI = terminal.NewUI(os.Stdin, cmd.stdout, cmd.terminalOutputSwitch.(*terminal.TeePrinter), cmd.logger)

	return cmd.newCmdRunner.Command([]string{"org", orgName}, deps, true)
}

func (cmd *CliRpcCmd) GetSpace(spaceName string, retVal *plugin_models.GetSpace_Model) error {
	deps := commandregistry.NewDependency(cmd.stdout, cmd.logger, dialTimeout)

	//set deps objs to be the one used by all other commands
	//once all commands are converted, we can make fresh deps for each command run
	deps.Config = cmd.cliConfig
	deps.RepoLocator = cmd.repoLocator
	deps.PluginModels.Space = retVal
	cmd.terminalOutputSwitch.DisableTerminalOutput(true)
	deps.UI = terminal.NewUI(os.Stdin, cmd.stdout, cmd.terminalOutputSwitch.(*terminal.TeePrinter), cmd.logger)

	return cmd.newCmdRunner.Command([]string{"space", spaceName}, deps, true)
}

func (cmd *CliRpcCmd) GetService(serviceInstance string, retVal *plugin_models.GetService_Model) error {
	deps := commandregistry.NewDependency(cmd.stdout, cmd.logger, dialTimeout)

	//set deps objs to be the one used by all other commands
	//once all commands are converted, we can make fresh deps for each command run
	deps.Config = cmd.cliConfig
	deps.RepoLocator = cmd.repoLocator
	deps.PluginModels.Service = retVal
	cmd.terminalOutputSwitch.DisableTerminalOutput(true)
	deps.UI = terminal.NewUI(os.Stdin, cmd.stdout, cmd.terminalOutputSwitch.(*terminal.TeePrinter), cmd.logger)

	return cmd.newCmdRunner.Command([]string{"service", serviceInstance}, deps, true)
}
