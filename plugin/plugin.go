package plugin

import "code.cloudfoundry.org/cli/plugin/models"

/**
	Command interface needs to be implemented for a runnable plugin of `cf`
**/
type Plugin interface {
	Run(cliConnection CliConnection, args []string)
	GetMetadata() PluginMetadata
}

//go:generate counterfeiter . CliConnection
/**
	List of commands available to CliConnection variable passed into run
**/
type CliConnection interface {
	CliCommandWithoutTerminalOutput(args ...string) ([]string, error)
	CliCommand(args ...string) ([]string, error)
	GetCurrentOrg() (plugin_models.Organization, error)
	GetCurrentSpace() (plugin_models.Space, error)
	Username() (string, error)
	UserGuid() (string, error)
	UserEmail() (string, error)
	IsLoggedIn() (bool, error)
	// IsSSLDisabled returns true if and only if the user is connected to the Cloud Controller API with the
	// `--skip-ssl-validation` flag set unless the CLI configuration file cannot be read, in which case it
	// returns an error.
	IsSSLDisabled() (bool, error)
	HasOrganization() (bool, error)
	HasSpace() (bool, error)
	ApiEndpoint() (string, error)
	ApiVersion() (string, error)
	HasAPIEndpoint() (bool, error)
	LoggregatorEndpoint() (string, error)
	DopplerEndpoint() (string, error)
	AccessToken() (string, error)
	GetApp(string) (plugin_models.GetAppModel, error)
	GetApps() ([]plugin_models.GetAppsModel, error)
	GetOrgs() ([]plugin_models.GetOrgs_Model, error)
	GetSpaces() ([]plugin_models.GetSpaces_Model, error)
	GetOrgUsers(string, ...string) ([]plugin_models.GetOrgUsers_Model, error)
	GetSpaceUsers(string, string) ([]plugin_models.GetSpaceUsers_Model, error)
	GetServices() ([]plugin_models.GetServices_Model, error)
	GetService(string) (plugin_models.GetService_Model, error)
	GetOrg(string) (plugin_models.GetOrg_Model, error)
	GetSpace(string) (plugin_models.GetSpace_Model, error)
}

type VersionType struct {
	Major int
	Minor int
	Build int
}

type PluginMetadata struct {
	Name          string
	Version       VersionType
	MinCliVersion VersionType
	Commands      []Command
}

type Usage struct {
	Usage   string
	Options map[string]string
}

type Command struct {
	Name         string
	Alias        string
	HelpText     string
	UsageDetails Usage //Detail usage to be displayed in `cf help <cmd>`
}
