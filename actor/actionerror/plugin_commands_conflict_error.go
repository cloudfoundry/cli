package actionerror

// PluginCommandsConflictError is returned when a plugin command name conflicts
// with a core or existing plugin command name.
type PluginCommandsConflictError struct {
	PluginName     string
	PluginVersion  string
	CommandNames   []string
	CommandAliases []string
}

func (PluginCommandsConflictError) Error() string {
	return ""
}
