package plugin

/**
	Command interface needs to be implemented for a runnable plugin of `cf`
**/
type Plugin interface {
	Run(args []string)
	GetMetadata() PluginMetadata
}

type PluginMetadata struct {
	Name     string
	Commands []Command
}

type Command struct {
	Name     string
	HelpText string
}
