package shared

import (
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command"
)

func HandleError(err error) error {
	switch e := err.(type) {
	case ccerror.APINotFoundError:
		return command.APINotFoundError{URL: e.URL}
	case ccerror.RequestError:
		return command.APIRequestError{Err: e.Err}
	case ccerror.SSLValidationHostnameError:
		return command.SSLCertErrorError{Message: e.Message}
	case ccerror.UnprocessableEntityError:
		if strings.Contains(e.Message, "Task must have a droplet. Specify droplet or assign current droplet to app.") {
			return RunTaskError{
				Message: "App is not staged."}
		}
	case ccerror.UnverifiedServerError:
		return command.InvalidSSLCertError{API: e.URL}

	case sharedaction.NotLoggedInError:
		return command.NotLoggedInError{BinaryName: e.BinaryName}
	case sharedaction.NoOrganizationTargetedError:
		return command.NoOrganizationTargetedError{BinaryName: e.BinaryName}
	case sharedaction.NoSpaceTargetedError:
		return command.NoSpaceTargetedError{BinaryName: e.BinaryName}
	case v3action.ApplicationNotFoundError:
		return command.ApplicationNotFoundError{Name: e.Name}
	case v3action.TaskWorkersUnavailableError:
		return RunTaskError{Message: "Task workers are unavailable."}
	case v3action.OrganizationNotFoundError:
		return OrganizationNotFoundError{Name: e.Name}
	case v3action.IsolationSegmentNotFoundError:
		return IsolationSegmentNotFoundError{Name: e.Name}
	case v3action.AssignDropletError:
		return AssignDropletError{Message: e.Message}
	}

	return err
}
