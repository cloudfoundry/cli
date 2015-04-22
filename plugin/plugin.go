package plugin

import "github.com/cloudfoundry/cli/plugin/models"

/**
	Command interface needs to be implemented for a runnable plugin of `cf`
**/
type Plugin interface {
	Run(cliConnection CliConnection, args []string)
	GetMetadata() PluginMetadata
}

/**
	List of commands avaiable to CliConnection variable passed into run
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
	IsSSLDisabled() (bool, error)
	HasOrganization() (bool, error)
	HasSpace() (bool, error)
	ApiEndpoint() (string, error)
	ApiVersion() (string, error)
	HasAPIEndpoint() (bool, error)
	LoggregatorEndpoint() (string, error)
	DopplerEndpoint() (string, error)
	AccessToken() (string, error)
}

type VersionType struct {
	Major int
	Minor int
	Build int
}

type PluginMetadata struct {
	Name     string
	Version  VersionType
	Commands []Command
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
