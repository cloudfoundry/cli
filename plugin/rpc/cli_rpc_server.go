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

	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/plugin"
	plugin_models "code.cloudfoundry.org/cli/v8/plugin/models"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"
	"code.cloudfoundry.org/cli/v8/version"
	"github.com/blang/semver/v4"
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
	PluginMetadata *plugin.PluginMetadata
	MetadataMutex  *sync.RWMutex
	outputDisabled bool
	cliConfig      *configv3.Config
	outputBucket   *bytes.Buffer
	actor          *v7action.Actor
	commandParser  CommandParser
	commandUI      *ui.UI
}

type OutputInterceptor struct {
	capturedOutput *bytes.Buffer
	stdout         io.Writer
}

func (oi *OutputInterceptor) Write(p []byte) (n int, err error) {
	n, err = oi.capturedOutput.Write(p)
	if oi.stdout != nil {
		_, _ = oi.stdout.Write(p)
	}
	return n, err
}

func NewRpcService(
	cliConfig *configv3.Config,
	rpcServer *rpc.Server,
	actor *v7action.Actor,
	commandParser CommandParser,
	commandUI *ui.UI,
) (*CliRpcService, error) {
	rpcService := &CliRpcService{
		Server: rpcServer,
		RpcCmd: &CliRpcCmd{
			PluginMetadata: &plugin.PluginMetadata{},
			MetadataMutex:  &sync.RWMutex{},
			cliConfig:      cliConfig,
			actor:          actor,
			commandParser:  commandParser,
			commandUI:      commandUI,
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
	cmd.outputDisabled = disable
	*retVal = true
	return nil
}

func (cmd *CliRpcCmd) CallCoreCommand(args []string, retVal *bool) error {
	cmd.outputBucket = &bytes.Buffer{}
	stdout := &OutputInterceptor{
		capturedOutput: cmd.outputBucket,
	}
	stderr := stdout
	if !cmd.outputDisabled && cmd.commandUI != nil {
		stdout.stdout = cmd.commandUI.GetOut()
		stderr = &OutputInterceptor{
			capturedOutput: cmd.outputBucket,
			stdout:         cmd.commandUI.GetErr(),
		}
	}

	// capture output for plugin
	pluginUI, err := ui.NewPluginUI(cmd.cliConfig, stdout, stderr)
	if err != nil {
		*retVal = false
		return err
	}

	// Use the command parser to execute the command
	exitCode, err := cmd.commandParser.ParseCommandFromArgs(pluginUI, args)
	if err != nil {
		*retVal = false
		return err
	}

	if exitCode != 0 {
		*retVal = false
		return fmt.Errorf("command exited with code %d", exitCode)
	}

	*retVal = true
	return nil
}

func (cmd *CliRpcCmd) GetOutputAndReset(args bool, retVal *[]string) error {
	v := strings.TrimSuffix(cmd.outputBucket.String(), "\n")
	*retVal = strings.Split(v, "\n")
	return nil
}

// displayWarnings displays warning messages using commandUI
func (cmd *CliRpcCmd) displayWarnings(warnings []string) {
	if cmd.commandUI != nil {
		for _, warning := range warnings {
			cmd.commandUI.DisplayTextLiteral(fmt.Sprintf("Warning: %s", warning))
		}
	}
}

func (cmd *CliRpcCmd) GetCurrentOrg(args string, retVal *plugin_models.Organization) error {
	org := cmd.cliConfig.TargetedOrganization()
	retVal.Name = org.Name
	retVal.Guid = org.GUID
	return nil
}

func (cmd *CliRpcCmd) GetCurrentSpace(args string, retVal *plugin_models.Space) error {
	space := cmd.cliConfig.TargetedSpace()
	retVal.Name = space.Name
	retVal.Guid = space.GUID

	return nil
}

func (cmd *CliRpcCmd) Username(args string, retVal *string) error {
	username, err := cmd.cliConfig.CurrentUserName()
	if err != nil {
		return err
	}
	*retVal = username

	return nil
}

func (cmd *CliRpcCmd) UserGuid(args string, retVal *string) error {
	user, err := cmd.cliConfig.CurrentUser()
	if err != nil {
		return err
	}
	*retVal = user.GUID

	return nil
}

func (cmd *CliRpcCmd) UserEmail(args string, retVal *string) error {
	user, err := cmd.cliConfig.CurrentUser()
	if err != nil {
		*retVal = ""
		return nil
	}
	*retVal = user.Name // For UAA users, Name contains the email

	return nil
}

func (cmd *CliRpcCmd) IsLoggedIn(args string, retVal *bool) error {
	*retVal = cmd.cliConfig.AccessToken() != ""

	return nil
}

func (cmd *CliRpcCmd) IsSSLDisabled(args string, retVal *bool) error {
	*retVal = cmd.cliConfig.SkipSSLValidation()

	return nil
}

func (cmd *CliRpcCmd) HasOrganization(args string, retVal *bool) error {
	*retVal = cmd.cliConfig.HasTargetedOrganization()

	return nil
}

func (cmd *CliRpcCmd) HasSpace(args string, retVal *bool) error {
	*retVal = cmd.cliConfig.HasTargetedSpace()

	return nil
}

func (cmd *CliRpcCmd) ApiEndpoint(args string, retVal *string) error {
	*retVal = cmd.cliConfig.Target()

	return nil
}

func (cmd *CliRpcCmd) HasAPIEndpoint(args string, retVal *bool) error {
	*retVal = cmd.cliConfig.Target() != ""

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
	*retVal = cmd.cliConfig.ConfigFile.DopplerEndpoint

	return nil
}

func (cmd *CliRpcCmd) AccessToken(args string, retVal *string) error {
	token, err := cmd.actor.RefreshAccessToken()
	if err != nil {
		return err
	}

	*retVal = token

	return nil
}

func (cmd *CliRpcCmd) GetApp(appName string, retVal *plugin_models.GetAppModel) error {
	spaceGUID := cmd.cliConfig.TargetedSpace().GUID
	if spaceGUID == "" {
		return fmt.Errorf("no space targeted")
	}

	// Get detailed app summary
	summary, warnings, err := cmd.actor.GetDetailedAppSummary(appName, spaceGUID, false)
	if err != nil {
		return err
	}

	cmd.displayWarnings(warnings)

	// Get service bindings for the app
	serviceBindings, ccWarnings, err := cmd.actor.CloudControllerClient.GetServiceCredentialBindings(
		ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{summary.GUID}},
	)
	cmd.displayWarnings(ccWarnings)
	if err != nil {
		return err
	}

	// Get stack information
	var stack resources.Stack
	if summary.CurrentDroplet.Stack != "" {
		stacks, ccWarnings, err := cmd.actor.CloudControllerClient.GetStacks(
			ccv3.Query{Key: ccv3.NameFilter, Values: []string{summary.CurrentDroplet.Stack}},
		)
		cmd.displayWarnings(ccWarnings)
		if err != nil {
			return err
		}
		if len(stacks) > 0 {
			stack = stacks[0]
		}
	}

	// Populate the plugin model
	*retVal = populateAppModel(summary, serviceBindings, stack)
	return nil
}

func (cmd *CliRpcCmd) GetApps(_ string, retVal *[]plugin_models.GetAppsModel) error {
	spaceGUID := cmd.cliConfig.TargetedSpace().GUID
	if spaceGUID == "" {
		return fmt.Errorf("no space targeted")
	}

	// Get app summaries for the space (omitStats=false to get instances)
	summaries, warnings, err := cmd.actor.GetAppSummariesForSpace(spaceGUID, "", false)
	if err != nil {
		return err
	}

	cmd.displayWarnings(warnings)

	// Populate plugin model using mapping function
	*retVal = populateAppsModel(summaries)

	return nil
}

func (cmd *CliRpcCmd) GetOrgs(_ string, retVal *[]plugin_models.GetOrgs_Model) error {
	orgs, warnings, err := cmd.actor.GetOrganizations("")
	if err != nil {
		return err
	}

	cmd.displayWarnings(warnings)

	// Populate plugin model using mapping function
	*retVal = populateOrgsModel(orgs)
	return nil
}

func (cmd *CliRpcCmd) GetSpaces(_ string, retVal *[]plugin_models.GetSpaces_Model) error {
	orgGUID := cmd.cliConfig.TargetedOrganization().GUID
	if orgGUID == "" {
		return fmt.Errorf("no organization targeted")
	}

	spaces, warnings, err := cmd.actor.GetOrganizationSpaces(orgGUID)
	if err != nil {
		return err
	}

	cmd.displayWarnings(warnings)

	// Populate plugin model using mapping function
	*retVal = populateSpacesModel(spaces)
	return nil
}

func (cmd *CliRpcCmd) GetServices(_ string, retVal *[]plugin_models.GetServices_Model) error {
	spaceGUID := cmd.cliConfig.TargetedSpace().GUID
	if spaceGUID == "" {
		return fmt.Errorf("no space targeted")
	}

	services, warnings, err := cmd.actor.GetServiceInstancesForSpace(spaceGUID, false)
	if err != nil {
		return err
	}

	cmd.displayWarnings(warnings)

	// Populate plugin model using mapping function
	*retVal = populateServicesModel(services)
	return nil
}

func (cmd *CliRpcCmd) GetOrgUsers(args []string, retVal *[]plugin_models.GetOrgUsers_Model) error {
	if len(args) == 0 {
		return fmt.Errorf("organization name required")
	}
	orgName := args[0]

	// Get org GUID from name
	org, warnings, err := cmd.actor.GetOrganizationByName(orgName)
	if err != nil {
		return err
	}

	// Handle warnings from org lookup
	cmd.displayWarnings(warnings)

	// Get users by role type
	usersByRole, warnings, err := cmd.actor.GetOrgUsersByRoleType(org.GUID)
	if err != nil {
		return err
	}

	cmd.displayWarnings(warnings)

	// Populate plugin model using mapping function
	*retVal = populateOrgUsersModel(usersByRole)
	return nil
}

func (cmd *CliRpcCmd) GetSpaceUsers(args []string, retVal *[]plugin_models.GetSpaceUsers_Model) error {
	if len(args) < 2 {
		return fmt.Errorf("organization and space names required")
	}
	orgName := args[0]
	spaceName := args[1]

	// Get org GUID from name
	org, warnings, err := cmd.actor.GetOrganizationByName(orgName)
	if err != nil {
		return err
	}

	// Handle warnings from org lookup
	cmd.displayWarnings(warnings)

	// Get space GUID from name
	space, warnings, err := cmd.actor.GetSpaceByNameAndOrganization(spaceName, org.GUID)
	if err != nil {
		return err
	}

	// Handle warnings from space lookup
	cmd.displayWarnings(warnings)

	// Get users by role type
	usersByRole, warnings, err := cmd.actor.GetSpaceUsersByRoleType(space.GUID)
	if err != nil {
		return err
	}

	cmd.displayWarnings(warnings)

	// Populate plugin model using mapping function
	*retVal = populateSpaceUsersModel(usersByRole)
	return nil
}

func (cmd *CliRpcCmd) GetOrg(orgName string, retVal *plugin_models.GetOrg_Model) error {
	// 1. Get the organization by name
	org, warnings, err := cmd.actor.GetOrganizationByName(orgName)
	if err != nil {
		return err
	}

	cmd.displayWarnings(warnings)

	// 2. Get organization quota
	var quota resources.OrganizationQuota
	if org.QuotaGUID != "" {
		var ccv3Warnings ccv3.Warnings
		quota, ccv3Warnings, err = cmd.actor.CloudControllerClient.GetOrganizationQuota(org.QuotaGUID)
		if err != nil {
			return err
		}
		cmd.displayWarnings(ccv3Warnings)
	}

	// 3. Get organization spaces
	spaces, warnings, err := cmd.actor.GetOrganizationSpaces(org.GUID)
	if err != nil {
		return err
	}
	cmd.displayWarnings(warnings)

	// 4. Get organization domains
	domains, warnings, err := cmd.actor.GetOrganizationDomains(org.GUID, "")
	if err != nil {
		return err
	}
	cmd.displayWarnings(warnings)

	// 5. Get space quotas for the organization
	spaceQuotas, warnings, err := cmd.actor.GetSpaceQuotasByOrgGUID(org.GUID)
	if err != nil {
		return err
	}
	cmd.displayWarnings(warnings)

	// Populate plugin model using mapping function
	*retVal = populateOrgModel(org, quota, spaces, domains, spaceQuotas)
	return nil
}

func (cmd *CliRpcCmd) GetSpace(spaceName string, retVal *plugin_models.GetSpace_Model) error {
	// Get the targeted organization GUID
	orgGUID := cmd.cliConfig.TargetedOrganization().GUID
	if orgGUID == "" {
		return fmt.Errorf("no organization targeted")
	}

	// 1. Get the space by name and organization
	space, warnings, err := cmd.actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	if err != nil {
		return err
	}
	cmd.displayWarnings(warnings)

	// 2. Get organization info
	org, warnings, err := cmd.actor.GetOrganizationByGUID(orgGUID)
	if err != nil {
		return err
	}
	cmd.displayWarnings(warnings)

	// 3. Get applications in the space
	apps, warnings, err := cmd.actor.GetApplicationsBySpace(space.GUID)
	if err != nil {
		return err
	}
	cmd.displayWarnings(warnings)

	// 4. Get service instances in the space
	serviceInstances, warnings, err := cmd.actor.GetServiceInstancesForSpace(space.GUID, false)
	if err != nil {
		return err
	}
	cmd.displayWarnings(warnings)

	// 5. Get domains for the organization (spaces inherit org domains)
	domains, warnings, err := cmd.actor.GetOrganizationDomains(orgGUID, "")
	if err != nil {
		return err
	}
	cmd.displayWarnings(warnings)

	// 6. Get space quota (if applied)
	var spaceQuota resources.SpaceQuota
	if space.Relationships[constant.RelationshipTypeQuota].GUID != "" {
		var ccv3Warnings ccv3.Warnings
		spaceQuota, ccv3Warnings, err = cmd.actor.CloudControllerClient.GetSpaceQuota(space.Relationships[constant.RelationshipTypeQuota].GUID)
		if err != nil {
			return err
		}
		cmd.displayWarnings(ccv3Warnings)
	}

	// 7. Get security groups (both running and staging)
	var allSecurityGroups []resources.SecurityGroup

	runningSecurityGroups, ccv3Warnings, err := cmd.actor.CloudControllerClient.GetRunningSecurityGroups(space.GUID)
	if err != nil {
		return err
	}
	cmd.displayWarnings(ccv3Warnings)
	allSecurityGroups = append(allSecurityGroups, runningSecurityGroups...)

	stagingSecurityGroups, ccv3Warnings, err := cmd.actor.CloudControllerClient.GetStagingSecurityGroups(space.GUID)
	if err != nil {
		return err
	}
	cmd.displayWarnings(ccv3Warnings)

	// Deduplicate security groups (some may be both running and staging)
	seenGroups := make(map[string]bool)
	for _, sg := range stagingSecurityGroups {
		if !seenGroups[sg.GUID] {
			allSecurityGroups = append(allSecurityGroups, sg)
			seenGroups[sg.GUID] = true
		}
	}

	// Populate plugin model using mapping function
	*retVal = populateSpaceModel(space, orgGUID, org.Name, apps, serviceInstances, domains, spaceQuota, allSecurityGroups)
	return nil
}

func (cmd *CliRpcCmd) GetService(serviceInstanceName string, retVal *plugin_models.GetService_Model) error {
	spaceGUID := cmd.cliConfig.TargetedSpace().GUID
	if spaceGUID == "" {
		return fmt.Errorf("no space targeted")
	}

	// Query to include service plan and service offering details
	queries := []ccv3.Query{
		{Key: ccv3.FieldsServicePlan, Values: []string{"name", "guid"}},
		{Key: ccv3.FieldsServicePlanServiceOffering, Values: []string{"name", "guid", "documentation_url"}},
	}

	serviceInstance, includedResources, warnings, err := cmd.actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(
		serviceInstanceName,
		spaceGUID,
		queries...,
	)
	if err != nil {
		return err
	}

	// Log warnings if any
	cmd.displayWarnings(warnings)

	// Populate the plugin model
	populateServiceModel(retVal, serviceInstance, includedResources)
	return nil
}
