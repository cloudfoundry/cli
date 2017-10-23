package shared

import (
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

func HandleError(err error) error {
	switch e := err.(type) {
	case ccerror.APINotFoundError:
		return translatableerror.APINotFoundError(e)
	case ccerror.RequestError:
		return translatableerror.APIRequestError(e)
	case ccerror.SSLValidationHostnameError:
		return translatableerror.SSLCertError(e)
	case ccerror.UnprocessableEntityError:
		if strings.Contains(e.Message, "Task must have a droplet. Specify droplet or assign current droplet to app.") {
			return translatableerror.RunTaskError{
				Message: "App is not staged."}
		}
	case ccerror.UnverifiedServerError:
		return translatableerror.InvalidSSLCertError{API: e.URL}

	case sharedaction.NotLoggedInError:
		return translatableerror.NotLoggedInError(e)
	case sharedaction.NoOrganizationTargetedError:
		return translatableerror.NoOrganizationTargetedError(e)
	case sharedaction.NoSpaceTargetedError:
		return translatableerror.NoSpaceTargetedError(e)
	case sharedaction.EmptyDirectoryError:
		return translatableerror.EmptyDirectoryError(e)

	case actionerror.ApplicationNotFoundError:
		return translatableerror.ApplicationNotFoundError(e)
	case actionerror.AssignDropletError:
		return translatableerror.AssignDropletError(e)
	case actionerror.IsolationSegmentNotFoundError:
		return translatableerror.IsolationSegmentNotFoundError(e)
	case actionerror.OrganizationNotFoundError:
		return translatableerror.OrganizationNotFoundError(e)
	case actionerror.ProcessInstanceNotFoundError:
		return translatableerror.ProcessInstanceNotFoundError(e)
	case actionerror.ProcessNotFoundError:
		return translatableerror.ProcessNotFoundError(e)
	case actionerror.StagingTimeoutError:
		return translatableerror.StagingTimeoutError(e)
	case actionerror.TaskWorkersUnavailableError:
		return translatableerror.RunTaskError{Message: "Task workers are unavailable."}
	}

	return err
}
