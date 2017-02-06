package shared

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/command"
)

func HandleError(err error) error {
	switch e := err.(type) {
	case cloudcontroller.APINotFoundError:
		return command.APINotFoundError{URL: e.URL}
	case cloudcontroller.RequestError:
		return command.APIRequestError{Err: e.Err}
	case cloudcontroller.SSLValidationHostnameError:
		return command.SSLCertErrorError{Message: e.Message}
	case cloudcontroller.UnverifiedServerError:
		return command.InvalidSSLCertError{API: e.URL}

	case ccv2.JobFailedError:
		return JobFailedError{JobGUID: e.JobGUID}
	case ccv2.JobTimeoutError:
		return JobTimeoutError{JobGUID: e.JobGUID}

	case sharedaction.NotLoggedInError:
		return command.NotLoggedInError{BinaryName: e.BinaryName}
	case sharedaction.NoTargetedOrganizationError:
		return command.NoTargetedOrganizationError{BinaryName: e.BinaryName}
	case sharedaction.NoTargetedSpaceError:
		return command.NoTargetedSpaceError{BinaryName: e.BinaryName}

	case v2action.ApplicationNotFoundError:
		return command.ApplicationNotFoundError{Name: e.Name}
	case v2action.OrganizationNotFoundError:
		return OrganizationNotFoundError{Name: e.Name}
	case v2action.ServiceInstanceNotFoundError:
		return command.ServiceInstanceNotFoundError{Name: e.Name}
	case v2action.SpaceNotFoundError:
		return SpaceNotFoundError{Name: e.Name}
	case v2action.HTTPHealthCheckInvalidError:
		return HTTPHealthCheckInvalidError{}
	}

	return err
}
