package shared

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
)

func HandleError(err error) error {
	switch e := err.(type) {
	case *json.SyntaxError:
		return JSONSyntaxError{Err: e}
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
	case pluginaction.NoCompatibleBinaryError:
		return NoCompatibleBinaryError{}
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
		return PluginNotFoundError{PluginName: e.PluginName}
	case pluginaction.RepositoryNameTakenError:
		return RepositoryNameTakenError{Name: e.Name}
	case pluginaction.RepositoryNotRegisteredError:
		return RepositoryNotRegisteredError{Name: e.Name}

	case PluginInstallationCancelled:
		return nil
	}
	return err
}
