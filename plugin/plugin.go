package plugin

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
}

type PluginMetadata struct {
	Name     string
	Commands []Command
}

type Command struct {
	Name     string
	HelpText string
}
