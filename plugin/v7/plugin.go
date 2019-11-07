// +build V7

package v7

import plugin_models "code.cloudfoundry.org/cli/plugin/v7/models"

/**
	Command interface needs to be implemented for a runnable plugin of `cf`
**/
type Plugin interface {
	Run(cliConnection CliConnection, args []string)
	GetMetadata() PluginMetadata
}

/**
	List of commands available to CliConnection variable passed into run
**/
type CliConnection interface {
	GetApp(string) (plugin_models.DetailedApplicationSummary, error)
	AccessToken() (string, error)
}

type VersionType struct {
	Major int
	Minor int
	Build int
}

type PluginMetadata struct {
	Name           string
	Version        VersionType
	LibraryVersion VersionType
	MinCliVersion  VersionType
	Commands       []Command
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
