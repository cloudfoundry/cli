package rpc

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/cli/v8/actor/sharedaction"
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	ccWrapper "code.cloudfoundry.org/cli/v8/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/v8/api/router"
	routingWrapper "code.cloudfoundry.org/cli/v8/api/router/wrapper"
	"code.cloudfoundry.org/cli/v8/api/uaa"
	uaaWrapper "code.cloudfoundry.org/cli/v8/api/uaa/wrapper"
	"code.cloudfoundry.org/cli/v8/cf/api"
	"code.cloudfoundry.org/cli/v8/cf/commandregistry"
	"code.cloudfoundry.org/cli/v8/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v8/cf/terminal"
	"code.cloudfoundry.org/cli/v8/plugin"
	plugin_models "code.cloudfoundry.org/cli/v8/plugin/models"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/version"
	"github.com/blang/semver/v4"

	"code.cloudfoundry.org/cli/v8/cf/trace"
)

var dialTimeout = os.Getenv("CF_DIAL_TIMEOUT")

// configAdapter adapts coreconfig.Repository to the various Config interfaces needed
type configAdapter struct {
	coreconfig.Repository
	timeout time.Duration
}

func (c *configAdapter) BinaryName() string {
	return "CF_PLUGIN_RPC"
}

func (c *configAdapter) BinaryVersion() string {
	return version.VersionString()
}

func (c *configAdapter) DialTimeout() time.Duration {
	return c.timeout
}

func (c *configAdapter) SetUAAEndpoint(endpoint string) {
	c.Repository.SetUaaEndpoint(endpoint)
}

func (c *configAdapter) SkipSSLValidation() bool {
	return c.Repository.IsSSLDisabled()
}

func (c *configAdapter) UAADisableKeepAlives() bool {
	return false // Default behavior
}

func (c *configAdapter) CurrentUserName() (string, error) {
	return c.Repository.Username(), nil
}

func (c *configAdapter) HasTargetedOrganization() bool {
	return c.Repository.HasOrganization()
}

func (c *configAdapter) HasTargetedSpace() bool {
	return c.Repository.HasSpace()
}

func (c *configAdapter) TargetedOrganizationName() string {
	return c.Repository.OrganizationFields().Name
}

func (c *configAdapter) Verbose() (bool, []string) {
	return false, nil // RPC doesn't need verbose output
}

func (c *configAdapter) CurrentUser() (configv3.User, error) {
	return configv3.User{
		Name: c.Repository.Username(),
		GUID: c.Repository.UserGUID(),
	}, nil
}

func (c *configAdapter) PollingInterval() time.Duration {
	return 2 * time.Second
}

func (c *configAdapter) StagingTimeout() time.Duration {
	return time.Duration(c.Repository.AsyncTimeout()) * time.Minute
}

func (c *configAdapter) StartupTimeout() time.Duration {
	return time.Duration(c.Repository.AsyncTimeout()) * time.Minute
}

func (c *configAdapter) Target() string {
	return c.Repository.APIEndpoint()
}

func (c *configAdapter) IsCFOnK8s() bool {
	return false // RPC plugins don't support k8s mode
}

func (c *configAdapter) SetKubernetesAuthInfo(authInfo string) {
	// No-op for RPC plugins
}

func (c *configAdapter) SetTargetInformation(args configv3.TargetInformationArgs) {
	// Forward to underlying repository methods
	if args.Api != "" {
		c.Repository.SetAPIEndpoint(args.Api)
	}
	if args.ApiVersion != "" {
		c.Repository.SetAPIVersion(args.ApiVersion)
	}
	if args.Auth != "" {
		c.Repository.SetAuthenticationEndpoint(args.Auth)
	}
	if args.Doppler != "" {
		c.Repository.SetDopplerEndpoint(args.Doppler)
	}
	if args.Routing != "" {
		c.Repository.SetRoutingAPIEndpoint(args.Routing)
	}
	c.Repository.SetUaaEndpoint(args.UAA)
	c.Repository.SetSSLDisabled(args.SkipSSLValidation)
}

func (c *configAdapter) SetTokenInformation(accessToken string, refreshToken string, sshOAuthClient string) {
	c.Repository.SetAccessToken(accessToken)
	c.Repository.SetRefreshToken(refreshToken)
}

func (c *configAdapter) SetUAAClientCredentials(client string, clientSecret string) {
	c.Repository.SetUAAOAuthClient(client)
	c.Repository.SetUAAOAuthClientSecret(clientSecret)
}

func (c *configAdapter) SetUAAGrantType(grantType string) {
	c.Repository.SetUAAGrantType(grantType)
}

func (c *configAdapter) UnsetOrganizationAndSpaceInformation() {
	// Use ClearSession which clears org and space info
	c.Repository.ClearSession()
}

func (c *configAdapter) SSHOAuthClient() string {
	return c.Repository.SSHOAuthClient()
}

// Simple logger that implements RequestLoggerOutput interfaces for all clients
type simpleRequestLogger struct {
	printer trace.Printer
}

func (l *simpleRequestLogger) DisplayBody(body []byte) error {
	l.printer.Printf("%s\n", body)
	return nil
}

func (l *simpleRequestLogger) DisplayHeader(name string, value string) error {
	l.printer.Printf("%s: %s\n", name, value)
	return nil
}

func (l *simpleRequestLogger) DisplayHost(name string) error {
	l.printer.Printf("Host: %s\n", name)
	return nil
}

func (l *simpleRequestLogger) DisplayJSONBody(body []byte) error {
	l.printer.Printf("%s\n", body)
	return nil
}

func (l *simpleRequestLogger) DisplayMessage(msg string) error {
	l.printer.Printf("%s\n", msg)
	return nil
}

func (l *simpleRequestLogger) DisplayRequestHeader(method string, uri string, httpProtocol string) error {
	l.printer.Printf("%s %s %s\n", method, uri, httpProtocol)
	return nil
}

func (l *simpleRequestLogger) DisplayResponseHeader(httpProtocol string, status string) error {
	l.printer.Printf("%s %s\n", httpProtocol, status)
	return nil
}

func (l *simpleRequestLogger) DisplayType(name string, requestDate time.Time) error {
	l.printer.Printf("%s: [%s]\n", name, requestDate.Format(time.RFC3339))
	return nil
}

func (l *simpleRequestLogger) HandleInternalError(err error) {
	l.printer.Printf("Error: %v\n", err)
}

func (l *simpleRequestLogger) Start() error {
	return nil
}

func (l *simpleRequestLogger) Stop() error {
	return nil
}

// getClientsForActor creates the CC v3, UAA, and Routing clients needed for v7action.Actor
// with proper authentication wrappers and CF_TRACE support
func getClientsForActor(config coreconfig.Repository, logger trace.Printer, envDialTimeout string) (*ccv3.Client, *uaa.Client, *router.Client, error) {
	// Handle nil config (used in tests)
	if config == nil {
		return nil, nil, nil, nil
	}

	// Parse dial timeout
	var timeout time.Duration
	if envDialTimeout != "" {
		parsedTimeout, err := strconv.Atoi(envDialTimeout)
		if err == nil {
			timeout = time.Duration(parsedTimeout) * time.Second
		}
	}
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	// Create config adapter for UAA client
	configAdapt := &configAdapter{
		Repository: config,
		timeout:    timeout,
	}

	// Check CF_TRACE environment variable for logging
	cfTrace := os.Getenv("CF_TRACE")
	verbose := cfTrace == "true" || cfTrace == "1"
	var traceLogger *simpleRequestLogger

	// If CF_TRACE is enabled, create request logger
	if verbose && logger != nil {
		traceLogger = &simpleRequestLogger{printer: logger}
	} else if cfTrace != "" && cfTrace != "false" && cfTrace != "0" {
		// CF_TRACE is a file path - create a file logger
		traceFile, err := os.OpenFile(cfTrace, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err == nil {
			// Create a logger that writes to the file
			traceLogger = &simpleRequestLogger{
				printer: trace.NewLogger(traceFile, false, "", ""),
			}
		}
	}

	// Create UAA client with authentication wrapper
	uaaClient := uaa.NewClient(configAdapt)

	// Add request logger if CF_TRACE is enabled
	if traceLogger != nil {
		uaaClient.WrapConnection(uaaWrapper.NewRequestLogger(traceLogger))
	}

	// Add UAA authentication wrapper (critical for token refresh)
	uaaAuthWrapper := uaaWrapper.NewUAAAuthentication(uaaClient, configAdapt)
	uaaClient.WrapConnection(uaaAuthWrapper)

	// Add retry wrapper for resilience
	uaaClient.WrapConnection(uaaWrapper.NewRetryRequest(2)) // default retry count

	err := uaaClient.SetupResources(config.UaaEndpoint(), config.AuthenticationEndpoint())
	if err != nil {
		return nil, nil, nil, err
	}

	// Create CC v3 client with authentication wrapper
	ccWrappers := []ccv3.ConnectionWrapper{}

	// Add request logger if CF_TRACE is enabled
	if traceLogger != nil {
		ccWrappers = append(ccWrappers, ccWrapper.NewRequestLogger(traceLogger))
	}

	// Add authentication wrapper (uses UAA client for token management)
	ccAuthWrapper := ccWrapper.NewUAAAuthentication(uaaClient, configAdapt)
	ccWrappers = append(ccWrappers, ccAuthWrapper)

	// Add retry wrapper for resilience
	ccWrappers = append(ccWrappers, ccWrapper.NewRetryRequest(2)) // default retry count

	ccClient := ccv3.NewClient(ccv3.Config{
		AppName:            "CF_PLUGIN_RPC",
		AppVersion:         version.VersionString(),
		JobPollingTimeout:  time.Duration(config.AsyncTimeout()) * time.Minute,
		JobPollingInterval: 2 * time.Second,
		Wrappers:           ccWrappers,
	})

	ccClient.TargetCF(ccv3.TargetSettings{
		URL:               config.APIEndpoint(),
		SkipSSLValidation: config.IsSSLDisabled(),
		DialTimeout:       timeout,
	})

	// Create routing client with authentication wrapper
	routingWrappers := []router.ConnectionWrapper{}

	// Add request logger if CF_TRACE is enabled
	if traceLogger != nil {
		routingWrappers = append(routingWrappers, routingWrapper.NewRequestLogger(traceLogger))
	}

	// Add authentication wrapper (uses UAA client for token management)
	routingAuthWrapper := routingWrapper.NewUAAAuthentication(uaaClient, configAdapt)
	routingWrappers = append(routingWrappers, routingAuthWrapper)

	routingClient := router.NewClient(router.Config{
		AppName:    "CF_PLUGIN_RPC",
		AppVersion: version.VersionString(),
		ConnectionConfig: router.ConnectionConfig{
			DialTimeout:       timeout,
			SkipSSLValidation: config.IsSSLDisabled(),
		},
		RoutingEndpoint: config.RoutingAPIEndpoint(),
		Wrappers:        routingWrappers,
	})

	return ccClient, uaaClient, routingClient, nil
}

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
	repoLocator          api.RepositoryLocator // Only used by GetApp and CallCoreCommand
	actor                *v7action.Actor
	sharedActor          *sharedaction.Actor
	newCmdRunner         CommandRunner
	outputBucket         *bytes.Buffer
	logger               trace.Printer
	stdout               io.Writer
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . TerminalOutputSwitch

type TerminalOutputSwitch interface {
	DisableTerminalOutput(bool)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . OutputCapture

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
	// Parse dial timeout
	var timeout time.Duration
	envDialTimeout := os.Getenv("CF_DIAL_TIMEOUT")
	if envDialTimeout != "" {
		parsedTimeout, err := strconv.Atoi(envDialTimeout)
		if err == nil {
			timeout = time.Duration(parsedTimeout) * time.Second
		}
	}
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	// Create config adapter for actors
	configAdapt := &configAdapter{
		Repository: cliConfig,
		timeout:    timeout,
	}

	// Create shared actor - handle nil config
	var sharedActor *sharedaction.Actor
	if cliConfig != nil {
		sharedActor = sharedaction.NewActor(configAdapt)
	}

	// Create v7 actor with clients
	ccClient, uaaClient, routingClient, err := getClientsForActor(cliConfig, logger, envDialTimeout)
	if err != nil {
		return nil, err
	}

	// Create v7 actor - handle nil clients for tests
	var actor *v7action.Actor
	if ccClient != nil && uaaClient != nil && routingClient != nil {
		actor = v7action.NewActor(ccClient, configAdapt, sharedActor, uaaClient, routingClient, nil)
	}

	rpcService := &CliRpcService{
		Server: rpcServer,
		RpcCmd: &CliRpcCmd{
			PluginMetadata:       &plugin.PluginMetadata{},
			MetadataMutex:        &sync.RWMutex{},
			outputCapture:        outputCapture,
			terminalOutputSwitch: terminalOutputSwitch,
			cliConfig:            cliConfig,
			repoLocator:          repoLocator,
			actor:                actor,
			sharedActor:          sharedActor,
			newCmdRunner:         newCmdRunner,
			logger:               logger,
			outputBucket:         &bytes.Buffer{},
			stdout:               w,
		},
	}

	err = rpcService.Server.Register(rpcService.RpcCmd)
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
	if version.VersionString() == version.DefaultVersion {
		*retVal = true
		return nil
	}

	actualVersion, err := semver.Make(version.VersionString())
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
		// NOTE: CallCoreCommand uses the legacy commandregistry approach for executing arbitrary commands.
		// This is intentional because:
		// 1. Plugins can call any CF command, and we need a generic execution mechanism
		// 2. The legacy commandregistry provides this via the Command() method
		// 3. RPC-specific methods (GetOrgs, GetSpaces, GetServices, etc.) have been migrated to v7action.Actor
		// 4. The output capture mechanism is tightly integrated with the legacy terminal.UI
		//
		// Future enhancement: Could add v7 command routing using util/command_parser package,
		// but this requires careful handling of output capture and UI initialization.
		deps := commandregistry.NewDependency(cmd.stdout, cmd.logger, dialTimeout)

		// set deps objs to be the one used by all other commands
		// once all commands are converted, we can make fresh deps for each command run
		deps.Config = cmd.cliConfig
		deps.RepoLocator = cmd.repoLocator

		// set command ui's TeePrinter to be the one used by RpcService, for output to be captured
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
	*retVal = ""

	return nil
}

func (cmd *CliRpcCmd) DopplerEndpoint(args string, retVal *string) error {
	*retVal = cmd.cliConfig.DopplerEndpoint()

	return nil
}

func (cmd *CliRpcCmd) AccessToken(args string, retVal *string) error {
	if cmd.actor == nil {
		return fmt.Errorf("v7 actor not initialized")
	}
	token, err := cmd.actor.RefreshAccessToken()
	if err != nil {
		return err
	}
	*retVal = token
	return nil
}

// Model mapping helper functions for v7action resources to plugin_models

func mapOrganizationToPluginModel(org resources.Organization) plugin_models.GetOrgs_Model {
	return plugin_models.GetOrgs_Model{
		Guid: org.GUID,
		Name: org.Name,
	}
}

func mapSpaceToPluginModel(space resources.Space) plugin_models.GetSpaces_Model {
	return plugin_models.GetSpaces_Model{
		Guid: space.GUID,
		Name: space.Name,
	}
}

func mapApplicationSummaryToPluginModel(summary v7action.ApplicationSummary) plugin_models.GetAppsModel {
	model := plugin_models.GetAppsModel{
		Name:  summary.Name,
		Guid:  summary.GUID,
		State: string(summary.State),
	}

	// Get data from the first (web) process if available
	if len(summary.ProcessSummaries) > 0 {
		webProcess := summary.ProcessSummaries[0]
		model.TotalInstances = webProcess.TotalInstanceCount()
		model.RunningInstances = webProcess.HealthyInstanceCount()

		if webProcess.MemoryInMB.IsSet {
			model.Memory = int64(webProcess.MemoryInMB.Value)
		}
		if webProcess.DiskInMB.IsSet {
			model.DiskQuota = int64(webProcess.DiskInMB.Value)
		}
	}

	// Map routes
	if len(summary.Routes) > 0 {
		model.Routes = make([]plugin_models.GetAppsRouteSummary, len(summary.Routes))
		for i, route := range summary.Routes {
			model.Routes[i] = plugin_models.GetAppsRouteSummary{
				Guid: route.GUID,
				Host: route.Host,
				Domain: plugin_models.GetAppsDomainFields{
					Guid: route.DomainGUID,
					Name: "", // Domain name not included in route resource, would need separate call
				},
			}
		}
	}

	return model
}

func mapServiceInstanceToPluginModel(instance v7action.ServiceInstance) plugin_models.GetServices_Model {
	return plugin_models.GetServices_Model{
		Guid: "", // Will need to be filled from resources.ServiceInstance if available
		Name: instance.Name,
		ServicePlan: plugin_models.GetServices_ServicePlan{
			Guid: "", // Not directly available in ServiceInstance summary
			Name: instance.ServicePlanName,
		},
		Service: plugin_models.GetServices_ServiceFields{
			Name: instance.ServiceOfferingName,
		},
		LastOperation: plugin_models.GetServices_LastOperation{
			Type:  "", // Not available in summary, would show in LastOperation string
			State: instance.LastOperation,
		},
		ApplicationNames: instance.BoundApps,
		IsUserProvided:   instance.Type == "user-provided",
	}
}

func mapRoleTypeToString(roleType constant.RoleType) string {
	// Map constant role types to plugin-friendly role names
	switch roleType {
	case constant.OrgUserRole:
		return "OrgUser"
	case constant.OrgAuditorRole:
		return "OrgAuditor"
	case constant.OrgManagerRole:
		return "OrgManager"
	case constant.OrgBillingManagerRole:
		return "BillingManager"
	case constant.SpaceDeveloperRole:
		return "SpaceDeveloper"
	case constant.SpaceAuditorRole:
		return "SpaceAuditor"
	case constant.SpaceManagerRole:
		return "SpaceManager"
	case constant.SpaceSupporterRole:
		return "SpaceSupporter"
	default:
		return string(roleType)
	}
}

func (cmd *CliRpcCmd) GetApp(appName string, retVal *plugin_models.GetAppModel) error {
	// NOTE: GetApp uses the legacy commandregistry approach.
	// This method requires extensive data gathering including:
	// - Application details (name, guid, memory, disk, instances, etc.)
	// - Instance states and statistics (CPU, memory, disk usage per instance)
	// - Routes and domains
	// - Bound services
	// - Stack information
	// - Environment variables
	// - Package state and staging details
	//
	// Migrating this to v7action.Actor would require coordinating multiple API calls:
	// - GetApplicationByNameAndSpace() for basic app info
	// - GetApplicationInstancesByApplication() for instance stats
	// - GetApplicationRoutes() for routes
	// - GetServiceBindingsByApplication() for bound services
	// - GetApplicationEnvironmentVariables() for env vars
	// - Additional calls for stack, package info, etc.
	//
	// The legacy command provides this in a single coordinated call with proper error handling
	// and data aggregation. Migration deferred until there's a clear benefit or v7action provides
	// a comprehensive GetApplicationDetails() method similar to GetSpaceSummary().
	deps := commandregistry.NewDependency(cmd.stdout, cmd.logger, dialTimeout)

	// set deps objs to be the one used by all other commands
	// once all commands are converted, we can make fresh deps for each command run
	deps.Config = cmd.cliConfig
	deps.RepoLocator = cmd.repoLocator
	deps.PluginModels.Application = retVal
	cmd.terminalOutputSwitch.DisableTerminalOutput(true)
	deps.UI = terminal.NewUI(os.Stdin, cmd.stdout, cmd.terminalOutputSwitch.(*terminal.TeePrinter), cmd.logger)

	return cmd.newCmdRunner.Command([]string{"app", appName}, deps, true)
}

func (cmd *CliRpcCmd) GetApps(_ string, retVal *[]plugin_models.GetAppsModel) error {
	if cmd.actor == nil {
		return fmt.Errorf("v7 actor not initialized")
	}
	// Get current space GUID from config
	spaceGUID := cmd.cliConfig.SpaceFields().GUID
	if spaceGUID == "" {
		return fmt.Errorf("no space targeted")
	}

	// Get app summaries for the space (omitStats=false to include instance info)
	summaries, _, err := cmd.actor.GetAppSummariesForSpace(spaceGUID, "", false)
	if err != nil {
		return err
	}

	// Convert v7action summaries to plugin_models
	result := make([]plugin_models.GetAppsModel, len(summaries))
	for i, summary := range summaries {
		result[i] = mapApplicationSummaryToPluginModel(summary)
	}
	*retVal = result
	return nil
}

func (cmd *CliRpcCmd) GetOrgs(_ string, retVal *[]plugin_models.GetOrgs_Model) error {
	if cmd.actor == nil {
		return fmt.Errorf("v7 actor not initialized")
	}
	orgs, _, err := cmd.actor.GetOrganizations("")
	if err != nil {
		return err
	}

	// Convert v7action resources to plugin_models
	result := make([]plugin_models.GetOrgs_Model, len(orgs))
	for i, org := range orgs {
		result[i] = mapOrganizationToPluginModel(org)
	}
	*retVal = result
	return nil
}

func (cmd *CliRpcCmd) GetSpaces(_ string, retVal *[]plugin_models.GetSpaces_Model) error {
	if cmd.actor == nil {
		return fmt.Errorf("v7 actor not initialized")
	}
	// Get current org GUID from config
	orgGUID := cmd.cliConfig.OrganizationFields().GUID
	if orgGUID == "" {
		return fmt.Errorf("no organization targeted")
	}

	spaces, _, err := cmd.actor.GetOrganizationSpaces(orgGUID)
	if err != nil {
		return err
	}

	// Convert v7action resources to plugin_models
	result := make([]plugin_models.GetSpaces_Model, len(spaces))
	for i, space := range spaces {
		result[i] = mapSpaceToPluginModel(space)
	}
	*retVal = result
	return nil
}

func (cmd *CliRpcCmd) GetServices(_ string, retVal *[]plugin_models.GetServices_Model) error {
	if cmd.actor == nil {
		return fmt.Errorf("v7 actor not initialized")
	}
	// Get current space GUID from config
	spaceGUID := cmd.cliConfig.SpaceFields().GUID
	if spaceGUID == "" {
		return fmt.Errorf("no space targeted")
	}

	// Get service instances for the space (omitApps=false to include bound apps)
	instances, _, err := cmd.actor.GetServiceInstancesForSpace(spaceGUID, false)
	if err != nil {
		return err
	}

	// Convert v7action service instances to plugin_models
	result := make([]plugin_models.GetServices_Model, len(instances))
	for i, instance := range instances {
		result[i] = mapServiceInstanceToPluginModel(instance)
	}
	*retVal = result
	return nil
}

func (cmd *CliRpcCmd) GetOrgUsers(args []string, retVal *[]plugin_models.GetOrgUsers_Model) error {
	if cmd.actor == nil {
		return fmt.Errorf("v7 actor not initialized")
	}
	if len(args) == 0 {
		return fmt.Errorf("organization name required")
	}
	orgName := args[0]

	// Get organization by name
	org, _, err := cmd.actor.GetOrganizationByName(orgName)
	if err != nil {
		return err
	}

	// Get users by role type
	usersByRole, _, err := cmd.actor.GetOrgUsersByRoleType(org.GUID)
	if err != nil {
		return err
	}

	// Build a map of unique users with their roles
	userMap := make(map[string]*plugin_models.GetOrgUsers_Model)

	for roleType, users := range usersByRole {
		roleName := mapRoleTypeToString(roleType)
		for _, user := range users {
			if existing, found := userMap[user.GUID]; found {
				existing.Roles = append(existing.Roles, roleName)
			} else {
				userMap[user.GUID] = &plugin_models.GetOrgUsers_Model{
					Guid:     user.GUID,
					Username: user.Username,
					IsAdmin:  false, // Plugin RPC doesn't have access to UAA admin status
					Roles:    []string{roleName},
				}
			}
		}
	}

	// Convert map to slice
	result := make([]plugin_models.GetOrgUsers_Model, 0, len(userMap))
	for _, user := range userMap {
		result = append(result, *user)
	}
	*retVal = result
	return nil
}

func (cmd *CliRpcCmd) GetSpaceUsers(args []string, retVal *[]plugin_models.GetSpaceUsers_Model) error {
	if cmd.actor == nil {
		return fmt.Errorf("v7 actor not initialized")
	}
	if len(args) < 2 {
		return fmt.Errorf("organization name and space name required")
	}
	orgName := args[0]
	spaceName := args[1]

	// Get organization by name
	org, _, err := cmd.actor.GetOrganizationByName(orgName)
	if err != nil {
		return err
	}

	// Get space by name and organization
	space, _, err := cmd.actor.GetSpaceByNameAndOrganization(spaceName, org.GUID)
	if err != nil {
		return err
	}

	// Get users by role type
	usersByRole, _, err := cmd.actor.GetSpaceUsersByRoleType(space.GUID)
	if err != nil {
		return err
	}

	// Build a map of unique users with their roles
	userMap := make(map[string]*plugin_models.GetSpaceUsers_Model)

	for roleType, users := range usersByRole {
		roleName := mapRoleTypeToString(roleType)
		for _, user := range users {
			if existing, found := userMap[user.GUID]; found {
				existing.Roles = append(existing.Roles, roleName)
			} else {
				userMap[user.GUID] = &plugin_models.GetSpaceUsers_Model{
					Guid:     user.GUID,
					Username: user.Username,
					IsAdmin:  false, // Plugin RPC doesn't have access to UAA admin status
					Roles:    []string{roleName},
				}
			}
		}
	}

	// Convert map to slice
	result := make([]plugin_models.GetSpaceUsers_Model, 0, len(userMap))
	for _, user := range userMap {
		result = append(result, *user)
	}
	*retVal = result
	return nil
}

func (cmd *CliRpcCmd) GetOrg(orgName string, retVal *plugin_models.GetOrg_Model) error {
	if cmd.actor == nil {
		return fmt.Errorf("v7 actor not initialized")
	}
	// Get organization by name
	org, _, err := cmd.actor.GetOrganizationByName(orgName)
	if err != nil {
		return err
	}

	// Get spaces in organization
	spaces, _, err := cmd.actor.GetOrganizationSpaces(org.GUID)
	if err != nil {
		return err
	}

	// Get domains for organization
	domains, _, err := cmd.actor.GetOrganizationDomains(org.GUID, "")
	if err != nil {
		return err
	}

	// Get space quotas for organization
	spaceQuotas, _, err := cmd.actor.GetSpaceQuotasByOrgGUID(org.GUID)
	if err != nil {
		return err
	}

	// Map spaces
	orgSpaces := make([]plugin_models.GetOrg_Space, len(spaces))
	for i, space := range spaces {
		orgSpaces[i] = plugin_models.GetOrg_Space{
			Guid: space.GUID,
			Name: space.Name,
		}
	}

	// Map domains
	orgDomains := make([]plugin_models.GetOrg_Domains, len(domains))
	for i, domain := range domains {
		orgDomains[i] = plugin_models.GetOrg_Domains{
			Guid:                   domain.GUID,
			Name:                   domain.Name,
			OwningOrganizationGuid: domain.OrganizationGUID,
			Shared:                 domain.OrganizationGUID != org.GUID, // Shared if owned by different org
		}
	}

	// Map space quotas
	orgSpaceQuotas := make([]plugin_models.GetOrg_SpaceQuota, len(spaceQuotas))
	for i, quota := range spaceQuotas {
		memoryLimit := int64(0)
		if quota.Apps.TotalMemory != nil && quota.Apps.TotalMemory.IsSet {
			memoryLimit = int64(quota.Apps.TotalMemory.Value)
		}

		instanceMemoryLimit := int64(0)
		if quota.Apps.InstanceMemory != nil && quota.Apps.InstanceMemory.IsSet {
			instanceMemoryLimit = int64(quota.Apps.InstanceMemory.Value)
		}

		routesLimit := 0
		if quota.Routes.TotalRoutes != nil && quota.Routes.TotalRoutes.IsSet {
			routesLimit = quota.Routes.TotalRoutes.Value
		}

		servicesLimit := 0
		if quota.Services.TotalServiceInstances != nil && quota.Services.TotalServiceInstances.IsSet {
			servicesLimit = quota.Services.TotalServiceInstances.Value
		}

		nonBasicServicesAllowed := true
		if quota.Services.PaidServicePlans != nil {
			nonBasicServicesAllowed = *quota.Services.PaidServicePlans
		}

		orgSpaceQuotas[i] = plugin_models.GetOrg_SpaceQuota{
			Guid:                    quota.GUID,
			Name:                    quota.Name,
			MemoryLimit:             memoryLimit,
			InstanceMemoryLimit:     instanceMemoryLimit,
			RoutesLimit:             routesLimit,
			ServicesLimit:           servicesLimit,
			NonBasicServicesAllowed: nonBasicServicesAllowed,
		}
	}

	// Build the result (quota definition would need separate call which we'll skip for now)
	*retVal = plugin_models.GetOrg_Model{
		Guid:            org.GUID,
		Name:            org.Name,
		QuotaDefinition: plugin_models.QuotaFields{}, // Would need separate GetOrganizationQuotaByName call
		Spaces:          orgSpaces,
		Domains:         orgDomains,
		SpaceQuotas:     orgSpaceQuotas,
	}
	return nil
}

func (cmd *CliRpcCmd) GetSpace(spaceName string, retVal *plugin_models.GetSpace_Model) error {
	if cmd.actor == nil {
		return fmt.Errorf("v7 actor not initialized")
	}
	// Get current org GUID from config
	orgGUID := cmd.cliConfig.OrganizationFields().GUID
	if orgGUID == "" {
		return fmt.Errorf("no organization targeted")
	}

	// Get organization info
	org, _, err := cmd.actor.GetOrganizationByGUID(orgGUID)
	if err != nil {
		return err
	}

	// Get space by name and organization
	space, _, err := cmd.actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	if err != nil {
		return err
	}

	// Get applications in space
	apps, _, err := cmd.actor.GetApplicationsBySpace(space.GUID)
	if err != nil {
		return err
	}

	// Get service instances in space
	services, _, err := cmd.actor.GetServiceInstancesForSpace(space.GUID, true) // omitApps=true for performance
	if err != nil {
		return err
	}

	// Get domains for organization (space doesn't have direct domains)
	domains, _, err := cmd.actor.GetOrganizationDomains(orgGUID, "")
	if err != nil {
		return err
	}

	// Get space summary for security groups info
	spaceSummary, _, err := cmd.actor.GetSpaceSummaryByNameAndOrganization(spaceName, orgGUID)
	if err != nil {
		return err
	}

	// Map applications
	spaceApps := make([]plugin_models.GetSpace_Apps, len(apps))
	for i, app := range apps {
		spaceApps[i] = plugin_models.GetSpace_Apps{
			Name: app.Name,
			Guid: app.GUID,
		}
	}

	// Map service instances
	spaceServices := make([]plugin_models.GetSpace_ServiceInstance, len(services))
	for i, service := range services {
		spaceServices[i] = plugin_models.GetSpace_ServiceInstance{
			Guid: "", // Not available in ServiceInstance list summary
			Name: service.Name,
		}
	}

	// Map domains
	spaceDomains := make([]plugin_models.GetSpace_Domains, len(domains))
	for i, domain := range domains {
		spaceDomains[i] = plugin_models.GetSpace_Domains{
			Guid:                   domain.GUID,
			Name:                   domain.Name,
			OwningOrganizationGuid: domain.OrganizationGUID,
			Shared:                 domain.OrganizationGUID != orgGUID,
		}
	}

	// Map security groups
	var securityGroups []plugin_models.GetSpace_SecurityGroup
	for _, sg := range append(spaceSummary.RunningSecurityGroups, spaceSummary.StagingSecurityGroups...) {
		// Convert rules to []map[string]interface{}
		rules := make([]map[string]interface{}, len(sg.Rules))
		for j, rule := range sg.Rules {
			ruleMap := make(map[string]interface{})
			if rule.Protocol != "" {
				ruleMap["protocol"] = rule.Protocol
			}
			if rule.Destination != "" {
				ruleMap["destination"] = rule.Destination
			}
			if rule.Ports != nil && *rule.Ports != "" {
				ruleMap["ports"] = *rule.Ports
			}
			rules[j] = ruleMap
		}

		securityGroups = append(securityGroups, plugin_models.GetSpace_SecurityGroup{
			Name:  sg.Name,
			Guid:  sg.GUID,
			Rules: rules,
		})
	}

	// Get space quota if exists
	spaceQuota := plugin_models.GetSpace_SpaceQuota{}
	if spaceSummary.QuotaName != "" {
		// Would need separate GetSpaceQuotaByName call to get full details
		// For now just set the name
		spaceQuota.Name = spaceSummary.QuotaName
	}

	*retVal = plugin_models.GetSpace_Model{
		GetSpaces_Model: plugin_models.GetSpaces_Model{
			Guid: space.GUID,
			Name: space.Name,
		},
		Organization: plugin_models.GetSpace_Orgs{
			Guid: org.GUID,
			Name: org.Name,
		},
		Applications:     spaceApps,
		ServiceInstances: spaceServices,
		Domains:          spaceDomains,
		SecurityGroups:   securityGroups,
		SpaceQuota:       spaceQuota,
	}
	return nil
}

func (cmd *CliRpcCmd) GetService(serviceInstance string, retVal *plugin_models.GetService_Model) error {
	if cmd.actor == nil {
		return fmt.Errorf("v7 actor not initialized")
	}
	// Get current space GUID from config
	spaceGUID := cmd.cliConfig.SpaceFields().GUID
	if spaceGUID == "" {
		return fmt.Errorf("no space targeted")
	}

	// Get service instance details (omitApps=false to include bound apps)
	details, _, err := cmd.actor.GetServiceInstanceDetails(serviceInstance, spaceGUID, false)
	if err != nil {
		return err
	}

	// Map to plugin model
	dashboardURL := ""
	if details.DashboardURL.IsSet {
		dashboardURL = details.DashboardURL.Value
	}

	model := plugin_models.GetService_Model{
		Guid:           details.GUID,
		Name:           details.Name,
		DashboardUrl:   dashboardURL,
		IsUserProvided: details.Type == "user-provided",
		ServiceOffering: plugin_models.GetService_ServiceFields{
			Name:             details.ServiceOffering.Name,
			DocumentationUrl: details.ServiceOffering.Description, // Use description as doc URL not available
		},
		ServicePlan: plugin_models.GetService_ServicePlan{
			Name: details.ServicePlan.Name,
			Guid: details.ServicePlan.GUID,
		},
		LastOperation: plugin_models.GetService_LastOperation{
			Type:        string(details.LastOperation.Type),
			State:       string(details.LastOperation.State),
			Description: details.LastOperation.Description,
			CreatedAt:   details.LastOperation.CreatedAt,
			UpdatedAt:   details.LastOperation.UpdatedAt,
		},
	}

	*retVal = model
	return nil
}
