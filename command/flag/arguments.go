package flag

type AppName struct {
	AppName string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
}

type Buildpack struct {
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

type Service struct {
	Service string `positional-arg-name:"SERVICE" required:"true" description:"The service offering name"`
}

type ServiceInstance struct {
	ServiceInstance string `positional-arg-name:"SERVICE_INSTANCE" required:"true" description:"The service instance name"`
}

type Organization struct {
	Organization string `positional-arg-name:"ORG" required:"true" description:"The organization"`
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
	ServiceGroup string `positional-arg-name:"SECURITY_GROUP" required:"true" description:"The security group"`
}

type ServiceBroker struct {
	ServiceBroker string `positional-arg-name:"SERVICE_BROKER" required:"true" description:"The service broker"`
}

type Space struct {
	Space string `positional-arg-name:"SPACE" required:"true" description:"The space"`
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
	Username string `positional-arg-name:"USERNAME" required:"true" description:"The username"`
	Password string `positional-arg-name:"PASSWORD" required:"true" description:"The password"`
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

type UnbindSecurityGroupArgs struct {
	SecurityGroupName string `positional-arg-name:"SECURITY_GROUP" required:"true" description:"The security group name"`
	OrganizationName  string `positional-arg-name:"ORG" required:"true" description:"The organization group name"`
	SpaceName         string `positional-arg-name:"SPACE" required:"true" description:"The space name"`
}

type FilesArgs struct {
	AppName string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	Path    string `positional-arg-name:"PATH" description:"The file path"`
}

type SetEnvironmentArgs struct {
	AppName                  string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	EnvironmentVariableName  string `positional-arg-name:"ENV_VAR_NAME" required:"true" description:"The environment variable name"`
	EnvironmentVariableValue string `positional-arg-name:"ENV_VAR_VALUE" required:"true" description:"The environment variable value"`
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
	ServiceOffering string `positional-arg-name:"SERVICE" required:"true" description:"The service offering"`
	ServicePlan     string `positional-arg-name:"SERVICE_PLAN" required:"true" description:"The service plan that the service instance will use"`
	ServiceInstance string `positional-arg-name:"SERVICE_INSTANCE" required:"true" description:"The service instance"`
}

type RenameServiceArgs struct {
	ServiceInstance        string `positional-arg-name:"SERVICE_INSTANCE" required:"true" description:"The service instance to rename"`
	NewServiceInstanceName string `positional-arg-name:"NEW_SERVICE_INSTANCE" required:"true" description:"The new name of the service instance"`
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
	NewOrgName string `positional-arg-name:"NEW_ORG" required:"true" description:"The new organization name"`
}

type RenameSpaceArgs struct {
	OldSpaceName string `positional-arg-name:"SPACE_NAME" required:"true" description:"The old space name"`
	NewSpaceName string `positional-arg-name:"NEW_SPACE_NAME" required:"true" description:"The new space name"`
}

type SetOrgQuotaArgs struct {
	Organization string `positional-arg-name:"ORG" required:"true" description:"The organization"`
	Quota        string `positional-arg-name:"QUOTA" required:"true" description:"The quota"`
}

type SetSpaceQuotaArgs struct {
	Space      string `positional-arg-name:"SPACE_NAME" required:"true" description:"The space"`
	SpaceQuota string `positional-arg-name:"SPACE_QUOTA" required:"true" description:"The space quota"`
}

type SetHealthCheckArgs struct {
	AppName string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	Port    string `positional:"PORT" description:"Set to port"`
	None    string `positional:"NONE" description:"Set to none"`
}

type CreateBuildpackArgs struct {
	Buildpack string `positional-arg-name:"BUILDPACK" required:"true" description:"The buildpack"`
	Path      string `positional-arg-name:"PATH" required:"true" description:"The path to the buildpack file"`
	Position  string `positional-arg-name:"POSITION" required:"true" description:"The position that sets priority"`
}

type RenameBuildpackArgs struct {
	OldBuildpackName string `positional-arg-name:"BUILDPACK_NAME" required:"true" description:"The old buildpack name"`
	NewBuildpackName string `positional-arg-name:"NEW_BUILDPACK_NAME" required:"true" description:"The new buildpack name"`
}

type SetOrgRoleArgs struct {
	Username     string `positional-arg-name:"USERNAME" required:"true" description:"The user"`
	Organization string `positional-arg-name:"ORG" required:"true" description:"The organization"`
	Role         string `positional-arg-name:"ROLE" required:"true" description:"The organization role"`
}

type SetSpaceRoleArgs struct {
	Username     string `positional-arg-name:"USERNAME" required:"true" description:"The user"`
	Organization string `positional-arg-name:"ORG" required:"true" description:"The organization"`
	Space        string `positional-arg-name:"ORG" required:"true" description:"The space"`
	Role         string `positional-arg-name:"ROLE" required:"true" description:"The space role"`
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

type ServiceBrokerArgs struct {
	ServiceBroker string `positional-arg-name:"SERVICE_BROKER" required:"true" description:"The service broker name"`
	Username      string `positional-arg-name:"USERNAME" required:"true" description:"The username"`
	Password      string `positional-arg-name:"PASSWORD" required:"true" description:"The password"`
	URL           string `positional-arg-name:"URL" required:"true" description:"The URL of the service broker"`
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
	SecurityGroup   string `positional-arg-name:"SECURITY_GROUP" required:"true" description:"The security group"`
	PathToJsonRules string `positional-arg-name:"PATH_TO_JSON_RULES_FILE" required:"true" description:"Path to file of JSON describing security group rules"`
}

type AddPluginRepoArgs struct {
	PluginRepoName string `positional-arg-name:"REPO_NAME" required:"true" description:"The plugin repo name"`
	PluginRepoURL  string `positional-arg-name:"URL" required:"true" description:"The URL to the plugin repo"`
}

type InstallPluginArgs struct {
	LocalPath string `positional-arg-name:"LOCAL_PATH/TO/PLUGIN" description:"The local path to the plugin, if the plugin exists locally"`
	URL       string `positional-arg-name:"URL" description:"The URL to the plugin, if the plugin exists online"`
}

type RunTaskArgs struct {
	AppName string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	Command string `positional-arg-name:"COMMAND" required:"true" description:"The command to execute"`
}

type TerminateTaskArgs struct {
	AppName    string `positional-arg-name:"APP_NAME" required:"true" description:"The application name"`
	SequenceID string `positional-arg-name:"TASK_ID" required:"true" description:"The task's unique sequence ID"`
}
