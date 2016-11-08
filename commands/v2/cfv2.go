package v2

import "code.cloudfoundry.org/cli/commands/v3"

var Commands commandList

const ExperimentalWarning = "This command is in EXPERIMENTAL stage and may change without notice"

type commandList struct {
	VerboseOrVersion                   bool                                      `short:"v" long:"version" description:"verbose and version flag"`
	App                                AppCommand                                `command:"app" description:"Display health and status for app"`
	Help                               HelpCommand                               `command:"help" alias:"h" description:"Show help"`
	Version                            VersionCommand                            `command:"version" description:"Print the version"`
	Login                              LoginCommand                              `command:"login" alias:"l" description:"Log user in"`
	Logout                             LogoutCommand                             `command:"logout" alias:"lo" description:"Log user out"`
	Passwd                             PasswdCommand                             `command:"passwd" alias:"pw" description:"Change user password"`
	Target                             TargetCommand                             `command:"target" alias:"t" description:"Set or view the targeted org or space"`
	Api                                ApiCommand                                `command:"api" description:"Set or view target api url"`
	Auth                               AuthCommand                               `command:"auth" description:"Authenticate user non-interactively"`
	Apps                               AppsCommand                               `command:"apps" alias:"a" description:"List all apps in the target space"`
	Push                               PushCommand                               `command:"push" alias:"p" description:"Push a new app or sync changes to an existing app"`
	Scale                              ScaleCommand                              `command:"scale" description:"Change or view the instance count, disk space limit, and memory limit for an app"`
	Delete                             DeleteCommand                             `command:"delete" alias:"d" description:"Delete an app"`
	Rename                             RenameCommand                             `command:"rename" description:"Rename an app"`
	Start                              StartCommand                              `command:"start" alias:"st" description:"Start an app"`
	Stop                               StopCommand                               `command:"stop" alias:"sp" description:"Stop an app"`
	Restart                            RestartCommand                            `command:"restart" alias:"rs" description:"Stop all instances of the app, then start them again. This may cause downtime."`
	Restage                            RestageCommand                            `command:"restage" alias:"rg" description:"Recreate the app's executable artifact using the latest pushed app files and the latest environment (variables, service bindings, buildpack, stack, etc.)"`
	RestartAppInstance                 RestartAppInstanceCommand                 `command:"restart-app-instance" description:"Terminate the running application Instance at the given index and instantiate a new instance of the application with the same index"`
	Events                             EventsCommand                             `command:"events" description:"Show recent app events"`
	Files                              FilesCommand                              `command:"files" alias:"f" description:"Print out a list of files in a directory or the contents of a specific file of an app running on the DEA backend"`
	Logs                               LogsCommand                               `command:"logs" description:"Tail or show recent logs for an app"`
	Env                                EnvCommand                                `command:"env" alias:"e" description:"Show all env variables for an app"`
	SetEnv                             SetEnvCommand                             `command:"set-env" alias:"se" description:"Set an env variable for an app"`
	UnsetEnv                           UnsetEnvCommand                           `command:"unset-env" description:"Remove an env variable"`
	Stacks                             StacksCommand                             `command:"stacks" description:"List all stacks (a stack is a pre-built file system, including an operating system, that can run apps)"`
	Stack                              StackCommand                              `command:"stack" description:"Show information for a stack (a stack is a pre-built file system, including an operating system, that can run apps)"`
	CopySource                         CopySourceCommand                         `command:"copy-source" description:"Copies the source code of an application to another existing application (and restarts that application)"`
	CreateAppManifest                  CreateAppManifestCommand                  `command:"create-app-manifest" description:"Create an app manifest for an app that has been pushed successfully"`
	GetHealthCheck                     GetHealthCheckCommand                     `command:"get-health-check" description:"Get the health_check_type value of an app"`
	SetHealthCheck                     SetHealthCheckCommand                     `command:"set-health-check" description:"Set health_check_type flag to either 'port' or 'none'"`
	EnableSSH                          EnableSSHCommand                          `command:"enable-ssh" description:"Enable ssh for the application"`
	DisableSSH                         DisableSSHCommand                         `command:"disable-ssh" description:"Disable ssh for the application"`
	SSHEnabled                         SSHEnabledCommand                         `command:"ssh-enabled" description:"Reports whether SSH is enabled on an application container instance"`
	SSH                                SSHCommand                                `command:"ssh" description:"SSH to an application container instance"`
	Marketplace                        MarketplaceCommand                        `command:"marketplace" alias:"m" description:"List available offerings in the marketplace"`
	Services                           ServicesCommand                           `command:"services" alias:"s" description:"List all service instances in the target space"`
	Service                            ServiceCommand                            `command:"service" description:"Show service instance info"`
	CreateService                      CreateServiceCommand                      `command:"create-service" alias:"cs" description:"Create a service instance"`
	UpdateService                      UpdateServiceCommand                      `command:"update-service" description:"Update a service instance"`
	DeleteService                      DeleteServiceCommand                      `command:"delete-service" alias:"ds" description:"Delete a service instance"`
	RenameService                      RenameServiceCommand                      `command:"rename-service" description:"Rename a service instance"`
	CreateServiceKey                   CreateServiceKeyCommand                   `command:"create-service-key" alias:"csk" description:"Create key for a service instance"`
	ServiceKeys                        ServiceKeysCommand                        `command:"service-keys" alias:"sk" description:"List keys for a service instance"`
	ServiceKey                         ServiceKeyCommand                         `command:"service-key" description:"Show service key info"`
	DeleteServiceKey                   DeleteServiceKeyCommand                   `command:"delete-service-key" alias:"dsk" description:"Delete a service key"`
	BindService                        BindServiceCommand                        `command:"bind-service" alias:"bs" description:"Bind a service instance to an app"`
	UnbindService                      UnbindServiceCommand                      `command:"unbind-service" alias:"us" description:"Unbind a service instance from an app"`
	BindRouteService                   BindRouteServiceCommand                   `command:"bind-route-service" alias:"brs" description:"Bind a service instance to an HTTP route"`
	UnbindRouteService                 UnbindRouteServiceCommand                 `command:"unbind-route-service" alias:"urs" description:"Unbind a service instance from an HTTP route"`
	CreateUserProvidedService          CreateUserProvidedServiceCommand          `command:"create-user-provided-service" alias:"cups" description:"Make a user-provided service instance available to CF apps"`
	UpdateUserProvidedService          UpdateUserProvidedServiceCommand          `command:"update-user-provided-service" alias:"uups" description:"Update user-provided service instance"`
	Orgs                               OrgsCommand                               `command:"orgs" alias:"o" description:"List all orgs"`
	Org                                OrgCommand                                `command:"org" description:"Show org info"`
	CreateOrg                          CreateOrgCommand                          `command:"create-org" alias:"co" description:"Create an org"`
	DeleteOrg                          DeleteOrgCommand                          `command:"delete-org" description:"Delete an org"`
	RenameOrg                          RenameOrgCommand                          `command:"rename-org" description:"Rename an org"`
	Spaces                             SpacesCommand                             `command:"spaces" description:"List all spaces in an org"`
	Space                              SpaceCommand                              `command:"space" description:"Show space info"`
	CreateSpace                        CreateSpaceCommand                        `command:"create-space" description:"Create a space"`
	DeleteSpace                        DeleteSpaceCommand                        `command:"delete-space" description:"Delete a space"`
	RenameSpace                        RenameSpaceCommand                        `command:"rename-space" description:"Rename a space"`
	AllowSpaceSSH                      AllowSpaceSSHCommand                      `command:"allow-space-ssh" description:"Allow SSH access for the space"`
	DisallowSpaceSSH                   DisallowSpaceSSHCommand                   `command:"disallow-space-ssh" description:"Disallow SSH access for the space"`
	SpaceSSHAllowed                    SpaceSSHAllowedCommand                    `command:"space-ssh-allowed" description:"Reports whether SSH is allowed in a space"`
	Domains                            DomainsCommand                            `command:"domains" description:"List domains in the target org"`
	CreateDomain                       CreateDomainCommand                       `command:"create-domain" description:"Create a domain in an org for later use"`
	DeleteDomain                       DeleteDomainCommand                       `command:"delete-domain" description:"Delete a domain"`
	CreateSharedDomain                 CreateSharedDomainCommand                 `command:"create-shared-domain" description:"Create a domain that can be used by all orgs (admin-only)"`
	DeleteSharedDomain                 DeleteSharedDomainCommand                 `command:"delete-shared-domain" description:"Delete a shared domain"`
	RouterGroups                       RouterGroupsCommand                       `command:"router-groups" description:"List router groups"`
	Routes                             RoutesCommand                             `command:"routes" alias:"r" description:"List all routes in the current space or the current organization"`
	CreateRoute                        CreateRouteCommand                        `command:"create-route" description:"Create a url route in a space for later use"`
	CheckRoute                         CheckRouteCommand                         `command:"check-route" description:"Perform a simple check to determine whether a route currently exists or not"`
	MapRoute                           MapRouteCommand                           `command:"map-route" description:"Add a url route to an app"`
	UnmapRoute                         UnmapRouteCommand                         `command:"unmap-route" description:"Remove a url route from an app"`
	DeleteRoute                        DeleteRouteCommand                        `command:"delete-route" description:"Delete a route"`
	DeleteOrphanedRoutes               DeleteOrphanedRoutesCommand               `command:"delete-orphaned-routes" description:"Delete all orphaned routes (i.e. those that are not mapped to an app)"`
	Buildpacks                         BuildpacksCommand                         `command:"buildpacks" description:"List all buildpacks"`
	CreateBuildpack                    CreateBuildpackCommand                    `command:"create-buildpack" description:"Create a buildpack"`
	UpdateBuildpack                    UpdateBuildpackCommand                    `command:"update-buildpack" description:"Update a buildpack"`
	RenameBuildpack                    RenameBuildpackCommand                    `command:"rename-buildpack" description:"Rename a buildpack"`
	DeleteBuildpack                    DeleteBuildpackCommand                    `command:"delete-buildpack" description:"Delete a buildpack"`
	CreateUser                         CreateUserCommand                         `command:"create-user" description:"Create a new user"`
	DeleteUser                         DeleteUserCommand                         `command:"delete-user" description:"Delete a user"`
	OrgUsers                           OrgUsersCommand                           `command:"org-users" description:"Show org users by role"`
	SetOrgRole                         SetOrgRoleCommand                         `command:"set-org-role" description:"Assign an org role to a user"`
	UnsetOrgRole                       UnsetOrgRoleCommand                       `command:"unset-org-role" description:"Remove an org role from a user"`
	SpaceUsers                         SpaceUsersCommand                         `command:"space-users" description:"Show space users by role"`
	SetSpaceRole                       SetSpaceRoleCommand                       `command:"set-space-role" description:"Assign a space role to a user"`
	UnsetSpaceRole                     UnsetSpaceRoleCommand                     `command:"unset-space-role" description:"Remove a space role from a user"`
	Quotas                             QuotasCommand                             `command:"quotas" description:"List available usage quotas"`
	Quota                              QuotaCommand                              `command:"quota" description:"Show quota info"`
	SetQuota                           SetQuotaCommand                           `command:"set-quota" description:"Assign a quota to an org"`
	CreateQuota                        CreateQuotaCommand                        `command:"create-quota" description:"Define a new resource quota"`
	DeleteQuota                        DeleteQuotaCommand                        `command:"delete-quota" description:"Delete a quota"`
	UpdateQuota                        UpdateQuotaCommand                        `command:"update-quota" description:"Update an existing resource quota"`
	SharePrivateDomain                 SharePrivateDomainCommand                 `command:"share-private-domain" description:"Share a private domain with an org"`
	UnsharePrivateDomain               UnsharePrivateDomainCommand               `command:"unshare-private-domain" description:"Unshare a private domain with an org"`
	SpaceQuotas                        SpaceQuotasCommand                        `command:"space-quotas" description:"List available space resource quotas"`
	SpaceQuota                         SpaceQuotaCommand                         `command:"space-quota" description:"Show space quota info"`
	CreateSpaceQuota                   CreateSpaceQuotaCommand                   `command:"create-space-quota" description:"Define a new space resource quota"`
	UpdateSpaceQuota                   UpdateSpaceQuotaCommand                   `command:"update-space-quota" description:"Update an existing space quota"`
	DeleteSpaceQuota                   DeleteSpaceQuotaCommand                   `command:"delete-space-quota" description:"Delete a space quota definition and unassign the space quota from all spaces"`
	SetSpaceQuota                      SetSpaceQuotaCommand                      `command:"set-space-quota" description:"Assign a space quota definition to a space"`
	UnsetSpaceQuota                    UnsetSpaceQuotaCommand                    `command:"unset-space-quota" description:"Unassign a quota from a space"`
	ServiceAuthTokens                  ServiceAuthTokensCommand                  `command:"service-auth-tokens" description:"List service auth tokens"`
	CreateServiceAuthToken             CreateServiceAuthTokenCommand             `command:"create-service-auth-token" description:"Create a service auth token"`
	UpdateServiceAuthToken             UpdateServiceAuthTokenCommand             `command:"update-service-auth-token" description:"Update a service auth token"`
	DeleteServiceAuthToken             DeleteServiceAuthTokenCommand             `command:"delete-service-auth-token" description:"Delete a service auth token"`
	ServiceBrokers                     ServiceBrokersCommand                     `command:"service-brokers" description:"List service brokers"`
	CreateServiceBroker                CreateServiceBrokerCommand                `command:"create-service-broker" alias:"csb" description:"Create a service broker"`
	UpdateServiceBroker                UpdateServiceBrokerCommand                `command:"update-service-broker" description:"Update a service broker"`
	DeleteServiceBroker                DeleteServiceBrokerCommand                `command:"delete-service-broker" description:"Delete a service broker"`
	RenameServiceBroker                RenameServiceBrokerCommand                `command:"rename-service-broker" description:"Rename a service broker"`
	MigrateServiceInstances            MigrateServiceInstancesCommand            `command:"migrate-service-instances" description:"Migrate service instances from one service plan to another"`
	PurgeServiceOffering               PurgeServiceOfferingCommand               `command:"purge-service-offering" description:"Recursively remove a service and child objects from Cloud Foundry database without making requests to a service broker"`
	PurgeServiceInstance               PurgeServiceInstanceCommand               `command:"purge-service-instance" description:"Recursively remove a service instance and child objects from Cloud Foundry database without making requests to a service broker"`
	ServiceAccess                      ServiceAccessCommand                      `command:"service-access" description:"List service access settings"`
	EnableServiceAccess                EnableServiceAccessCommand                `command:"enable-service-access" description:"Enable access to a service or service plan for one or all orgs"`
	DisableServiceAccess               DisableServiceAccessCommand               `command:"disable-service-access" description:"Disable access to a service or service plan for one or all orgs"`
	SecurityGroup                      SecurityGroupCommand                      `command:"security-group" description:"Show a single security group"`
	SecurityGroups                     SecurityGroupsCommand                     `command:"security-groups" description:"List all security groups"`
	CreateSecurityGroup                CreateSecurityGroupCommand                `command:"create-security-group" description:"Create a security group"`
	UpdateSecurityGroup                UpdateSecurityGroupCommand                `command:"update-security-group" description:"Update a security group"`
	DeleteSecurityGroup                DeleteSecurityGroupCommand                `command:"delete-security-group" description:"Deletes a security group"`
	BindSecurityGroup                  BindSecurityGroupCommand                  `command:"bind-security-group" description:"Bind a security group to a particular space, or all existing spaces of an org"`
	UnbindSecurityGroup                UnbindSecurityGroupCommand                `command:"unbind-security-group" description:"Unbind a security group from a space"`
	BindStagingSecurityGroup           BindStagingSecurityGroupCommand           `command:"bind-staging-security-group" description:"Bind a security group to the list of security groups to be used for staging applications"`
	StagingSecurityGroups              StagingSecurityGroupsCommand              `command:"staging-security-groups" description:"List security groups in the staging set for applications"`
	UnbindStagingSecurityGroup         UnbindStagingSecurityGroupCommand         `command:"unbind-staging-security-group" description:"Unbind a security group from the set of security groups for staging applications"`
	BindRunningSecurityGroup           BindRunningSecurityGroupCommand           `command:"bind-running-security-group" description:"Bind a security group to the list of security groups to be used for running applications"`
	RunningSecurityGroups              RunningSecurityGroupsCommand              `command:"running-security-groups" description:"List security groups in the set of security groups for running applications"`
	UnbindRunningSecurityGroup         UnbindRunningSecurityGroupCommand         `command:"unbind-running-security-group" description:"Unbind a security group from the set of security groups for running applications"`
	RunningEnvironmentVariableGroup    RunningEnvironmentVariableGroupCommand    `command:"running-environment-variable-group" alias:"revg" description:"Retrieve the contents of the running environment variable group"`
	StagingEnvironmentVariableGroup    StagingEnvironmentVariableGroupCommand    `command:"staging-environment-variable-group" alias:"sevg" description:"Retrieve the contents of the staging environment variable group"`
	SetStagingEnvironmentVariableGroup SetStagingEnvironmentVariableGroupCommand `command:"set-staging-environment-variable-group" alias:"ssevg" description:"Pass parameters as JSON to create a staging environment variable group"`
	SetRunningEnvironmentVariableGroup SetRunningEnvironmentVariableGroupCommand `command:"set-running-environment-variable-group" alias:"srevg" description:"Pass parameters as JSON to create a running environment variable group"`
	FeatureFlags                       FeatureFlagsCommand                       `command:"feature-flags" description:"Retrieve list of feature flags with status of each flag-able feature"`
	FeatureFlag                        FeatureFlagCommand                        `command:"feature-flag" description:"Retrieve an individual feature flag with status"`
	EnableFeatureFlag                  EnableFeatureFlagCommand                  `command:"enable-feature-flag" description:"Enable the use of a feature so that users have access to and can use the feature"`
	DisableFeatureFlag                 DisableFeatureFlagCommand                 `command:"disable-feature-flag" description:"Disable the use of a feature so that users have access to and can use the feature"`
	Curl                               CurlCommand                               `command:"curl" description:"Executes a request to the targeted API endpoint"`
	Config                             ConfigCommand                             `command:"config" description:"Write default values to the config"`
	OauthToken                         OauthTokenCommand                         `command:"oauth-token" description:"Retrieve and display the OAuth token for the current session"`
	SSHCode                            SSHCodeCommand                            `command:"ssh-code" description:"Get a one time password for ssh clients"`
	AddPluginRepo                      AddPluginRepoCommand                      `command:"add-plugin-repo" description:"Add a new plugin repository"`
	RemovePluginRepo                   RemovePluginRepoCommand                   `command:"remove-plugin-repo" description:"Remove a plugin repository"`
	ListPluginRepos                    ListPluginReposCommand                    `command:"list-plugin-repos" description:"List all the added plugin repositories"`
	RepoPlugins                        RepoPluginsCommand                        `command:"repo-plugins" description:"List all available plugins in specified repository or in all added repositories"`
	Plugins                            PluginsCommand                            `command:"plugins" description:"List all available plugin commands"`
	InstallPlugin                      InstallPluginCommand                      `command:"install-plugin" description:"Install CLI plugin"`
	UninstallPlugin                    UninstallPluginCommand                    `command:"uninstall-plugin" description:"Uninstall the plugin defined in command argument"`
	RunTask                            v3.RunTaskCommand                         `command:"run-task" alias:"rt" description:"Run a one-off task on an app"`
	Tasks                              v3.TasksCommand                           `command:"tasks" description:"List tasks of an app"`
}
