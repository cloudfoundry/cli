package shared

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/actor/actionerror"
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
		return translatableerror.DownloadPluginHTTPError(e)
	case pluginerror.UnverifiedServerError:
		return translatableerror.DownloadPluginHTTPError{Message: e.Error()}

	case actionerror.AddPluginRepositoryError:
		return translatableerror.AddPluginRepositoryError(e)
	case actionerror.GettingPluginRepositoryError:
		return translatableerror.GettingPluginRepositoryError(e)
	case actionerror.NoCompatibleBinaryError:
		return translatableerror.NoCompatibleBinaryError{}
	case actionerror.PluginCommandsConflictError:
		return translatableerror.PluginCommandsConflictError(e)
	case actionerror.PluginInvalidError:
		return translatableerror.PluginInvalidError(e)
	case actionerror.PluginNotFoundError:
		return translatableerror.PluginNotFoundError(e)
	case actionerror.RepositoryNameTakenError:
		return translatableerror.RepositoryNameTakenError(e)
	case actionerror.RepositoryNotRegisteredError:
		return translatableerror.RepositoryNotRegisteredError(e)

	case PluginInstallationCancelled:
		return nil
	}
	return err
}
