package shared

import (
	"code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
)

func HandleError(err error) error {
	switch e := err.(type) {
	case pluginerror.RawHTTPStatusError:
		return DownloadPluginHTTPError{Message: e.Status}
	case pluginerror.SSLValidationHostnameError:
		return DownloadPluginHTTPError{Message: e.Error()}
	case pluginerror.UnverifiedServerError:
		return DownloadPluginHTTPError{Message: e.Error()}

	case pluginaction.AddPluginRepositoryError:
		return AddPluginRepositoryError{Name: e.Name, URL: e.URL, Message: e.Message}
	case pluginaction.GettingPluginRepositoryError:
		return GettingPluginRepositoryError{Name: e.Name, Message: e.Message}
	case pluginaction.PluginCommandsConflictError:
		return PluginCommandsConflictError{
			PluginName:     e.PluginName,
			PluginVersion:  e.PluginVersion,
			CommandNames:   e.CommandNames,
			CommandAliases: e.CommandAliases,
		}
	case pluginaction.PluginInvalidError:
		return PluginInvalidError{}
	case pluginaction.PluginNotFoundError:
		return PluginNotFoundError{PluginName: e.PluginName, RepositoryName: e.RepositoryName}
	case pluginaction.RepositoryNameTakenError:
		return RepositoryNameTakenError{Name: e.Name}
	}
	return err
}
