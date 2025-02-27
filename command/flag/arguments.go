package flag

type AppName struct {
	AppName string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
}

type OptionalAppName struct {
	AppName string `positional-arg-name:"APP_NAME" description:"The application name"`
}

type AppDroplet struct {
	AppName     string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	DropletGUID string `positional-arg-name:"DROPLET_GUID" required:"true" description:"The droplet guid"`
}

type BuildpackName struct {
	Buildpack string `positional-arg-name:"BUILDPACK" required:"true" description:"The buildpack"`
}

type CommandName struct {
	CommandName string `positional-arg-name:"COMMAND_NAME" description:"The command name"`
}

type Domain struct {
	Domain string `positional-arg-name:"DOMAIN" required:"true" description:"The domain"`
}

type Feature struct {
	Feature string `positional-arg-name:"FEATURE_NAME" required:"true" description:"The feature flag name"`
}

type ParamsAsJSON struct {
	JSON string `positional-arg-name:"JSON" required:"true" description:"Parameters as JSON"`
}

type ServiceOffering struct {
	ServiceOffering string `positional-arg-name:"SERVICE_OFFERING" required:"true" description:"The service offering name"`
}

type ServiceInstance struct {
	ServiceInstance TrimmedString `positional-arg-name:"SERVICE_INSTANCE" required:"true" description:"The service instance name"`
}

type Organization struct {
	Organization string `positional-arg-name:"ORG" required:"true" description:"The organization"`
}

type OrganizationQuota struct {
	OrganizationQuotaName string `positional-arg-name:"ORG_QUOTA_NAME" required:"true" description:"The organization quota name"`
}

type APIPath struct {
	Path string `positional-arg-name:"PATH" required:"true" description:"The API endpoint"`
}

type PluginRepoName struct {
	PluginRepoName string `positional-arg-name:"REPO_NAME" required:"true" description:"The plugin repo name"`
}

type PluginName struct {
	PluginName string `positional-arg-name:"PLUGIN_NAME" required:"true" description:"The plugin name"`
}

type Quota struct {
	Quota string `positional-arg-name:"QUOTA" required:"true" description:"The organization quota"`
}

type SecurityGroup struct {
	SecurityGroup string `positional-arg-name:"SECURITY_GROUP" required:"true" description:"The security group"`
}

type Space struct {
	Space string `positional-arg-name:"SPACE" required:"true" description:"The space"`
}

type Rename struct {
	OldAppName string `positional-arg-name:"APP_NAME" required:"true" description:"The current app name"`
	NewAppName string `positional-arg-name:"NEW_APP_NAME" required:"true" description:"The new app name"`
}

type RenameSpace struct {
	OldSpaceName string `positional-arg-name:"SPACE" required:"true" description:"The old space name"`
	NewSpaceName string `positional-arg-name:"NEW_SPACE_NAME" required:"true" description:"The new space name"`
}

type SpaceQuota struct {
	SpaceQuota string `positional-arg-name:"SPACE_QUOTA_NAME" required:"true" description:"The space quota"`
}

type StackName struct {
	StackName string `positional-arg-name:"STACK_NAME" required:"true" description:"The stack name"`
}

type Username struct {
	Username string `positional-arg-name:"USERNAME" required:"true" description:"The username"`
}

type APITarget struct {
	URL string `positional-arg-name:"URL" description:"API URL to target"`
}

type Authentication struct {
	Username string `positional-arg-name:"USERNAME" description:"The username"`
	Password string `positional-arg-name:"PASSWORD" description:"The password"`
}

type CreateUser struct {
	Username string  `positional-arg-name:"USERNAME" required:"true" description:"The username"`
	Password *string `positional-arg-name:"PASSWORD" description:"The password"`
}

type AppInstance struct {
	AppName string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	Index   int    `positional-arg-name:"INDEX" required:"true" description:"The index of the application instance"`
}

type OrgSpace struct {
	Organization string `positional-arg-name:"ORG" required:"true" description:"The organization"`
	Space        string `positional-arg-name:"SPACE" required:"true" description:"The space"`
}

type ServiceInstanceKey struct {
	ServiceInstance string `positional-arg-name:"SERVICE_INSTANCE" required:"true" description:"The service instance"`
	ServiceKey      string `positional-arg-name:"SERVICE_KEY" required:"true" description:"The service key"`
}

type AppDomain struct {
	App    string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	Domain string `positional-arg-name:"DOMAIN" required:"true" description:"The domain"`
}

type HostDomain struct {
	Host   string `positional-arg-name:"HOST" required:"true" description:"The hostname"`
	Domain string `positional-arg-name:"DOMAIN" required:"true" description:"The domain"`
}

type OrgDomain struct {
	Organization string `positional-arg-name:"ORG" required:"true" description:"The organization"`
	Domain       string `positional-arg-name:"DOMAIN" required:"true" description:"The domain"`
}

type SpaceDomain struct {
	Space  string `positional-arg-name:"SPACE" required:"true" description:"The space"`
	Domain string `positional-arg-name:"DOMAIN" required:"true" description:"The domain"`
}

type BindSecurityGroupArgs struct {
	SecurityGroupName string `positional-arg-name:"SECURITY_GROUP" required:"true" description:"The security group name"`
	OrganizationName  string `positional-arg-name:"ORG" required:"true" description:"The organization group name"`
	SpaceName         string `positional-arg-name:"SPACE" description:"The space name"`
}

type BindSecurityGroupV7Args struct {
	SecurityGroupName string `positional-arg-name:"SECURITY_GROUP" required:"true" description:"The security group name"`
	OrganizationName  string `positional-arg-name:"ORG" required:"true" description:"The organization group name"`
}

type UnbindSecurityGroupArgs struct {
	SecurityGroupName string `positional-arg-name:"SECURITY_GROUP" required:"true" description:"The security group name"`
	OrganizationName  string `positional-arg-name:"ORG" description:"The organization group name"`
	SpaceName         string `positional-arg-name:"SPACE" description:"The space name"`
}

type UnbindSecurityGroupV7Args struct {
	SecurityGroupName string `positional-arg-name:"SECURITY_GROUP" required:"true" description:"The security group name"`
	OrganizationName  string `positional-arg-name:"ORG" required:"true" description:"The organization group name"`
	SpaceName         string `positional-arg-name:"SPACE" required:"true" description:"The space name"`
}

type FilesArgs struct {
	AppName string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	Path    string `positional-arg-name:"PATH" description:"The file path"`
}

type EnvironmentArgs struct {
	AppName string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
}

type SetEnvironmentArgs struct {
	AppName                  string              `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	EnvironmentVariableName  string              `positional-arg-name:"ENV_VAR_NAME" required:"true" description:"The environment variable name"`
	EnvironmentVariableValue EnvironmentVariable `positional-arg-name:"ENV_VAR_VALUE" required:"true" description:"The environment variable value"`
}

type UnsetEnvironmentArgs struct {
	AppName                 string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	EnvironmentVariableName string `positional-arg-name:"ENV_VAR_NAME" required:"true" description:"The environment variable name"`
}

type CopySourceArgs struct {
	SourceAppName string `positional-arg-name:"SOURCE-APP" required:"true" description:"The old application name"`
	TargetAppName string `positional-arg-name:"TARGET-NAME" required:"true" description:"The new application name"`
}

type CreateServiceArgs struct {
	ServiceOffering string `positional-arg-name:"SERVICE_OFFERING" required:"true" description:"The service offering"`
	ServicePlan     string `positional-arg-name:"SERVICE_PLAN" required:"true" description:"The service plan that the service instance will use"`
	ServiceInstance string `positional-arg-name:"SERVICE_INSTANCE" required:"true" description:"The service instance"`
}

type RenameServiceArgs struct {
	ServiceInstance        string `positional-arg-name:"SERVICE_INSTANCE" required:"true" description:"The service instance to rename"`
	NewServiceInstanceName string `positional-arg-name:"NEW_SERVICE_INSTANCE" required:"true" description:"The new name of the service instance"`
}

type ShareServiceArgs struct {
	ServiceInstance string `positional-arg-name:"SERVICE_INSTANCE" required:"true" description:"The service instance to rename"`
}

type BindServiceArgs struct {
	AppName             string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	ServiceInstanceName string `positional-arg-name:"SERVICE_INSTANCE" required:"true" description:"The service instance"`
}

type RouteServiceArgs struct {
	Domain          string `positional-arg-name:"DOMAIN" required:"true" description:"The domain of the route"`
	ServiceInstance string `positional-arg-name:"SERVICE_INSTANCE" required:"true" description:"The service instance"`
}

type AppRenameArgs struct {
	OldAppName string `positional-arg-name:"APP_NAME" required:"true" description:"The old application name"`
	NewAppName string `positional-arg-name:"NEW_APP_NAME" required:"true" description:"The new application name"`
}

type RenameOrgArgs struct {
	OldOrgName string `positional-arg-name:"ORG" required:"true" description:"The old organization name"`
	NewOrgName string `positional-arg-name:"NEW_ORG_NAME" required:"true" description:"The new organization name"`
}

type RenameSpaceArgs struct {
	OldSpaceName string `positional-arg-name:"SPACE_NAME" required:"true" description:"The old space name"`
	NewSpaceName string `positional-arg-name:"NEW_SPACE_NAME" required:"true" description:"The new space name"`
}

type SetOrgQuotaArgs struct {
	Organization      string `positional-arg-name:"ORG" required:"true" description:"The organization"`
	OrganizationQuota string `positional-arg-name:"QUOTA" required:"true" description:"The quota"`
}

type SetSpaceQuotaArgs struct {
	Space      string `positional-arg-name:"SPACE_NAME" required:"true" description:"The space"`
	SpaceQuota string `positional-arg-name:"QUOTA" required:"true" description:"The space quota"`
}

type UnsetSpaceQuotaArgs struct {
	Space      string `positional-arg-name:"SPACE_NAME" required:"true" description:"The space"`
	SpaceQuota string `positional-arg-name:"SPACE_QUOTA" required:"true" description:"The space quota"`
}

type SetEnvVarGroup struct {
	EnvVarGroupJson string `positional-arg-name:"JSON_STRING" required:"true" description:"json string"`
}

type V6SetHealthCheckArgs struct {
	AppName     string                             `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	HealthCheck HealthCheckTypeWithDeprecatedValue `positional-arg-name:"HEALTH_CHECK_TYPE" required:"true" description:"Set to 'port' or 'none'"`
}

type SetHealthCheckArgs struct {
	AppName     string          `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	HealthCheck HealthCheckType `positional-arg-name:"HEALTH_CHECK_TYPE" required:"true" description:"Set to 'port'"`
}

type CreateBuildpackArgs struct {
	Buildpack string                      `positional-arg-name:"BUILDPACK" required:"true" description:"The buildpack"`
	Path      PathWithExistenceCheckOrURL `positional-arg-name:"PATH" required:"true" description:"The path to the buildpack file"`
	Position  int                         `positional-arg-name:"POSITION" required:"true" description:"The position that sets priority"`
}

type RenameBuildpackArgs struct {
	OldBuildpackName string `positional-arg-name:"BUILDPACK_NAME" required:"true" description:"The old buildpack name"`
	NewBuildpackName string `positional-arg-name:"NEW_BUILDPACK_NAME" required:"true" description:"The new buildpack name"`
}

type LabelsArgs struct {
	ResourceType string `positional-arg-name:"RESOURCE" required:"true" description:"The type of resource to label"`
	ResourceName string `positional-arg-name:"RESOURCE_NAME" required:"true" description:"The name of the resource"`
}

type SetLabelArgs struct {
	ResourceType string   `positional-arg-name:"RESOURCE" required:"true" description:"The type of resource to label"`
	ResourceName string   `positional-arg-name:"RESOURCE_NAME" required:"true" description:"The name of the resource"`
	Labels       []string `positional-arg-name:"KEY=VALUE" required:"true" description:"A space-separated list of labels to set on the resource"`
}

type UnsetLabelArgs struct {
	ResourceType string   `positional-arg-name:"RESOURCE" required:"true" description:"The type of resource"`
	ResourceName string   `positional-arg-name:"RESOURCE_NAME" required:"true" description:"The name of the resource"`
	LabelKeys    []string `positional-arg-name:"KEY" required:"true" description:"A label to unset on the resource"`
}
type OrgRoleArgs struct {
	Username     string  `positional-arg-name:"USERNAME" required:"true" description:"The user"`
	Organization string  `positional-arg-name:"ORG" required:"true" description:"The organization"`
	Role         OrgRole `positional-arg-name:"ROLE" required:"true" description:"The organization role"`
}

type SpaceRoleArgs struct {
	Username     string    `positional-arg-name:"USERNAME" required:"true" description:"The user"`
	Organization string    `positional-arg-name:"ORG" required:"true" description:"The organization"`
	Space        string    `positional-arg-name:"SPACE" required:"true" description:"The space"`
	Role         SpaceRole `positional-arg-name:"ROLE" required:"true" description:"The space role"`
}

type SpaceUsersArgs struct {
	Organization string `positional-arg-name:"ORG" required:"true" description:"The organization"`
	Space        string `positional-arg-name:"SPACE" required:"true" description:"The space"`
}

type ServiceAuthTokenArgs struct {
	Label    string `positional-arg-name:"LABEL" required:"true" description:"The token label"`
	Provider string `positional-arg-name:"PROVIDER" required:"true" description:"The token provider"`
	Token    string `positional-arg-name:"TOKEN" required:"true" description:"The token"`
}

type DeleteServiceAuthTokenArgs struct {
	Label    string `positional-arg-name:"LABEL" required:"true" description:"The token label"`
	Provider string `positional-arg-name:"PROVIDER" required:"true" description:"The token provider"`
}

type ServiceBroker struct {
	ServiceBroker string `positional-arg-name:"SERVICE_BROKER" required:"true" description:"The service broker"`
}

type ServiceBrokerArgs struct {
	ServiceBroker string `positional-arg-name:"SERVICE_BROKER" required:"true" description:"The service broker name"`
	Username      string `positional-arg-name:"USERNAME" required:"true" description:"The username"`
	PasswordOrURL string `positional-arg-name:"URL" required:"true" description:"The URL of the service broker"`
	URL           string `positional-arg-name:"URL" description:"The URL of the service broker"`
}

type RenameServiceBrokerArgs struct {
	OldServiceBrokerName string `positional-arg-name:"SERVICE_BROKER" required:"true" description:"The old service broker name"`
	NewServiceBrokerName string `positional-arg-name:"NEW_SERVICE_BROKER" required:"true" description:"The new service broker name"`
}

type MigrateServiceInstancesArgs struct {
	V1Service  string `positional-arg-name:"v1_SERVICE" required:"true" description:"The old service offering"`
	V1Provider string `positional-arg-name:"v1_PROVIDER" required:"true" description:"The old service provider"`
	V1Plan     string `positional-arg-name:"v1_PLAN" required:"true" description:"The old service plan"`
	V2Service  string `positional-arg-name:"v2_SERVICE" required:"true" description:"The new service offering"`
	V2Plan     string `positional-arg-name:"v2_PLAN" required:"true" description:"The new service plan"`
}

type SecurityGroupArgs struct {
	SecurityGroup   string                 `positional-arg-name:"SECURITY_GROUP" required:"true" description:"The security group"`
	PathToJSONRules PathWithExistenceCheck `positional-arg-name:"PATH_TO_JSON_RULES_FILE" required:"true" description:"Path to file of JSON describing security group rules"`
}

type AddPluginRepoArgs struct {
	PluginRepoName string `positional-arg-name:"REPO_NAME" required:"true" description:"The plugin repo name"`
	PluginRepoURL  string `positional-arg-name:"URL" required:"true" description:"The URL to the plugin repo"`
}

type InstallPluginArgs struct {
	PluginNameOrLocation Path `positional-arg-name:"PLUGIN_NAME_OR_LOCATION" required:"true" description:"The local path to the plugin, if the plugin exists locally; the URL to the plugin, if the plugin exists online; or the plugin name, if a repo is specified"`
}

type RunTaskArgs struct {
	AppName string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	Command string `positional-arg-name:"COMMAND" required:"true" description:"The command to execute"`
}

type RunTaskArgsV7 struct {
	AppName string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
}

type TerminateTaskArgs struct {
	AppName    string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	SequenceID string `positional-arg-name:"TASK_ID" required:"true" description:"The task's unique sequence ID"`
}

type IsolationSegmentName struct {
	IsolationSegmentName string `positional-arg-name:"SEGMENT_NAME" required:"true" description:"The isolation segment name"`
}

type OrgIsolationArgs struct {
	OrganizationName     string `positional-arg-name:"ORG_NAME" required:"true" description:"The organization name"`
	IsolationSegmentName string `positional-arg-name:"SEGMENT_NAME" required:"true" description:"The isolation segment name"`
}

type SpaceIsolationArgs struct {
	SpaceName            string `positional-arg-name:"SPACE_NAME" required:"true" description:"The space name"`
	IsolationSegmentName string `positional-arg-name:"SEGMENT_NAME" required:"true" description:"The isolation segment name"`
}

type ResetSpaceIsolationArgs struct {
	SpaceName string `positional-arg-name:"SPACE_NAME" required:"true" description:"The space name"`
}

type ResetOrgDefaultIsolationArgs struct {
	OrgName string `positional-arg-name:"ORG_NAME" required:"true" description:"The organization name"`
}

type AddNetworkPolicyArgs struct {
	SourceApp string `positional-arg-name:"SOURCE_APP" required:"true" description:"The source app"`
}

type AddNetworkPolicyArgsV7 struct {
	SourceApp string `positional-arg-name:"SOURCE_APP" required:"true" description:"The source app"`
	DestApp   string `positional-arg-name:"DESTINATION_APP" required:"true" description:"The destination app"`
}

type RemoveNetworkPolicyArgs struct {
	SourceApp string
}

type RemoveNetworkPolicyArgsV7 struct {
	SourceApp string `positional-arg-name:"SOURCE_APP" required:"true" description:"The source app"`
	DestApp   string `positional-arg-name:"DESTINATION_APP" required:"true" description:"The destination app"`
}

type TaskArgs struct {
	AppName string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	TaskID  int    `positional-arg-name:"TASK_ID" required:"true" description:"The Task ID for the application"`
}
