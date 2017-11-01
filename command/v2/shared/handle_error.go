package shared

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifest"
)

func HandleError(err error) error {
	switch e := err.(type) {
	case ccerror.APINotFoundError:
		return translatableerror.APINotFoundError(e)
	case ccerror.RequestError:
		return translatableerror.APIRequestError(e)
	case ccerror.SSLValidationHostnameError:
		return translatableerror.SSLCertError(e)
	case ccerror.UnverifiedServerError:
		return translatableerror.InvalidSSLCertError{API: e.URL}

	case ccerror.JobFailedError:
		return translatableerror.JobFailedError(e)
	case ccerror.JobTimeoutError:
		return translatableerror.JobTimeoutError{JobGUID: e.JobGUID}

	case uaa.BadCredentialsError:
		return translatableerror.BadCredentialsError{}
	case uaa.InvalidAuthTokenError:
		return translatableerror.InvalidRefreshTokenError{}

	case actionerror.NotLoggedInError:
		return translatableerror.NotLoggedInError(e)
	case actionerror.NoOrganizationTargetedError:
		return translatableerror.NoOrganizationTargetedError(e)
	case actionerror.NoSpaceTargetedError:
		return translatableerror.NoSpaceTargetedError(e)

	case actionerror.ApplicationNotFoundError:
		return translatableerror.ApplicationNotFoundError{Name: e.Name}
	case actionerror.OrganizationNotFoundError:
		return translatableerror.OrganizationNotFoundError{Name: e.Name}
	case actionerror.SecurityGroupNotFoundError:
		return translatableerror.SecurityGroupNotFoundError(e)
	case actionerror.ServiceInstanceNotFoundError:
		return translatableerror.ServiceInstanceNotFoundError(e)
	case actionerror.SpaceNotFoundError:
		return translatableerror.SpaceNotFoundError{Name: e.Name}
	case actionerror.StackNotFoundError:
		return translatableerror.StackNotFoundError(e)
	case actionerror.HTTPHealthCheckInvalidError:
		return translatableerror.HTTPHealthCheckInvalidError{}
	case actionerror.RouteInDifferentSpaceError:
		return translatableerror.RouteInDifferentSpaceError(e)
	case actionerror.FileChangedError:
		return translatableerror.FileChangedError(e)
	case actionerror.EmptyDirectoryError:
		return translatableerror.EmptyDirectoryError(e)
	case actionerror.DomainNotFoundError:
		return translatableerror.DomainNotFoundError(e)
	case actionerror.NoMatchingDomainError:
		return translatableerror.NoMatchingDomainError(e)

	case actionerror.HostnameWithTCPDomainError:
		return translatableerror.HostnameWithTCPDomainError(e)
	case actionerror.InvalidHTTPRouteSettings:
		return translatableerror.PortNotAllowedWithHTTPDomainError(e)

	case actionerror.AppNotFoundInManifestError:
		return translatableerror.AppNotFoundInManifestError(e)
	case actionerror.CommandLineOptionsWithMultipleAppsError:
		return translatableerror.CommandLineArgsWithMultipleAppsError{}
	case actionerror.NoDomainsFoundError:
		return translatableerror.NoDomainsFoundError{}
	case actionerror.NoHostnameAndSharedDomainError:
		return translatableerror.NoHostnameAndSharedDomainError{}
	case actionerror.NonexistentAppPathError:
		return translatableerror.FileNotFoundError(e)
	case actionerror.MissingNameError:
		return translatableerror.RequiredNameForPushError{}
	case actionerror.UploadFailedError:
		return translatableerror.UploadFailedError{Err: HandleError(e.Err)}
	case actionerror.PropertyCombinationError:
		return translatableerror.PropertyCombinationError(e)
	case actionerror.DockerPasswordNotSetError:
		return translatableerror.DockerPasswordNotSetError{}
	case manifest.ManifestCreationError:
		return translatableerror.ManifestCreationError(e)
	case uaa.InsufficientScopeError:
		return translatableerror.UnauthorizedToPerformActionError{}
	}

	return err
}
