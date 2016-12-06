package common

import (
	"code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v3"
)

var Commands commandList

type commandList struct {
	VerboseOrVersion                   bool                                         `short:"v" long:"version" description:"verbose and version flag"`
	App                                v2.AppCommand                                `command:"app" description:"Display health and status for app"`
	Help                               HelpCommand                                  `command:"help" alias:"h" description:"Show help"`
	Version                            VersionCommand                               `command:"version" description:"Print the version"`
	Login                              v2.LoginCommand                              `command:"login" alias:"l" description:"Log user in"`
	Logout                             v2.LogoutCommand                             `command:"logout" alias:"lo" description:"Log user out"`
	Passwd                             v2.PasswdCommand                             `command:"passwd" alias:"pw" description:"Change user password"`
	Target                             v2.TargetCommand                             `command:"target" alias:"t" description:"Set or view the targeted org or space"`
	Api                                v2.ApiCommand                                `command:"api" description:"Set or view target api url"`
	Auth                               v2.AuthCommand                               `command:"auth" description:"Authenticate user non-interactively"`
	Apps                               v2.AppsCommand                               `command:"apps" alias:"a" description:"List all apps in the target space"`
	Push                               v2.PushCommand                               `command:"push" alias:"p" description:"Push a new app or sync changes to an existing app"`
	Scale                              v2.ScaleCommand                              `command:"scale" description:"Change or view the instance count, disk space limit, and memory limit for an app"`
	Delete                             v2.DeleteCommand                             `command:"delete" alias:"d" description:"Delete an app"`
	Rename                             v2.RenameCommand                             `command:"rename" description:"Rename an app"`
	Start                              v2.StartCommand                              `command:"start" alias:"st" description:"Start an app"`
	Stop                               v2.StopCommand                               `command:"stop" alias:"sp" description:"Stop an app"`
	Restart                            v2.RestartCommand                            `command:"restart" alias:"rs" description:"Stop all instances of the app, then start them again. This may cause downtime."`
	Restage                            v2.RestageCommand                            `command:"restage" alias:"rg" description:"Recreate the app's executable artifact using the latest pushed app files and the latest environment (variables, service bindings, buildpack, stack, etc.)"`
	RestartAppInstance                 v2.RestartAppInstanceCommand                 `command:"restart-app-instance" description:"Terminate the running application Instance at the given index and instantiate a new instance of the application with the same index"`
	Events                             v2.EventsCommand                             `command:"events" description:"Show recent app events"`
	Files                              v2.FilesCommand                              `command:"files" alias:"f" description:"Print out a list of files in a directory or the contents of a specific file of an app running on the DEA backend"`
	Logs                               v2.LogsCommand                               `command:"logs" description:"Tail or show recent logs for an app"`
	Env                                v2.EnvCommand                                `command:"env" alias:"e" description:"Show all env variables for an app"`
	SetEnv                             v2.SetEnvCommand                             `command:"set-env" alias:"se" description:"Set an env variable for an app"`
	UnsetEnv                           v2.UnsetEnvCommand                           `command:"unset-env" description:"Remove an env variable"`
	Stacks                             v2.StacksCommand                             `command:"stacks" description:"List all stacks (a stack is a pre-built file system, including an operating system, that can run apps)"`
	Stack                              v2.StackCommand                              `command:"stack" description:"Show information for a stack (a stack is a pre-built file system, including an operating system, that can run apps)"`
	CopySource                         v2.CopySourceCommand                         `command:"copy-source" description:"Copies the source code of an application to another existing application (and restarts that application)"`
	CreateAppManifest                  v2.CreateAppManifestCommand                  `command:"create-app-manifest" description:"Create an app manifest for an app that has been pushed successfully"`
	GetHealthCheck                     v2.GetHealthCheckCommand                     `command:"get-health-check" description:"Get the health_check_type value of an app"`
	SetHealthCheck                     v2.SetHealthCheckCommand                     `command:"set-health-check" description:"Set health_check_type flag to either 'port' or 'none'"`
	EnableSSH                          v2.EnableSSHCommand                          `command:"enable-ssh" description:"Enable ssh for the application"`
	DisableSSH                         v2.DisableSSHCommand                         `command:"disable-ssh" description:"Disable ssh for the application"`
	SSHEnabled                         v2.SSHEnabledCommand                         `command:"ssh-enabled" description:"Reports whether SSH is enabled on an application container instance"`
	SSH                                v2.SSHCommand                                `command:"ssh" description:"SSH to an application container instance"`
	Marketplace                        v2.MarketplaceCommand                        `command:"marketplace" alias:"m" description:"List available offerings in the marketplace"`
	Services                           v2.ServicesCommand                           `command:"services" alias:"s" description:"List all service instances in the target space"`
	Service                            v2.ServiceCommand                            `command:"service" description:"Show service instance info"`
	CreateService                      v2.CreateServiceCommand                      `command:"create-service" alias:"cs" description:"Create a service instance"`
	UpdateService                      v2.UpdateServiceCommand                      `command:"update-service" description:"Update a service instance"`
	DeleteService                      v2.DeleteServiceCommand                      `command:"delete-service" alias:"ds" description:"Delete a service instance"`
	RenameService                      v2.RenameServiceCommand                      `command:"rename-service" description:"Rename a service instance"`
	CreateServiceKey                   v2.CreateServiceKeyCommand                   `command:"create-service-key" alias:"csk" description:"Create key for a service instance"`
	ServiceKeys                        v2.ServiceKeysCommand                        `command:"service-keys" alias:"sk" description:"List keys for a service instance"`
	ServiceKey                         v2.ServiceKeyCommand                         `command:"service-key" description:"Show service key info"`
	DeleteServiceKey                   v2.DeleteServiceKeyCommand                   `command:"delete-service-key" alias:"dsk" description:"Delete a service key"`
	BindService                        v2.BindServiceCommand                        `command:"bind-service" alias:"bs" description:"Bind a service instance to an app"`
	UnbindService                      v2.UnbindServiceCommand                      `command:"unbind-service" alias:"us" description:"Unbind a service instance from an app"`
	BindRouteService                   v2.BindRouteServiceCommand                   `command:"bind-route-service" alias:"brs" description:"Bind a service instance to an HTTP route"`
	UnbindRouteService                 v2.UnbindRouteServiceCommand                 `command:"unbind-route-service" alias:"urs" description:"Unbind a service instance from an HTTP route"`
	CreateUserProvidedService          v2.CreateUserProvidedServiceCommand          `command:"create-user-provided-service" alias:"cups" description:"Make a user-provided service instance available to CF apps"`
	UpdateUserProvidedService          v2.UpdateUserProvidedServiceCommand          `command:"update-user-provided-service" alias:"uups" description:"Update user-provided service instance"`
	Orgs                               v2.OrgsCommand                               `command:"orgs" alias:"o" description:"List all orgs"`
	Org                                v2.OrgCommand                                `command:"org" description:"Show org info"`
	CreateOrg                          v2.CreateOrgCommand                          `command:"create-org" alias:"co" description:"Create an org"`
	DeleteOrg                          v2.DeleteOrgCommand                          `command:"delete-org" description:"Delete an org"`
	RenameOrg                          v2.RenameOrgCommand                          `command:"rename-org" description:"Rename an org"`
	Spaces                             v2.SpacesCommand                             `command:"spaces" description:"List all spaces in an org"`
	Space                              v2.SpaceCommand                              `command:"space" description:"Show space info"`
	CreateSpace                        v2.CreateSpaceCommand                        `command:"create-space" description:"Create a space"`
	DeleteSpace                        v2.DeleteSpaceCommand                        `command:"delete-space" description:"Delete a space"`
	RenameSpace                        v2.RenameSpaceCommand                        `command:"rename-space" description:"Rename a space"`
	AllowSpaceSSH                      v2.AllowSpaceSSHCommand                      `command:"allow-space-ssh" description:"Allow SSH access for the space"`
	DisallowSpaceSSH                   v2.DisallowSpaceSSHCommand                   `command:"disallow-space-ssh" description:"Disallow SSH access for the space"`
	SpaceSSHAllowed                    v2.SpaceSSHAllowedCommand                    `command:"space-ssh-allowed" description:"Reports whether SSH is allowed in a space"`
	Domains                            v2.DomainsCommand                            `command:"domains" description:"List domains in the target org"`
	CreateDomain                       v2.CreateDomainCommand                       `command:"create-domain" description:"Create a domain in an org for later use"`
	DeleteDomain                       v2.DeleteDomainCommand                       `command:"delete-domain" description:"Delete a domain"`
	CreateSharedDomain                 v2.CreateSharedDomainCommand                 `command:"create-shared-domain" description:"Create a domain that can be used by all orgs (admin-only)"`
	DeleteSharedDomain                 v2.DeleteSharedDomainCommand                 `command:"delete-shared-domain" description:"Delete a shared domain"`
	RouterGroups                       v2.RouterGroupsCommand                       `command:"router-groups" description:"List router groups"`
	Routes                             v2.RoutesCommand                             `command:"routes" alias:"r" description:"List all routes in the current space or the current organization"`
	CreateRoute                        v2.CreateRouteCommand                        `command:"create-route" description:"Create a url route in a space for later use"`
	CheckRoute                         v2.CheckRouteCommand                         `command:"check-route" description:"Perform a simple check to determine whether a route currently exists or not"`
	MapRoute                           v2.MapRouteCommand                           `command:"map-route" description:"Add a url route to an app"`
	UnmapRoute                         v2.UnmapRouteCommand                         `command:"unmap-route" description:"Remove a url route from an app"`
	DeleteRoute                        v2.DeleteRouteCommand                        `command:"delete-route" description:"Delete a route"`
	DeleteOrphanedRoutes               v2.DeleteOrphanedRoutesCommand               `command:"delete-orphaned-routes" description:"Delete all orphaned routes (i.e. those that are not mapped to an app)"`
	Buildpacks                         v2.BuildpacksCommand                         `command:"buildpacks" description:"List all buildpacks"`
	CreateBuildpack                    v2.CreateBuildpackCommand                    `command:"create-buildpack" description:"Create a buildpack"`
	UpdateBuildpack                    v2.UpdateBuildpackCommand                    `command:"update-buildpack" description:"Update a buildpack"`
	RenameBuildpack                    v2.RenameBuildpackCommand                    `command:"rename-buildpack" description:"Rename a buildpack"`
	DeleteBuildpack                    v2.DeleteBuildpackCommand                    `command:"delete-buildpack" description:"Delete a buildpack"`
	CreateUser                         v2.CreateUserCommand                         `command:"create-user" description:"Create a new user"`
	DeleteUser                         v2.DeleteUserCommand                         `command:"delete-user" description:"Delete a user"`
	OrgUsers                           v2.OrgUsersCommand                           `command:"org-users" description:"Show org users by role"`
	SetOrgRole                         v2.SetOrgRoleCommand                         `command:"set-org-role" description:"Assign an org role to a user"`
	UnsetOrgRole                       v2.UnsetOrgRoleCommand                       `command:"unset-org-role" description:"Remove an org role from a user"`
	SpaceUsers                         v2.SpaceUsersCommand                         `command:"space-users" description:"Show space users by role"`
	SetSpaceRole                       v2.SetSpaceRoleCommand                       `command:"set-space-role" description:"Assign a space role to a user"`
	UnsetSpaceRole                     v2.UnsetSpaceRoleCommand                     `command:"unset-space-role" description:"Remove a space role from a user"`
	Quotas                             v2.QuotasCommand                             `command:"quotas" description:"List available usage quotas"`
	Quota                              v2.QuotaCommand                              `command:"quota" description:"Show quota info"`
	SetQuota                           v2.SetQuotaCommand                           `command:"set-quota" description:"Assign a quota to an org"`
	CreateQuota                        v2.CreateQuotaCommand                        `command:"create-quota" description:"Define a new resource quota"`
	DeleteQuota                        v2.DeleteQuotaCommand                        `command:"delete-quota" description:"Delete a quota"`
	UpdateQuota                        v2.UpdateQuotaCommand                        `command:"update-quota" description:"Update an existing resource quota"`
	SharePrivateDomain                 v2.SharePrivateDomainCommand                 `command:"share-private-domain" description:"Share a private domain with an org"`
	UnsharePrivateDomain               v2.UnsharePrivateDomainCommand               `command:"unshare-private-domain" description:"Unshare a private domain with an org"`
	SpaceQuotas                        v2.SpaceQuotasCommand                        `command:"space-quotas" description:"List available space resource quotas"`
	SpaceQuota                         v2.SpaceQuotaCommand                         `command:"space-quota" description:"Show space quota info"`
	CreateSpaceQuota                   v2.CreateSpaceQuotaCommand                   `command:"create-space-quota" description:"Define a new space resource quota"`
	UpdateSpaceQuota                   v2.UpdateSpaceQuotaCommand                   `command:"update-space-quota" description:"Update an existing space quota"`
	DeleteSpaceQuota                   v2.DeleteSpaceQuotaCommand                   `command:"delete-space-quota" description:"Delete a space quota definition and unassign the space quota from all spaces"`
	SetSpaceQuota                      v2.SetSpaceQuotaCommand                      `command:"set-space-quota" description:"Assign a space quota definition to a space"`
	UnsetSpaceQuota                    v2.UnsetSpaceQuotaCommand                    `command:"unset-space-quota" description:"Unassign a quota from a space"`
	ServiceAuthTokens                  v2.ServiceAuthTokensCommand                  `command:"service-auth-tokens" description:"List service auth tokens"`
	CreateServiceAuthToken             v2.CreateServiceAuthTokenCommand             `command:"create-service-auth-token" description:"Create a service auth token"`
	UpdateServiceAuthToken             v2.UpdateServiceAuthTokenCommand             `command:"update-service-auth-token" description:"Update a service auth token"`
	DeleteServiceAuthToken             v2.DeleteServiceAuthTokenCommand             `command:"delete-service-auth-token" description:"Delete a service auth token"`
	ServiceBrokers                     v2.ServiceBrokersCommand                     `command:"service-brokers" description:"List service brokers"`
	CreateServiceBroker                v2.CreateServiceBrokerCommand                `command:"create-service-broker" alias:"csb" description:"Create a service broker"`
	UpdateServiceBroker                v2.UpdateServiceBrokerCommand                `command:"update-service-broker" description:"Update a service broker"`
	DeleteServiceBroker                v2.DeleteServiceBrokerCommand                `command:"delete-service-broker" description:"Delete a service broker"`
	RenameServiceBroker                v2.RenameServiceBrokerCommand                `command:"rename-service-broker" description:"Rename a service broker"`
	MigrateServiceInstances            v2.MigrateServiceInstancesCommand            `command:"migrate-service-instances" description:"Migrate service instances from one service plan to another"`
	PurgeServiceOffering               v2.PurgeServiceOfferingCommand               `command:"purge-service-offering" description:"Recursively remove a service and child objects from Cloud Foundry database without making requests to a service broker"`
	PurgeServiceInstance               v2.PurgeServiceInstanceCommand               `command:"purge-service-instance" description:"Recursively remove a service instance and child objects from Cloud Foundry database without making requests to a service broker"`
	ServiceAccess                      v2.ServiceAccessCommand                      `command:"service-access" description:"List service access settings"`
	EnableServiceAccess                v2.EnableServiceAccessCommand                `command:"enable-service-access" description:"Enable access to a service or service plan for one or all orgs"`
	DisableServiceAccess               v2.DisableServiceAccessCommand               `command:"disable-service-access" description:"Disable access to a service or service plan for one or all orgs"`
	SecurityGroup                      v2.SecurityGroupCommand                      `command:"security-group" description:"Show a single security group"`
	SecurityGroups                     v2.SecurityGroupsCommand                     `command:"security-groups" description:"List all security groups"`
	CreateSecurityGroup                v2.CreateSecurityGroupCommand                `command:"create-security-group" description:"Create a security group"`
	UpdateSecurityGroup                v2.UpdateSecurityGroupCommand                `command:"update-security-group" description:"Update a security group"`
	DeleteSecurityGroup                v2.DeleteSecurityGroupCommand                `command:"delete-security-group" description:"Deletes a security group"`
	BindSecurityGroup                  v2.BindSecurityGroupCommand                  `command:"bind-security-group" description:"Bind a security group to a particular space, or all existing spaces of an org"`
	UnbindSecurityGroup                v2.UnbindSecurityGroupCommand                `command:"unbind-security-group" description:"Unbind a security group from a space"`
	BindStagingSecurityGroup           v2.BindStagingSecurityGroupCommand           `command:"bind-staging-security-group" description:"Bind a security group to the list of security groups to be used for staging applications"`
	StagingSecurityGroups              v2.StagingSecurityGroupsCommand              `command:"staging-security-groups" description:"List security groups in the staging set for applications"`
	UnbindStagingSecurityGroup         v2.UnbindStagingSecurityGroupCommand         `command:"unbind-staging-security-group" description:"Unbind a security group from the set of security groups for staging applications"`
	BindRunningSecurityGroup           v2.BindRunningSecurityGroupCommand           `command:"bind-running-security-group" description:"Bind a security group to the list of security groups to be used for running applications"`
	RunningSecurityGroups              v2.RunningSecurityGroupsCommand              `command:"running-security-groups" description:"List security groups in the set of security groups for running applications"`
	UnbindRunningSecurityGroup         v2.UnbindRunningSecurityGroupCommand         `command:"unbind-running-security-group" description:"Unbind a security group from the set of security groups for running applications"`
	RunningEnvironmentVariableGroup    v2.RunningEnvironmentVariableGroupCommand    `command:"running-environment-variable-group" alias:"revg" description:"Retrieve the contents of the running environment variable group"`
	StagingEnvironmentVariableGroup    v2.StagingEnvironmentVariableGroupCommand    `command:"staging-environment-variable-group" alias:"sevg" description:"Retrieve the contents of the staging environment variable group"`
	SetStagingEnvironmentVariableGroup v2.SetStagingEnvironmentVariableGroupCommand `command:"set-staging-environment-variable-group" alias:"ssevg" description:"Pass parameters as JSON to create a staging environment variable group"`
	SetRunningEnvironmentVariableGroup v2.SetRunningEnvironmentVariableGroupCommand `command:"set-running-environment-variable-group" alias:"srevg" description:"Pass parameters as JSON to create a running environment variable group"`
	FeatureFlags                       v2.FeatureFlagsCommand                       `command:"feature-flags" description:"Retrieve list of feature flags with status of each flag-able feature"`
	FeatureFlag                        v2.FeatureFlagCommand                        `command:"feature-flag" description:"Retrieve an individual feature flag with status"`
	EnableFeatureFlag                  v2.EnableFeatureFlagCommand                  `command:"enable-feature-flag" description:"Enable the use of a feature so that users have access to and can use the feature"`
	DisableFeatureFlag                 v2.DisableFeatureFlagCommand                 `command:"disable-feature-flag" description:"Disable the use of a feature so that users have access to and can use the feature"`
	Curl                               v2.CurlCommand                               `command:"curl" description:"Executes a request to the targeted API endpoint"`
	Config                             v2.ConfigCommand                             `command:"config" description:"Write default values to the config"`
	OauthToken                         v2.OauthTokenCommand                         `command:"oauth-token" description:"Retrieve and display the OAuth token for the current session"`
	SSHCode                            v2.SSHCodeCommand                            `command:"ssh-code" description:"Get a one time password for ssh clients"`
	AddPluginRepo                      v2.AddPluginRepoCommand                      `command:"add-plugin-repo" description:"Add a new plugin repository"`
	RemovePluginRepo                   v2.RemovePluginRepoCommand                   `command:"remove-plugin-repo" description:"Remove a plugin repository"`
	ListPluginRepos                    v2.ListPluginReposCommand                    `command:"list-plugin-repos" description:"List all the added plugin repositories"`
	RepoPlugins                        v2.RepoPluginsCommand                        `command:"repo-plugins" description:"List all available plugins in specified repository or in all added repositories"`
	Plugins                            v2.PluginsCommand                            `command:"plugins" description:"List all available plugin commands"`
	InstallPlugin                      v2.InstallPluginCommand                      `command:"install-plugin" description:"Install CLI plugin"`
	UninstallPlugin                    v2.UninstallPluginCommand                    `command:"uninstall-plugin" description:"Uninstall the plugin defined in command argument"`
	RunTask                            v3.RunTaskCommand                            `command:"run-task" alias:"rt" description:"Run a one-off task on an app"`
	Tasks                              v3.TasksCommand                              `command:"tasks" description:"List tasks of an app"`
	TerminateTask                      v3.TerminateTaskCommand                      `command:"terminate-task" description:"Terminate a running task of an app"`
}
