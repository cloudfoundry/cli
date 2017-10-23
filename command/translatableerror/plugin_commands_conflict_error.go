package translatableerror

import "strings"

// PluginCommandsConflictError is returned when a plugin command name conflicts
// with a native or existing plugin command name.
type PluginCommandsConflictError struct {
	PluginName     string
	PluginVersion  string
	CommandNames   []string
	CommandAliases []string
}

func (e PluginCommandsConflictError) Error() string {
	switch {
	case len(e.CommandNames) > 0 && len(e.CommandAliases) > 0:
		return "Plugin {{.PluginName}} v{{.PluginVersion}} could not be installed as it contains commands with names and aliases that are already used: {{.CommandNamesAndAliases}}."
	case len(e.CommandNames) > 0:
		return "Plugin {{.PluginName}} v{{.PluginVersion}} could not be installed as it contains commands with names that are already used: {{.CommandNames}}."
	case len(e.CommandAliases) > 0:
		return "Plugin {{.PluginName}} v{{.PluginVersion}} could not be installed as it contains commands with aliases that are already used: {{.CommandAliases}}."
	default:
		return "Plugin {{.PluginName}} v{{.PluginVersion}} could not be installed as it contains commands with names or aliases that are already used."
	}
}

func (e PluginCommandsConflictError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"PluginName":             e.PluginName,
		"PluginVersion":          e.PluginVersion,
		"CommandNames":           strings.Join(e.CommandNames, ", "),
		"CommandAliases":         strings.Join(e.CommandAliases, ", "),
		"CommandNamesAndAliases": strings.Join(append(e.CommandNames, e.CommandAliases...), ", "),
	})
}
