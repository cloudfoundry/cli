package shared

import "strings"

type JSONSyntaxError struct {
	Err error
}

func (e JSONSyntaxError) Error() string {
	return "Invalid JSON content from server: {{.Err}}"
}

func (e JSONSyntaxError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Err": e.Err.Error(),
	})
}

// PluginInstallationCancelled is used to ignore the scenario when the user
// responds with 'no' when prompted to install plugin and exit 0.
type PluginInstallationCancelled struct {
}

func (_ PluginInstallationCancelled) Error() string {
	return "Plugin installation cancelled"
}

type PluginNotFoundInRepositoryError struct {
	BinaryName     string
	PluginName     string
	RepositoryName string
}

func (e PluginNotFoundInRepositoryError) Error() string {
	return "Plugin {{.PluginName}} not found in repository {{.RepositoryName}}.\nUse '{{.BinaryName}} repo-plugins -r {{.RepositoryName}}' to list plugins available in the repo."
}

func (e PluginNotFoundInRepositoryError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"PluginName":     e.PluginName,
		"RepositoryName": e.RepositoryName,
		"BinaryName":     e.BinaryName,
	})
}

type PluginNotFoundError struct {
	PluginName string
}

func (e PluginNotFoundError) Error() string {
	return "Plugin {{.PluginName}} does not exist."
}

func (e PluginNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"PluginName": e.PluginName,
	})
}

type PluginNotFoundOnDiskOrInAnyRepositoryError struct {
	PluginName string
	BinaryName string
}

func (e PluginNotFoundOnDiskOrInAnyRepositoryError) Error() string {
	return "Plugin {{.PluginName}} not found on disk or in any registered repo.\nUse '{{.BinaryName}} repo-plugins' to list plugins available in the repos."
}

func (e PluginNotFoundOnDiskOrInAnyRepositoryError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"PluginName": e.PluginName,
		"BinaryName": e.BinaryName,
	})
}

// NoCompatibleBinaryError is returned when a repository contains a specified
// plugin but not for the specified platform
type NoCompatibleBinaryError struct {
}

func (e NoCompatibleBinaryError) Error() string {
	return "Plugin requested has no binary available for your platform."
}

func (e NoCompatibleBinaryError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}

type NoPluginRepositoriesError struct{}

func (_ NoPluginRepositoriesError) Error() string {
	return "No plugin repositories registered to search for plugin updates."
}

func (e NoPluginRepositoriesError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}

// GettingPluginRepositoryError is returned when there's an error
// accessing the plugin repository
type GettingPluginRepositoryError struct {
	Name    string
	Message string
}

func (_ GettingPluginRepositoryError) Error() string {
	return "Could not get plugin repository '{{.RepositoryName}}'\n{{.ErrorMessage}}"
}

func (e GettingPluginRepositoryError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{"RepositoryName": e.Name, "ErrorMessage": e.Message})
}

// RepositoryNameTakenError is returned when adding a plugin repository
// fails due to a repository already existing with the same name
type RepositoryNameTakenError struct {
	Name string
}

func (_ RepositoryNameTakenError) Error() string {
	return "Plugin repo named '{{.RepositoryName}}' already exists, please use another name."
}

func (e RepositoryNameTakenError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{"RepositoryName": e.Name})
}

type AddPluginRepositoryError struct {
	Name    string
	URL     string
	Message string
}

func (_ AddPluginRepositoryError) Error() string {
	return "Could not add repository '{{.RepositoryName}}' from {{.RepositoryURL}}: {{.Message}}"
}

func (e AddPluginRepositoryError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"RepositoryName": e.Name,
		"RepositoryURL":  e.URL,
		"Message":        e.Message,
	})
}

// FileNotFoundError is returned when a local plugin binary is not found during
// installation.
type FileNotFoundError struct {
	Path string
}

func (_ FileNotFoundError) Error() string {
	return "File not found locally, make sure the file exists at given path {{.FilePath}}"
}

func (e FileNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"FilePath": e.Path,
	})
}

// PluginInvalidError is returned with a plugin is invalid because it is
// missing a name or has 0 commands.
type PluginInvalidError struct {
}

func (e PluginInvalidError) Error() string {
	return "File is not a valid cf CLI plugin binary."
}

func (e PluginInvalidError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}

// PluginCommandConflictError is returned when a plugin command name conflicts
// with a native or existing plugin command name.
type PluginCommandsConflictError struct {
	PluginName     string
	PluginVersion  string
	CommandNames   []string
	CommandAliases []string
}

// PluginBinaryRemoveFailedError is returned when the removal of a plugin
// binary fails.
type PluginBinaryRemoveFailedError struct {
	Err error
}

func (e PluginBinaryRemoveFailedError) Error() string {
	return "The plugin has been uninstalled but removing the plugin binary failed.\nRemove it manually or subsequent installations of the plugin may fail\n{{.Err}}"
}

func (e PluginBinaryRemoveFailedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Err": e.Err,
	})
}

// PluginBinaryUninstallError is returned when running the plugin's uninstall
// hook fails.
type PluginBinaryUninstallError struct {
	Err error
}

func (e PluginBinaryUninstallError) Error() string {
	return "The plugin's uninstall method returned an unexpected error.\nThe plugin uninstall will proceed. Contact the plugin author if you need help.\n{{.Err}}"
}

func (e PluginBinaryUninstallError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Err": e.Err,
	})
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

// PluginAlreadyInstalledError is returned when the plugin has the same name as
// an installed plugin.
type PluginAlreadyInstalledError struct {
	BinaryName string
	Name       string
	Version    string
}

func (_ PluginAlreadyInstalledError) Error() string {
	return "Plugin {{.Name}} {{.Version}} could not be installed. A plugin with that name is already installed.\nTIP: Use '{{.BinaryName}} install-plugin -f' to force a reinstall."
}

func (e PluginAlreadyInstalledError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"BinaryName": e.BinaryName,
		"Name":       e.Name,
		"Version":    e.Version,
	})
}

type DownloadPluginHTTPError struct {
	Message string
}

func (_ DownloadPluginHTTPError) Error() string {
	return "Download attempt failed; server returned {{.ErrorMessage}}\nUnable to install; plugin is not available from the given URL."
}

func (e DownloadPluginHTTPError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ErrorMessage": e.Message,
	})
}

type FetchingPluginInfoFromRepositoriesError struct {
	Message        string
	RepositoryName string
}

func (_ FetchingPluginInfoFromRepositoriesError) Error() string {
	return "Plugin list download failed; repository {{.RepositoryName}} returned {{.ErrorMessage}}."
}

func (e FetchingPluginInfoFromRepositoriesError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"RepositoryName": e.RepositoryName,
		"ErrorMessage":   e.Message,
	})
}

type RepositoryNotRegisteredError struct {
	Name string
}

func (_ RepositoryNotRegisteredError) Error() string {
	return "Plugin repository {{.Name}} not found.\nUse 'cf list-plugin-repos' to list registered repos."
}

func (e RepositoryNotRegisteredError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Name": e.Name,
	})
}
