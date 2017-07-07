package shared

import (
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

func HandleError(err error) error {
	switch e := err.(type) {
	case ccerror.APINotFoundError:
		return translatableerror.APINotFoundError{URL: e.URL}
	case ccerror.RequestError:
		return translatableerror.APIRequestError{Err: e.Err}
	case ccerror.SSLValidationHostnameError:
		return translatableerror.SSLCertErrorError{Message: e.Message}
	case ccerror.UnprocessableEntityError:
		if strings.Contains(e.Message, "Task must have a droplet. Specify droplet or assign current droplet to app.") {
			return translatableerror.RunTaskError{
				Message: "App is not staged."}
		}
	case ccerror.UnverifiedServerError:
		return translatableerror.InvalidSSLCertError{API: e.URL}

	case sharedaction.NotLoggedInError:
		return translatableerror.NotLoggedInError{BinaryName: e.BinaryName}
	case sharedaction.NoOrganizationTargetedError:
		return translatableerror.NoOrganizationTargetedError{BinaryName: e.BinaryName}
	case sharedaction.NoSpaceTargetedError:
		return translatableerror.NoSpaceTargetedError{BinaryName: e.BinaryName}
	case v3action.ApplicationNotFoundError:
		return translatableerror.ApplicationNotFoundError{Name: e.Name}
	case v3action.TaskWorkersUnavailableError:
		return translatableerror.RunTaskError{Message: "Task workers are unavailable."}
	case v3action.OrganizationNotFoundError:
		return translatableerror.OrganizationNotFoundError{Name: e.Name}
	case v3action.IsolationSegmentNotFoundError:
		return translatableerror.IsolationSegmentNotFoundError{Name: e.Name}
	case v3action.AssignDropletError:
		return translatableerror.AssignDropletError{Message: e.Message}
	}

	return err
}
