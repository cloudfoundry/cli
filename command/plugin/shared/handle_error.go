package shared

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

func HandleError(err error) error {
	switch e := err.(type) {
	case *json.SyntaxError:
		return translatableerror.JSONSyntaxError{Err: e}
	case pluginerror.RawHTTPStatusError:
		return translatableerror.DownloadPluginHTTPError{Message: e.Status}
	case pluginerror.SSLValidationHostnameError:
		return translatableerror.DownloadPluginHTTPError{Message: e.Error()}
	case pluginerror.UnverifiedServerError:
		return translatableerror.DownloadPluginHTTPError{Message: e.Error()}

	case pluginaction.AddPluginRepositoryError:
		return translatableerror.AddPluginRepositoryError{Name: e.Name, URL: e.URL, Message: e.Message}
	case pluginaction.GettingPluginRepositoryError:
		return translatableerror.GettingPluginRepositoryError{Name: e.Name, Message: e.Message}
	case pluginaction.NoCompatibleBinaryError:
		return translatableerror.NoCompatibleBinaryError{}
	case pluginaction.PluginCommandsConflictError:
		return translatableerror.PluginCommandsConflictError{
			PluginName:     e.PluginName,
			PluginVersion:  e.PluginVersion,
			CommandNames:   e.CommandNames,
			CommandAliases: e.CommandAliases,
		}
	case pluginaction.PluginInvalidError:
		return translatableerror.PluginInvalidError{Err: e.Err}
	case pluginaction.PluginNotFoundError:
		return translatableerror.PluginNotFoundError{PluginName: e.PluginName}
	case pluginaction.RepositoryNameTakenError:
		return translatableerror.RepositoryNameTakenError{Name: e.Name}
	case pluginaction.RepositoryNotRegisteredError:
		return translatableerror.RepositoryNotRegisteredError{Name: e.Name}

	case PluginInstallationCancelled:
		return nil
	}
	return err
}
