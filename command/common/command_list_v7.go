// +build V7

package common

import (
	"reflect"

	"code.cloudfoundry.org/cli/command/plugin"
	v6 "code.cloudfoundry.org/cli/command/v6"
	v7 "code.cloudfoundry.org/cli/command/v7"
)

var Commands commandList
var FallbackCommands V2CommandList

type V2CommandList struct {
	V2App v6.V3AppCommand `command:"app" description:"Display health and status for an app"`
}

type commandList struct {
	VerboseOrVersion bool `short:"v" long:"version" description:"verbose and version flag"`

	V3CancelZdtPush v6.V3CancelZdtPushCommand       `command:"v3-cancel-zdt-push" description:"Cancel the most recent deployment for an app"`
	V3ZdtRestart    v6.V3ZeroDowntimeRestartCommand `command:"v3-zdt-restart" description:"Sequentially restart each instance of an app."`

	V3Push v7.PushCommand `command:"v3-push" description:"Push a new app or sync changes to an existing app" hidden:"true"`

	API                                v7.APICommand                                `command:"api" description:"Set or view target api url"`
	AddNetworkPolicy                   v6.AddNetworkPolicyCommand                   `command:"add-network-policy" description:"Create policy to allow direct network traffic from one app to another"`
	AddPluginRepo                      plugin.AddPluginRepoCommand                  `command:"add-plugin-repo" description:"Add a new plugin repository"`
	AllowSpaceSSH                      v7.AllowSpaceSSHCommand                      `command:"allow-space-ssh" description:"Allow SSH access for the space"`
	App                                v7.AppCommand                                `command:"app" description:"Display health and status for an app"`
	ApplyManifest                      v7.ApplyManifestCommand                      `command:"apply-manifest" description:"Apply manifest properties to a space"`
	Apps                               v7.AppsCommand                               `command:"apps" alias:"a" description:"List all apps in the target space"`
	Auth                               v7.AuthCommand                               `command:"auth" description:"Authenticate non-interactively"`
	BindRouteService                   v6.BindRouteServiceCommand                   `command:"bind-route-service" alias:"brs" description:"Bind a service instance to an HTTP route"`
	BindRunningSecurityGroup           v6.BindRunningSecurityGroupCommand           `command:"bind-running-security-group" description:"Bind a security group to the list of security groups to be used for running applications"`
	BindSecurityGroup                  v6.BindSecurityGroupCommand                  `command:"bind-security-group" description:"Bind a security group to a particular space, or all existing spaces of an org"`
	BindService                        v6.BindServiceCommand                        `command:"bind-service" alias:"bs" description:"Bind a service instance to an app"`
	BindStagingSecurityGroup           v6.BindStagingSecurityGroupCommand           `command:"bind-staging-security-group" description:"Bind a security group to the list of security groups to be used for staging applications"`
	Buildpacks                         v7.BuildpacksCommand                         `command:"buildpacks" description:"List all buildpacks"`
	CancelDeployment                   v7.CancelDeploymentCommand                   `command:"cancel-deployment" description:"Cancel the most recent deployment for an app. Resets the current droplet to the previous deployment's droplet."`
	CheckRoute                         v7.CheckRouteCommand                         `command:"check-route" description:"Perform a check to determine whether a route currently exists or not"`
	Config                             v6.ConfigCommand                             `command:"config" description:"Write default values to the config"`
	CopySource                         v6.CopySourceCommand                         `command:"copy-source" description:"Copies the source code of an application to another existing application (and restarts that application)"`
	CreateApp                          v7.CreateAppCommand                          `command:"create-app" description:"Create an Application in the target space"`
	CreateAppManifest                  v7.CreateAppManifestCommand                  `command:"create-app-manifest" description:"Create an app manifest for an app that has been pushed successfully"`
	CreateBuildpack                    v7.CreateBuildpackCommand                    `command:"create-buildpack" description:"Create a buildpack"`
	CreatePackage                      v7.CreatePackageCommand                      `command:"create-package" description:"Uploads a Package"`
	CreateIsolationSegment             v6.CreateIsolationSegmentCommand             `command:"create-isolation-segment" description:"Create an isolation segment"`
	CreateOrg                          v7.CreateOrgCommand                          `command:"create-org" alias:"co" description:"Create an org"`
	CreateOrgQuota                     v7.CreateOrgQuotaCommand                     `command:"create-org-quota" description:"Define a new quota for an organization"`
	CreatePrivateDomain                v7.CreatePrivateDomainCommand                `command:"create-private-domain" description:"Create a private domain for a specific org"`
	CreateRoute                        v7.CreateRouteCommand                        `command:"create-route" description:"Create a route for later use"`
	CreateSecurityGroup                v6.CreateSecurityGroupCommand                `command:"create-security-group" description:"Create a security group"`
	CreateService                      v6.CreateServiceCommand                      `command:"create-service" alias:"cs" description:"Create a service instance"`
	CreateServiceBroker                v7.CreateServiceBrokerCommand                `command:"create-service-broker" alias:"csb" description:"Create a service broker"`
	CreateServiceKey                   v6.CreateServiceKeyCommand                   `command:"create-service-key" alias:"csk" description:"Create key for a service instance"`
	CreateSharedDomain                 v7.CreateSharedDomainCommand                 `command:"create-shared-domain" description:"Create a domain that can be used by all orgs (admin-only)"`
	CreateSpace                        v7.CreateSpaceCommand                        `command:"create-space" alias:"csp" description:"Create a space"`
	CreateSpaceQuota                   v7.CreateSpaceQuotaCommand                   `command:"create-space-quota" description:"Define a new quota for a space"`
	CreateUser                         v7.CreateUserCommand                         `command:"create-user" description:"Create a new user"`
	CreateUserProvidedService          v6.CreateUserProvidedServiceCommand          `command:"create-user-provided-service" alias:"cups" description:"Make a user-provided service instance available to CF apps"`
	Curl                               v6.CurlCommand                               `command:"curl" description:"Executes a request to the targeted API endpoint"`
	Delete                             v7.DeleteCommand                             `command:"delete" alias:"d" description:"Delete an app"`
	DeleteBuildpack                    v7.DeleteBuildpackCommand                    `command:"delete-buildpack" description:"Delete a buildpack"`
	DeleteIsolationSegment             v6.DeleteIsolationSegmentCommand             `command:"delete-isolation-segment" description:"Delete an isolation segment"`
	DeleteOrg                          v7.DeleteOrgCommand                          `command:"delete-org" description:"Delete an org"`
	DeleteOrgQuota                     v7.DeleteOrgQuotaCommand                     `command:"delete-org-quota" description:"Delete an organization quota"`
	DeleteOrphanedRoutes               v7.DeleteOrphanedRoutesCommand               `command:"delete-orphaned-routes" description:"Delete all orphaned routes in the currently targeted space (i.e. those that are not mapped to an app or service instance)"`
	DeletePrivateDomain                v7.DeletePrivateDomainCommand                `command:"delete-private-domain" description:"Delete a private domain"`
	DeleteRoute                        v7.DeleteRouteCommand                        `command:"delete-route" description:"Delete a route"`
	DeleteSecurityGroup                v6.DeleteSecurityGroupCommand                `command:"delete-security-group" description:"Deletes a security group"`
	DeleteService                      v6.DeleteServiceCommand                      `command:"delete-service" alias:"ds" description:"Delete a service instance"`
	DeleteServiceBroker                v7.DeleteServiceBrokerCommand                `command:"delete-service-broker" description:"Delete a service broker"`
	DeleteServiceKey                   v6.DeleteServiceKeyCommand                   `command:"delete-service-key" alias:"dsk" description:"Delete a service key"`
	DeleteSharedDomain                 v7.DeleteSharedDomainCommand                 `command:"delete-shared-domain" description:"Delete a shared domain"`
	DeleteSpace                        v7.DeleteSpaceCommand                        `command:"delete-space" description:"Delete a space"`
	DeleteSpaceQuota                   v7.DeleteSpaceQuotaCommand                   `command:"delete-space-quota" description:"Delete a space quota"`
	DeleteUser                         v7.DeleteUserCommand                         `command:"delete-user" description:"Delete a user"`
	DisableFeatureFlag                 v7.DisableFeatureFlagCommand                 `command:"disable-feature-flag" description:"Prevent use of a feature"`
	DisableOrgIsolation                v6.DisableOrgIsolationCommand                `command:"disable-org-isolation" description:"Revoke an organization's entitlement to an isolation segment"`
	DisableSSH                         v7.DisableSSHCommand                         `command:"disable-ssh" description:"Disable ssh for the application"`
	DisableServiceAccess               v6.DisableServiceAccessCommand               `command:"disable-service-access" description:"Disable access to a service or service plan for one or all orgs"`
	DisallowSpaceSSH                   v7.DisallowSpaceSSHCommand                   `command:"disallow-space-ssh" description:"Disallow SSH access for the space"`
	Domains                            v7.DomainsCommand                            `command:"domains" description:"List domains in the target org"`
	Droplets                           v7.DropletsCommand                           `command:"droplets" description:"List droplets of an app"`
	EnableFeatureFlag                  v7.EnableFeatureFlagCommand                  `command:"enable-feature-flag" description:"Allow use of a feature"`
	EnableOrgIsolation                 v6.EnableOrgIsolationCommand                 `command:"enable-org-isolation" description:"Entitle an organization to an isolation segment"`
	EnableSSH                          v7.EnableSSHCommand                          `command:"enable-ssh" description:"Enable ssh for the application"`
	EnableServiceAccess                v6.EnableServiceAccessCommand                `command:"enable-service-access" description:"Enable access to a service or service plan for one or all orgs"`
	Env                                v7.EnvCommand                                `command:"env" alias:"e" description:"Show all env variables for an app"`
	Events                             v7.EventsCommand                             `command:"events" description:"Show recent app events"`
	FeatureFlag                        v7.FeatureFlagCommand                        `command:"feature-flag" description:"Retrieve an individual feature flag with status"`
	FeatureFlags                       v7.FeatureFlagsCommand                       `command:"feature-flags" description:"Retrieve list of feature flags with status"`
	GetHealthCheck                     v7.GetHealthCheckCommand                     `command:"get-health-check" description:"Show the type of health check performed on an app"`
	Help                               HelpCommand                                  `command:"help" alias:"h" description:"Show help"`
	InstallPlugin                      InstallPluginCommand                         `command:"install-plugin" description:"Install CLI plugin"`
	IsolationSegments                  v6.IsolationSegmentsCommand                  `command:"isolation-segments" description:"List all isolation segments"`
	Labels                             v7.LabelsCommand                             `command:"labels" description:"List all labels (key-value pairs) for an API resource"`
	ListPluginRepos                    plugin.ListPluginReposCommand                `command:"list-plugin-repos" description:"List all the added plugin repositories"`
	Login                              v6.LoginCommand                              `command:"login" alias:"l" description:"Log user in"`
	Logout                             v6.LogoutCommand                             `command:"logout" alias:"lo" description:"Log user out"`
	Logs                               v7.LogsCommand                               `command:"logs" description:"Tail or show recent logs for an app"`
	MapRoute                           v7.MapRouteCommand                           `command:"map-route" description:"Map a route to an app"`
	Marketplace                        v6.MarketplaceCommand                        `command:"marketplace" alias:"m" description:"List available offerings in the marketplace"`
	NetworkPolicies                    v6.NetworkPoliciesCommand                    `command:"network-policies" description:"List direct network traffic policies"`
	OauthToken                         v6.OauthTokenCommand                         `command:"oauth-token" description:"Retrieve and display the OAuth token for the current session"`
	Org                                v7.OrgCommand                                `command:"org" description:"Show org info"`
	OrgQuotas                          v7.OrgQuotasCommand                          `command:"org-quotas" description:"List available organization quotas"`
	OrgQuota                           v7.OrgQuotaCommand                           `command:"org-quota" description:"Show organization quota"`
	OrgUsers                           v7.OrgUsersCommand                           `command:"org-users" description:"Show org users by role"`
	Orgs                               v7.OrgsCommand                               `command:"orgs" alias:"o" description:"List all orgs"`
	Packages                           v7.PackagesCommand                           `command:"packages" description:"List packages of an app"`
	Passwd                             v6.PasswdCommand                             `command:"passwd" alias:"pw" description:"Change user password"`
	Plugins                            plugin.PluginsCommand                        `command:"plugins" description:"List commands of installed plugins"`
	PurgeServiceInstance               v6.PurgeServiceInstanceCommand               `command:"purge-service-instance" description:"Recursively remove a service instance and child objects from Cloud Foundry database without making requests to a service broker"`
	PurgeServiceOffering               v6.PurgeServiceOfferingCommand               `command:"purge-service-offering" description:"Recursively remove a service and child objects from Cloud Foundry database without making requests to a service broker"`
	Push                               v7.PushCommand                               `command:"push" alias:"p" description:"Push a new app or sync changes to an existing app"`
	RemoveNetworkPolicy                v6.RemoveNetworkPolicyCommand                `command:"remove-network-policy" description:"Remove network traffic policy of an app"`
	RemovePluginRepo                   plugin.RemovePluginRepoCommand               `command:"remove-plugin-repo" description:"Remove a plugin repository"`
	Rename                             v7.RenameCommand                             `command:"rename" description:"Rename an app"`
	RenameBuildpack                    v6.RenameBuildpackCommand                    `command:"rename-buildpack" description:"Rename a buildpack"`
	RenameOrg                          v7.RenameOrgCommand                          `command:"rename-org" description:"Rename an org"`
	RenameService                      v6.RenameServiceCommand                      `command:"rename-service" description:"Rename a service instance"`
	RenameServiceBroker                v7.RenameServiceBrokerCommand                `command:"rename-service-broker" description:"Rename a service broker"`
	RenameSpace                        v7.RenameSpaceCommand                        `command:"rename-space" description:"Rename a space"`
	RepoPlugins                        plugin.RepoPluginsCommand                    `command:"repo-plugins" description:"List all available plugins in specified repository or in all added repositories"`
	ResetOrgDefaultIsolationSegment    v6.ResetOrgDefaultIsolationSegmentCommand    `command:"reset-org-default-isolation-segment" description:"Reset the default isolation segment used for apps in spaces of an org"`
	ResetSpaceIsolationSegment         v6.ResetSpaceIsolationSegmentCommand         `command:"reset-space-isolation-segment" description:"Reset the space's isolation segment to the org default"`
	Restage                            v7.RestageCommand                            `command:"restage" alias:"rg" description:"Recreate the app's executable artifact using the latest pushed app files and the latest environment (variables, service bindings, buildpack, stack, etc.)."`
	Stage                              v7.StageCommand                              `command:"stage" description:"Create a new droplet for an app, defaults to the newest package"`
	Restart                            v7.RestartCommand                            `command:"restart" alias:"rs" description:"Stop all instances of the app, then start them again. This causes downtime."`
	RestartAppInstance                 v7.RestartAppInstanceCommand                 `command:"restart-app-instance" description:"Terminate, then instantiate an app instance"`
	RouterGroups                       v6.RouterGroupsCommand                       `command:"router-groups" description:"List router groups"`
	Routes                             v7.RoutesCommand                             `command:"routes" alias:"r" description:"List all routes in the current space or the current organization"`
	RunTask                            v7.RunTaskCommand                            `command:"run-task" alias:"rt" description:"Run a one-off task on an app"`
	RunningEnvironmentVariableGroup    v7.RunningEnvironmentVariableGroupCommand    `command:"running-environment-variable-group" alias:"revg" description:"Retrieve the contents of the running environment variable group"`
	RunningSecurityGroups              v6.RunningSecurityGroupsCommand              `command:"running-security-groups" description:"List security groups in the set of security groups for running applications"`
	SSH                                v7.SSHCommand                                `command:"ssh" description:"SSH to an application container instance"`
	SSHCode                            v6.SSHCodeCommand                            `command:"ssh-code" description:"Get a one time password for ssh clients"`
	SSHEnabled                         v7.SSHEnabledCommand                         `command:"ssh-enabled" description:"Reports whether SSH is enabled on an application container instance"`
	Scale                              v7.ScaleCommand                              `command:"scale" description:"Change or view the instance count, disk space limit, and memory limit for an app"`
	SecurityGroup                      v6.SecurityGroupCommand                      `command:"security-group" description:"Show a single security group"`
	SecurityGroups                     v6.SecurityGroupsCommand                     `command:"security-groups" description:"List all security groups"`
	Service                            v6.ServiceCommand                            `command:"service" description:"Show service instance info"`
	ServiceAccess                      v6.ServiceAccessCommand                      `command:"service-access" description:"List service access settings"`
	ServiceBrokers                     v7.ServiceBrokersCommand                     `command:"service-brokers" description:"List service brokers"`
	ServiceKey                         v6.ServiceKeyCommand                         `command:"service-key" description:"Show service key info"`
	ServiceKeys                        v6.ServiceKeysCommand                        `command:"service-keys" alias:"sk" description:"List keys for a service instance"`
	Services                           v6.ServicesCommand                           `command:"services" alias:"s" description:"List all service instances in the target space"`
	SetDroplet                         v7.SetDropletCommand                         `command:"set-droplet" description:"Set the droplet used to run an app"`
	SetEnv                             v7.SetEnvCommand                             `command:"set-env" alias:"se" description:"Set an env variable for an app"`
	SetHealthCheck                     v7.SetHealthCheckCommand                     `command:"set-health-check" description:"Change type of health check performed on an app's process"`
	SetLabel                           v7.SetLabelCommand                           `command:"set-label" description:"Set a label (key-value pairs) for an API resource"`
	SetOrgDefaultIsolationSegment      v6.SetOrgDefaultIsolationSegmentCommand      `command:"set-org-default-isolation-segment" description:"Set the default isolation segment used for apps in spaces in an org"`
	SetOrgRole                         v7.SetOrgRoleCommand                         `command:"set-org-role" description:"Assign an org role to a user"`
	SetOrgQuota                        v7.SetOrgQuotaCommand                        `command:"set-org-quota" description:"Assign a quota to an organization"`
	SetRunningEnvironmentVariableGroup v7.SetRunningEnvironmentVariableGroupCommand `command:"set-running-environment-variable-group" alias:"srevg" description:"Pass parameters as JSON to create a running environment variable group"`
	SetSpaceIsolationSegment           v6.SetSpaceIsolationSegmentCommand           `command:"set-space-isolation-segment" description:"Assign the isolation segment for a space"`
	SetSpaceQuota                      v7.SetSpaceQuotaCommand                      `command:"set-space-quota" description:"Assign a quota to a space"`
	SetSpaceRole                       v7.SetSpaceRoleCommand                       `command:"set-space-role" description:"Assign a space role to a user"`
	SetStagingEnvironmentVariableGroup v7.SetStagingEnvironmentVariableGroupCommand `command:"set-staging-environment-variable-group" alias:"ssevg" description:"Pass parameters as JSON to create a staging environment variable group"`
	SharePrivateDomain                 v7.SharePrivateDomainCommand                 `command:"share-private-domain" description:"Share a private domain with a specific org"`
	ShareService                       v6.ShareServiceCommand                       `command:"share-service" description:"Share a service instance with another space"`
	Space                              v7.SpaceCommand                              `command:"space" description:"Show space info"`
	SpaceQuota                         v7.SpaceQuotaCommand                         `command:"space-quota" description:"Show space quota info"`
	SpaceQuotas                        v7.SpaceQuotasCommand                        `command:"space-quotas" description:"List available space quotas"`
	SpaceSSHAllowed                    v7.SpaceSSHAllowedCommand                    `command:"space-ssh-allowed" description:"Reports whether SSH is allowed in a space"`
	SpaceUsers                         v7.SpaceUsersCommand                         `command:"space-users" description:"Show space users by role"`
	Spaces                             v7.SpacesCommand                             `command:"spaces" description:"List all spaces in an org"`
	Stack                              v7.StackCommand                              `command:"stack" description:"Show information for a stack (a stack is a pre-built file system, including an operating system, that can run apps)"`
	Stacks                             v7.StacksCommand                             `command:"stacks" description:"List all stacks (a stack is a pre-built file system, including an operating system, that can run apps)"`
	StagingEnvironmentVariableGroup    v7.StagingEnvironmentVariableGroupCommand    `command:"staging-environment-variable-group" alias:"sevg" description:"Retrieve the contents of the staging environment variable group"`
	StagingSecurityGroups              v6.StagingSecurityGroupsCommand              `command:"staging-security-groups" description:"List security groups in the staging set for applications"`
	Start                              v7.StartCommand                              `command:"start" alias:"st" description:"Start an app"`
	Stop                               v7.StopCommand                               `command:"stop" alias:"sp" description:"Stop an app"`
	Target                             v7.TargetCommand                             `command:"target" alias:"t" description:"Set or view the targeted org or space"`
	Tasks                              v6.TasksCommand                              `command:"tasks" description:"List tasks of an app"`
	TerminateTask                      v6.TerminateTaskCommand                      `command:"terminate-task" description:"Terminate a running task of an app"`
	UnbindRouteService                 v6.UnbindRouteServiceCommand                 `command:"unbind-route-service" alias:"urs" description:"Unbind a service instance from an HTTP route"`
	UnbindRunningSecurityGroup         v6.UnbindRunningSecurityGroupCommand         `command:"unbind-running-security-group" description:"Unbind a security group from the set of security groups for running applications"`
	UnbindSecurityGroup                v6.UnbindSecurityGroupCommand                `command:"unbind-security-group" description:"Unbind a security group from a space"`
	UnbindService                      v6.UnbindServiceCommand                      `command:"unbind-service" alias:"us" description:"Unbind a service instance from an app"`
	UnbindStagingSecurityGroup         v6.UnbindStagingSecurityGroupCommand         `command:"unbind-staging-security-group" description:"Unbind a security group from the set of security groups for staging applications"`
	UninstallPlugin                    plugin.UninstallPluginCommand                `command:"uninstall-plugin" description:"Uninstall CLI plugin"`
	UnmapRoute                         v7.UnmapRouteCommand                         `command:"unmap-route" description:"Remove a route from an app"`
	UnsetEnv                           v7.UnsetEnvCommand                           `command:"unset-env" alias:"ue" description:"Remove an env variable from an app"`
	UnsetLabel                         v7.UnsetLabelCommand                         `command:"unset-label" description:"Unset a label (key-value pairs) for an API resource"`
	UnsetOrgRole                       v7.UnsetOrgRoleCommand                       `command:"unset-org-role" description:"Remove an org role from a user"`
	UnsetSpaceQuota                    v7.UnsetSpaceQuotaCommand                    `command:"unset-space-quota" description:"Unassign a quota from a space"`
	UnsetSpaceRole                     v7.UnsetSpaceRoleCommand                     `command:"unset-space-role" description:"Remove a space role from a user"`
	UnsharePrivateDomain               v7.UnsharePrivateDomainCommand               `command:"unshare-private-domain" description:"Unshare a private domain with a specific org"`
	UnshareService                     v6.UnshareServiceCommand                     `command:"unshare-service" description:"Unshare a shared service instance from a space"`
	UpdateBuildpack                    v7.UpdateBuildpackCommand                    `command:"update-buildpack" description:"Update a buildpack"`
	UpdateOrgQuota                     v7.UpdateOrgQuotaCommand                     `command:"update-org-quota" description:"Update an existing organization quota"`
	UpdateSecurityGroup                v6.UpdateSecurityGroupCommand                `command:"update-security-group" description:"Update a security group"`
	UpdateService                      v6.UpdateServiceCommand                      `command:"update-service" description:"Update a service instance"`
	UpdateServiceBroker                v7.UpdateServiceBrokerCommand                `command:"update-service-broker" description:"Update a service broker"`
	UpdateSpaceQuota                   v7.UpdateSpaceQuotaCommand                   `command:"update-space-quota" description:"Update an existing space quota"`
	UpdateUserProvidedService          v6.UpdateUserProvidedServiceCommand          `command:"update-user-provided-service" alias:"uups" description:"Update user-provided service instance"`
	Version                            VersionCommand                               `command:"version" description:"Print the version"`
}

// HasCommand returns true if the command name is in the command list.
func (c commandList) HasCommand(name string) bool {
	if name == "" {
		return false
	}

	cType := reflect.TypeOf(c)
	_, found := cType.FieldByNameFunc(
		func(fieldName string) bool {
			field, _ := cType.FieldByName(fieldName)
			return field.Tag.Get("command") == name
		},
	)

	return found
}

// HasAlias returns true if the command alias is in the command list.
func (c commandList) HasAlias(alias string) bool {
	if alias == "" {
		return false
	}

	cType := reflect.TypeOf(c)
	_, found := cType.FieldByNameFunc(
		func(fieldName string) bool {
			field, _ := cType.FieldByName(fieldName)
			return field.Tag.Get("alias") == alias
		},
	)

	return found
}
