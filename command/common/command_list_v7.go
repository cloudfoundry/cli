// +build V7

package common

import (
	"reflect"

	"code.cloudfoundry.org/cli/command/plugin"
	"code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v7"
)

var Commands commandList
var FallbackCommands V2CommandList

type V2CommandList struct {
	V2App v6.AppCommand `command:"app" description:"Display health and status for an app"`
}

type commandList struct {
	VerboseOrVersion bool `short:"v" long:"version" description:"verbose and version flag"`

	App                  v7.AppCommand                   `command:"app" description:"Display health and status for an app"`
	V3ApplyManifest      v6.V3ApplyManifestCommand       `command:"v3-apply-manifest" description:"Applies manifest properties to an application"`
	V3Apps               v6.V3AppsCommand                `command:"v3-apps" description:"List all apps in the target space"`
	V3CancelZdtPush      v6.V3CancelZdtPushCommand       `command:"v3-cancel-zdt-push" description:"Cancel the most recent deployment for an app"`
	V3CreateApp          v6.V3CreateAppCommand           `command:"v3-create-app" description:"Create a V3 App"`
	V3CreatePackage      v6.V3CreatePackageCommand       `command:"v3-create-package" description:"Uploads a V3 Package"`
	V3Droplets           v6.V3DropletsCommand            `command:"v3-droplets" description:"List droplets of an app"`
	V3Packages           v6.V3PackagesCommand            `command:"v3-packages" description:"List packages of an app"`
	V3Restart            v6.V3RestartCommand             `command:"v3-restart" description:"Stop all instances of the app, then start them again. This causes downtime."`
	V3RestartAppInstance v6.V3RestartAppInstanceCommand  `command:"v3-restart-app-instance" description:"Terminate, then instantiate an app instance"`
	V3SetDroplet         v6.V3SetDropletCommand          `command:"v3-set-droplet" description:"Set the droplet used to run an app"`
	V3Stage              v6.V3StageCommand               `command:"v3-stage" description:"Create a new droplet for an app"`
	V3Start              v6.V3StartCommand               `command:"v3-start" description:"Start an app"`
	V3Stop               v6.V3StopCommand                `command:"v3-stop" description:"Stop an app"`
	V3ZdtRestart         v6.V3ZeroDowntimeRestartCommand `command:"v3-zdt-restart" description:"Sequentially restart each instance of an app."`

	AddPluginRepo                      plugin.AddPluginRepoCommand                  `command:"add-plugin-repo" description:"Add a new plugin repository"`
	AddNetworkPolicy                   v6.AddNetworkPolicyCommand                   `command:"add-network-policy" description:"Create policy to allow direct network traffic from one app to another"`
	AllowSpaceSSH                      v6.AllowSpaceSSHCommand                      `command:"allow-space-ssh" description:"Allow SSH access for the space"`
	Api                                v6.ApiCommand                                `command:"api" description:"Set or view target api url"`
	Apps                               v6.AppsCommand                               `command:"apps" alias:"a" description:"List all apps in the target space"`
	Auth                               v6.AuthCommand                               `command:"auth" description:"Authenticate non-interactively"`
	BindRouteService                   v6.BindRouteServiceCommand                   `command:"bind-route-service" alias:"brs" description:"Bind a service instance to an HTTP route"`
	BindRunningSecurityGroup           v6.BindRunningSecurityGroupCommand           `command:"bind-running-security-group" description:"Bind a security group to the list of security groups to be used for running applications"`
	BindSecurityGroup                  v6.BindSecurityGroupCommand                  `command:"bind-security-group" description:"Bind a security group to a particular space, or all existing spaces of an org"`
	BindService                        v6.BindServiceCommand                        `command:"bind-service" alias:"bs" description:"Bind a service instance to an app"`
	BindStagingSecurityGroup           v6.BindStagingSecurityGroupCommand           `command:"bind-staging-security-group" description:"Bind a security group to the list of security groups to be used for staging applications"`
	Buildpacks                         v6.BuildpacksCommand                         `command:"buildpacks" description:"List all buildpacks"`
	CheckRoute                         v6.CheckRouteCommand                         `command:"check-route" description:"Perform a simple check to determine whether a route currently exists or not"`
	Config                             v6.ConfigCommand                             `command:"config" description:"Write default values to the config"`
	CopySource                         v6.CopySourceCommand                         `command:"copy-source" description:"Copies the source code of an application to another existing application (and restarts that application)"`
	CreateAppManifest                  v6.CreateAppManifestCommand                  `command:"create-app-manifest" description:"Create an app manifest for an app that has been pushed successfully"`
	CreateBuildpack                    v6.CreateBuildpackCommand                    `command:"create-buildpack" description:"Create a buildpack"`
	CreateDomain                       v6.CreateDomainCommand                       `command:"create-domain" description:"Create a domain in an org for later use"`
	CreateIsolationSegment             v6.CreateIsolationSegmentCommand             `command:"create-isolation-segment" description:"Create an isolation segment"`
	CreateOrg                          v6.CreateOrgCommand                          `command:"create-org" alias:"co" description:"Create an org"`
	CreateQuota                        v6.CreateQuotaCommand                        `command:"create-quota" description:"Define a new resource quota"`
	CreateRoute                        v6.CreateRouteCommand                        `command:"create-route" description:"Create a url route in a space for later use"`
	CreateSecurityGroup                v6.CreateSecurityGroupCommand                `command:"create-security-group" description:"Create a security group"`
	CreateServiceBroker                v6.CreateServiceBrokerCommand                `command:"create-service-broker" alias:"csb" description:"Create a service broker"`
	CreateServiceKey                   v6.CreateServiceKeyCommand                   `command:"create-service-key" alias:"csk" description:"Create key for a service instance"`
	CreateService                      v6.CreateServiceCommand                      `command:"create-service" alias:"cs" description:"Create a service instance"`
	CreateSharedDomain                 v6.CreateSharedDomainCommand                 `command:"create-shared-domain" description:"Create a domain that can be used by all orgs (admin-only)"`
	CreateSpaceQuota                   v6.CreateSpaceQuotaCommand                   `command:"create-space-quota" description:"Define a new space resource quota"`
	CreateSpace                        v6.CreateSpaceCommand                        `command:"create-space" description:"Create a space"`
	CreateUserProvidedService          v6.CreateUserProvidedServiceCommand          `command:"create-user-provided-service" alias:"cups" description:"Make a user-provided service instance available to CF apps"`
	CreateUser                         v6.CreateUserCommand                         `command:"create-user" description:"Create a new user"`
	Curl                               v6.CurlCommand                               `command:"curl" description:"Executes a request to the targeted API endpoint"`
	DeleteBuildpack                    v6.DeleteBuildpackCommand                    `command:"delete-buildpack" description:"Delete a buildpack"`
	DeleteDomain                       v6.DeleteDomainCommand                       `command:"delete-domain" description:"Delete a domain"`
	DeleteIsolationSegment             v6.DeleteIsolationSegmentCommand             `command:"delete-isolation-segment" description:"Delete an isolation segment"`
	DeleteOrg                          v6.DeleteOrgCommand                          `command:"delete-org" description:"Delete an org"`
	DeleteOrphanedRoutes               v6.DeleteOrphanedRoutesCommand               `command:"delete-orphaned-routes" description:"Delete all orphaned routes (i.e. those that are not mapped to an app)"`
	DeleteQuota                        v6.DeleteQuotaCommand                        `command:"delete-quota" description:"Delete a quota"`
	DeleteRoute                        v6.DeleteRouteCommand                        `command:"delete-route" description:"Delete a route"`
	DeleteSecurityGroup                v6.DeleteSecurityGroupCommand                `command:"delete-security-group" description:"Deletes a security group"`
	DeleteServiceBroker                v6.DeleteServiceBrokerCommand                `command:"delete-service-broker" description:"Delete a service broker"`
	DeleteServiceKey                   v6.DeleteServiceKeyCommand                   `command:"delete-service-key" alias:"dsk" description:"Delete a service key"`
	DeleteService                      v6.DeleteServiceCommand                      `command:"delete-service" alias:"ds" description:"Delete a service instance"`
	DeleteSharedDomain                 v6.DeleteSharedDomainCommand                 `command:"delete-shared-domain" description:"Delete a shared domain"`
	DeleteSpaceQuota                   v6.DeleteSpaceQuotaCommand                   `command:"delete-space-quota" description:"Delete a space quota definition and unassign the space quota from all spaces"`
	DeleteSpace                        v6.DeleteSpaceCommand                        `command:"delete-space" description:"Delete a space"`
	DeleteUser                         v6.DeleteUserCommand                         `command:"delete-user" description:"Delete a user"`
	Delete                             v7.DeleteCommand                             `command:"delete" alias:"d" description:"Delete an app"`
	DisableFeatureFlag                 v6.DisableFeatureFlagCommand                 `command:"disable-feature-flag" description:"Prevent use of a feature"`
	DisableOrgIsolation                v6.DisableOrgIsolationCommand                `command:"disable-org-isolation" description:"Revoke an organization's entitlement to an isolation segment"`
	DisableServiceAccess               v6.DisableServiceAccessCommand               `command:"disable-service-access" description:"Disable access to a service or service plan for one or all orgs"`
	DisableSSH                         v6.DisableSSHCommand                         `command:"disable-ssh" description:"Disable ssh for the application"`
	DisallowSpaceSSH                   v6.DisallowSpaceSSHCommand                   `command:"disallow-space-ssh" description:"Disallow SSH access for the space"`
	Domains                            v6.DomainsCommand                            `command:"domains" description:"List domains in the target org"`
	EnableFeatureFlag                  v6.EnableFeatureFlagCommand                  `command:"enable-feature-flag" description:"Allow use of a feature"`
	EnableOrgIsolation                 v6.EnableOrgIsolationCommand                 `command:"enable-org-isolation" description:"Entitle an organization to an isolation segment"`
	EnableServiceAccess                v6.EnableServiceAccessCommand                `command:"enable-service-access" description:"Enable access to a service or service plan for one or all orgs"`
	EnableSSH                          v6.EnableSSHCommand                          `command:"enable-ssh" description:"Enable ssh for the application"`
	Env                                v7.EnvCommand                                `command:"env" alias:"e" description:"Show all env variables for an app"`
	Events                             v6.EventsCommand                             `command:"events" description:"Show recent app events"`
	FeatureFlags                       v6.FeatureFlagsCommand                       `command:"feature-flags" description:"Retrieve list of feature flags with status"`
	FeatureFlag                        v6.FeatureFlagCommand                        `command:"feature-flag" description:"Retrieve an individual feature flag with status"`
	GetHealthCheck                     v7.GetHealthCheckCommand                     `command:"get-health-check" description:"Show the type of health check performed on an app"`
	Help                               HelpCommand                                  `command:"help" alias:"h" description:"Show help"`
	InstallPlugin                      InstallPluginCommand                         `command:"install-plugin" description:"Install CLI plugin"`
	IsolationSegments                  v6.IsolationSegmentsCommand                  `command:"isolation-segments" description:"List all isolation segments"`
	NetworkPolicies                    v6.NetworkPoliciesCommand                    `command:"network-policies" description:"List direct network traffic policies"`
	ListPluginRepos                    plugin.ListPluginReposCommand                `command:"list-plugin-repos" description:"List all the added plugin repositories"`
	Login                              v6.LoginCommand                              `command:"login" alias:"l" description:"Log user in"`
	Logout                             v6.LogoutCommand                             `command:"logout" alias:"lo" description:"Log user out"`
	Logs                               v6.LogsCommand                               `command:"logs" description:"Tail or show recent logs for an app"`
	MapRoute                           v6.MapRouteCommand                           `command:"map-route" description:"Add a url route to an app"`
	Marketplace                        v6.MarketplaceCommand                        `command:"marketplace" alias:"m" description:"List available offerings in the marketplace"`
	OauthToken                         v6.OauthTokenCommand                         `command:"oauth-token" description:"Retrieve and display the OAuth token for the current session"`
	Orgs                               v6.OrgsCommand                               `command:"orgs" alias:"o" description:"List all orgs"`
	OrgUsers                           v6.OrgUsersCommand                           `command:"org-users" description:"Show org users by role"`
	Org                                v6.OrgCommand                                `command:"org" description:"Show org info"`
	Passwd                             v6.PasswdCommand                             `command:"passwd" alias:"pw" description:"Change user password"`
	Plugins                            plugin.PluginsCommand                        `command:"plugins" description:"List commands of installed plugins"`
	PurgeServiceInstance               v6.PurgeServiceInstanceCommand               `command:"purge-service-instance" description:"Recursively remove a service instance and child objects from Cloud Foundry database without making requests to a service broker"`
	PurgeServiceOffering               v6.PurgeServiceOfferingCommand               `command:"purge-service-offering" description:"Recursively remove a service and child objects from Cloud Foundry database without making requests to a service broker"`
	Push                               v7.PushCommand                               `command:"push" alias:"p" description:"Push a new app or sync changes to an existing app"`
	Quotas                             v6.QuotasCommand                             `command:"quotas" description:"List available usage quotas"`
	Quota                              v6.QuotaCommand                              `command:"quota" description:"Show quota info"`
	RemoveNetworkPolicy                v6.RemoveNetworkPolicyCommand                `command:"remove-network-policy" description:"Remove network traffic policy of an app"`
	RemovePluginRepo                   plugin.RemovePluginRepoCommand               `command:"remove-plugin-repo" description:"Remove a plugin repository"`
	RenameBuildpack                    v6.RenameBuildpackCommand                    `command:"rename-buildpack" description:"Rename a buildpack"`
	RenameOrg                          v6.RenameOrgCommand                          `command:"rename-org" description:"Rename an org"`
	RenameServiceBroker                v6.RenameServiceBrokerCommand                `command:"rename-service-broker" description:"Rename a service broker"`
	RenameService                      v6.RenameServiceCommand                      `command:"rename-service" description:"Rename a service instance"`
	RenameSpace                        v6.RenameSpaceCommand                        `command:"rename-space" description:"Rename a space"`
	Rename                             v6.RenameCommand                             `command:"rename" description:"Rename an app"`
	RepoPlugins                        plugin.RepoPluginsCommand                    `command:"repo-plugins" description:"List all available plugins in specified repository or in all added repositories"`
	ResetOrgDefaultIsolationSegment    v6.ResetOrgDefaultIsolationSegmentCommand    `command:"reset-org-default-isolation-segment" description:"Reset the default isolation segment used for apps in spaces of an org"`
	ResetSpaceIsolationSegment         v6.ResetSpaceIsolationSegmentCommand         `command:"reset-space-isolation-segment" description:"Reset the space's isolation segment to the org default"`
	Restage                            v6.RestageCommand                            `command:"restage" alias:"rg" description:"Recreate the app's executable artifact using the latest pushed app files and the latest environment (variables, service bindings, buildpack, stack, etc.)"`
	RestartAppInstance                 v6.RestartAppInstanceCommand                 `command:"restart-app-instance" description:"Terminate, then restart an app instance"`
	Restart                            v6.RestartCommand                            `command:"restart" alias:"rs" description:"Stop all instances of the app, then start them again. This causes downtime."`
	RouterGroups                       v6.RouterGroupsCommand                       `command:"router-groups" description:"List router groups"`
	Routes                             v6.RoutesCommand                             `command:"routes" alias:"r" description:"List all routes in the current space or the current organization"`
	RunningEnvironmentVariableGroup    v6.RunningEnvironmentVariableGroupCommand    `command:"running-environment-variable-group" alias:"revg" description:"Retrieve the contents of the running environment variable group"`
	RunningSecurityGroups              v6.RunningSecurityGroupsCommand              `command:"running-security-groups" description:"List security groups in the set of security groups for running applications"`
	RunTask                            v6.RunTaskCommand                            `command:"run-task" alias:"rt" description:"Run a one-off task on an app"`
	Scale                              v7.ScaleCommand                              `command:"scale" description:"Change or view the instance count, disk space limit, and memory limit for an app"`
	SecurityGroups                     v6.SecurityGroupsCommand                     `command:"security-groups" description:"List all security groups"`
	SecurityGroup                      v6.SecurityGroupCommand                      `command:"security-group" description:"Show a single security group"`
	ServiceAccess                      v6.ServiceAccessCommand                      `command:"service-access" description:"List service access settings"`
	ServiceBrokers                     v6.ServiceBrokersCommand                     `command:"service-brokers" description:"List service brokers"`
	ServiceKeys                        v6.ServiceKeysCommand                        `command:"service-keys" alias:"sk" description:"List keys for a service instance"`
	ServiceKey                         v6.ServiceKeyCommand                         `command:"service-key" description:"Show service key info"`
	Services                           v6.ServicesCommand                           `command:"services" alias:"s" description:"List all service instances in the target space"`
	Service                            v6.ServiceCommand                            `command:"service" description:"Show service instance info"`
	SetEnv                             v7.SetEnvCommand                             `command:"set-env" alias:"se" description:"Set an env variable for an app"`
	SetHealthCheck                     v7.SetHealthCheckCommand                     `command:"set-health-check" description:"Change type of health check performed on an app's process"`
	SetOrgDefaultIsolationSegment      v6.SetOrgDefaultIsolationSegmentCommand      `command:"set-org-default-isolation-segment" description:"Set the default isolation segment used for apps in spaces in an org"`
	SetOrgRole                         v6.SetOrgRoleCommand                         `command:"set-org-role" description:"Assign an org role to a user"`
	SetQuota                           v6.SetQuotaCommand                           `command:"set-quota" description:"Assign a quota to an org"`
	SetRunningEnvironmentVariableGroup v6.SetRunningEnvironmentVariableGroupCommand `command:"set-running-environment-variable-group" alias:"srevg" description:"Pass parameters as JSON to create a running environment variable group"`
	SetSpaceIsolationSegment           v6.SetSpaceIsolationSegmentCommand           `command:"set-space-isolation-segment" description:"Assign the isolation segment for a space"`
	SetSpaceQuota                      v6.SetSpaceQuotaCommand                      `command:"set-space-quota" description:"Assign a space quota definition to a space"`
	SetSpaceRole                       v6.SetSpaceRoleCommand                       `command:"set-space-role" description:"Assign a space role to a user"`
	SetStagingEnvironmentVariableGroup v6.SetStagingEnvironmentVariableGroupCommand `command:"set-staging-environment-variable-group" alias:"ssevg" description:"Pass parameters as JSON to create a staging environment variable group"`
	SharePrivateDomain                 v6.SharePrivateDomainCommand                 `command:"share-private-domain" description:"Share a private domain with an org"`
	ShareService                       v6.ShareServiceCommand                       `command:"share-service" description:"Share a service instance with another space"`
	SpaceQuotas                        v6.SpaceQuotasCommand                        `command:"space-quotas" description:"List available space resource quotas"`
	SpaceQuota                         v6.SpaceQuotaCommand                         `command:"space-quota" description:"Show space quota info"`
	SpaceSSHAllowed                    v6.SpaceSSHAllowedCommand                    `command:"space-ssh-allowed" description:"Reports whether SSH is allowed in a space"`
	Spaces                             v6.SpacesCommand                             `command:"spaces" description:"List all spaces in an org"`
	SpaceUsers                         v6.SpaceUsersCommand                         `command:"space-users" description:"Show space users by role"`
	Space                              v6.SpaceCommand                              `command:"space" description:"Show space info"`
	SSHCode                            v6.SSHCodeCommand                            `command:"ssh-code" description:"Get a one time password for ssh clients"`
	SSHEnabled                         v6.SSHEnabledCommand                         `command:"ssh-enabled" description:"Reports whether SSH is enabled on an application container instance"`
	SSH                                v7.SSHCommand                                `command:"ssh" description:"SSH to an application container instance"`
	Stacks                             v6.StacksCommand                             `command:"stacks" description:"List all stacks (a stack is a pre-built file system, including an operating system, that can run apps)"`
	Stack                              v6.StackCommand                              `command:"stack" description:"Show information for a stack (a stack is a pre-built file system, including an operating system, that can run apps)"`
	StagingEnvironmentVariableGroup    v6.StagingEnvironmentVariableGroupCommand    `command:"staging-environment-variable-group" alias:"sevg" description:"Retrieve the contents of the staging environment variable group"`
	StagingSecurityGroups              v6.StagingSecurityGroupsCommand              `command:"staging-security-groups" description:"List security groups in the staging set for applications"`
	Start                              v6.StartCommand                              `command:"start" alias:"st" description:"Start an app"`
	Stop                               v6.StopCommand                               `command:"stop" alias:"sp" description:"Stop an app"`
	Target                             v6.TargetCommand                             `command:"target" alias:"t" description:"Set or view the targeted org or space"`
	Tasks                              v6.TasksCommand                              `command:"tasks" description:"List tasks of an app"`
	TerminateTask                      v6.TerminateTaskCommand                      `command:"terminate-task" description:"Terminate a running task of an app"`
	UnbindRouteService                 v6.UnbindRouteServiceCommand                 `command:"unbind-route-service" alias:"urs" description:"Unbind a service instance from an HTTP route"`
	UnbindRunningSecurityGroup         v6.UnbindRunningSecurityGroupCommand         `command:"unbind-running-security-group" description:"Unbind a security group from the set of security groups for running applications"`
	UnbindSecurityGroup                v6.UnbindSecurityGroupCommand                `command:"unbind-security-group" description:"Unbind a security group from a space"`
	UnbindService                      v6.UnbindServiceCommand                      `command:"unbind-service" alias:"us" description:"Unbind a service instance from an app"`
	UnbindStagingSecurityGroup         v6.UnbindStagingSecurityGroupCommand         `command:"unbind-staging-security-group" description:"Unbind a security group from the set of security groups for staging applications"`
	UninstallPlugin                    plugin.UninstallPluginCommand                `command:"uninstall-plugin" description:"Uninstall CLI plugin"`
	UnmapRoute                         v6.UnmapRouteCommand                         `command:"unmap-route" description:"Remove a url route from an app"`
	UnsetEnv                           v7.UnsetEnvCommand                           `command:"unset-env" description:"Remove an env variable from an app"`
	UnsetOrgRole                       v6.UnsetOrgRoleCommand                       `command:"unset-org-role" description:"Remove an org role from a user"`
	UnsetSpaceQuota                    v6.UnsetSpaceQuotaCommand                    `command:"unset-space-quota" description:"Unassign a quota from a space"`
	UnsetSpaceRole                     v6.UnsetSpaceRoleCommand                     `command:"unset-space-role" description:"Remove a space role from a user"`
	UnsharePrivateDomain               v6.UnsharePrivateDomainCommand               `command:"unshare-private-domain" description:"Unshare a private domain with an org"`
	UnshareService                     v6.UnshareServiceCommand                     `command:"unshare-service" description:"Unshare a shared service instance from a space"`
	UpdateBuildpack                    v6.UpdateBuildpackCommand                    `command:"update-buildpack" description:"Update a buildpack"`
	UpdateQuota                        v6.UpdateQuotaCommand                        `command:"update-quota" description:"Update an existing resource quota"`
	UpdateSecurityGroup                v6.UpdateSecurityGroupCommand                `command:"update-security-group" description:"Update a security group"`
	UpdateServiceBroker                v6.UpdateServiceBrokerCommand                `command:"update-service-broker" description:"Update a service broker"`
	UpdateService                      v6.UpdateServiceCommand                      `command:"update-service" description:"Update a service instance"`
	UpdateSpaceQuota                   v6.UpdateSpaceQuotaCommand                   `command:"update-space-quota" description:"Update an existing space quota"`
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
