package shared

import (
	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

func HandleError(err error) error {
	switch e := err.(type) {
	case ccerror.APINotFoundError:
		return translatableerror.APINotFoundError{URL: e.URL}
	case ccerror.RequestError:
		return translatableerror.APIRequestError{Err: e.Err}
	case ccerror.SSLValidationHostnameError:
		return translatableerror.SSLCertError{Message: e.Message}
	case ccerror.UnverifiedServerError:
		return translatableerror.InvalidSSLCertError{API: e.URL}

	case ccerror.JobFailedError:
		return translatableerror.JobFailedError{
			JobGUID: e.JobGUID,
			Message: e.Message,
		}
	case ccerror.JobTimeoutError:
		return translatableerror.JobTimeoutError{JobGUID: e.JobGUID}

	case uaa.BadCredentialsError:
		return translatableerror.BadCredentialsError{}
	case uaa.InvalidAuthTokenError:
		return translatableerror.InvalidRefreshTokenError{}

	case sharedaction.NotLoggedInError:
		return translatableerror.NotLoggedInError{BinaryName: e.BinaryName}
	case sharedaction.NoOrganizationTargetedError:
		return translatableerror.NoOrganizationTargetedError{BinaryName: e.BinaryName}
	case sharedaction.NoSpaceTargetedError:
		return translatableerror.NoSpaceTargetedError{BinaryName: e.BinaryName}

	case v2action.ApplicationNotFoundError:
		return translatableerror.ApplicationNotFoundError{Name: e.Name}
	case v2action.OrganizationNotFoundError:
		return translatableerror.OrganizationNotFoundError{Name: e.Name}
	case v2action.SecurityGroupNotFoundError:
		return translatableerror.SecurityGroupNotFoundError{Name: e.Name}
	case v2action.ServiceInstanceNotFoundError:
		return translatableerror.ServiceInstanceNotFoundError{Name: e.Name}
	case v2action.SpaceNotFoundError:
		return translatableerror.SpaceNotFoundError{Name: e.Name}
	case v2action.HTTPHealthCheckInvalidError:
		return translatableerror.HTTPHealthCheckInvalidError{}
	case v2action.RouteInDifferentSpaceError:
		return translatableerror.RouteInDifferentSpaceError{Route: e.Route}
	case v2action.FileChangedError:
		return translatableerror.FileChangedError{Filename: e.Filename}
	case v2action.EmptyDirectoryError:
		return translatableerror.EmptyDirectoryError{Path: e.Path}

	case pushaction.NoDomainsFoundError:
		return translatableerror.NoDomainsFoundError{}
	case pushaction.UploadFailedError:
		return translatableerror.UploadFailedError{Err: HandleError(e.Err)}
	}

	return err
}
