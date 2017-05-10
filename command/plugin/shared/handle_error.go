package shared

import (
	"code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
)

func HandleError(err error) error {
	switch e := err.(type) {
	case pluginaction.PluginNotFoundError:
		return PluginNotFoundError{Name: e.Name}
	case pluginaction.GettingPluginRepositoryError:
		return GettingPluginRepositoryError{Name: e.Name, Message: e.Message}
	case pluginaction.RepositoryNameTakenError:
		return RepositoryNameTakenError{Name: e.Name}
	case pluginaction.AddPluginRepositoryError:
		return AddPluginRepositoryError{Name: e.Name, URL: e.URL, Message: e.Message}
	case pluginaction.PluginInvalidError:
		return PluginInvalidError{}
	case pluginaction.PluginCommandsConflictError:
		return PluginCommandsConflictError{
			PluginName:     e.PluginName,
			PluginVersion:  e.PluginVersion,
			CommandNames:   e.CommandNames,
			CommandAliases: e.CommandAliases,
		}
	case pluginerror.RawHTTPStatusError:
		return DownloadPluginRawHTTPStatusError{Status: e.Status}
	}
	return err
}
