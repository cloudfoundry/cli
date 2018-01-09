package translatableerror

import (
	"encoding/json"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/util/clissh/ssherror"
	"code.cloudfoundry.org/cli/util/manifest"
	log "github.com/sirupsen/logrus"
)

func ConvertToTranslatableError(err error) error {
	log.WithField("err", fmt.Sprintf("%#v", err)).Debugf("convert to translatable error")

	switch e := err.(type) {
	// Action Errors
	case actionerror.AddPluginRepositoryError:
		return AddPluginRepositoryError(e)
	case actionerror.ApplicationNotFoundError:
		return ApplicationNotFoundError(e)
	case actionerror.ApplicationNotStartedError:
		return ApplicationNotStartedError(e)
	case actionerror.AppNotFoundInManifestError:
		return AppNotFoundInManifestError(e)
	case actionerror.AssignDropletError:
		return AssignDropletError(e)
	case actionerror.CommandLineOptionsWithMultipleAppsError:
		return CommandLineArgsWithMultipleAppsError{}
	case actionerror.DockerPasswordNotSetError:
		return DockerPasswordNotSetError{}
	case actionerror.DomainNotFoundError:
		return DomainNotFoundError(e)
	case actionerror.EmptyDirectoryError:
		return EmptyDirectoryError(e)
	case actionerror.FileChangedError:
		return FileChangedError(e)
	case actionerror.GettingPluginRepositoryError:
		return GettingPluginRepositoryError(e)
	case actionerror.HostnameWithTCPDomainError:
		return HostnameWithTCPDomainError(e)
	case actionerror.HTTPHealthCheckInvalidError:
		return HTTPHealthCheckInvalidError{}
	case actionerror.InvalidHTTPRouteSettings:
		return PortNotAllowedWithHTTPDomainError(e)
	case actionerror.InvalidRouteError:
		return InvalidRouteError(e)
	case actionerror.InvalidTCPRouteSettings:
		return HostAndPathNotAllowedWithTCPDomainError(e)
	case actionerror.IsolationSegmentNotFoundError:
		return IsolationSegmentNotFoundError(e)
	case actionerror.MissingNameError:
		return RequiredNameForPushError{}
	case actionerror.NoCompatibleBinaryError:
		return NoCompatibleBinaryError{}
	case actionerror.NoDomainsFoundError:
		return NoDomainsFoundError{}
	case actionerror.NoHostnameAndSharedDomainError:
		return NoHostnameAndSharedDomainError{}
	case actionerror.NoMatchingDomainError:
		return NoMatchingDomainError(e)
	case actionerror.NonexistentAppPathError:
		return FileNotFoundError(e)
	case actionerror.NoOrganizationTargetedError:
		return NoOrganizationTargetedError(e)
	case actionerror.NoSpaceTargetedError:
		return NoSpaceTargetedError(e)
	case actionerror.NotLoggedInError:
		return NotLoggedInError(e)
	case actionerror.OrganizationNotFoundError:
		return OrganizationNotFoundError(e)
	case actionerror.PluginCommandsConflictError:
		return PluginCommandsConflictError(e)
	case actionerror.PluginInvalidError:
		return PluginInvalidError(e)
	case actionerror.PluginNotFoundError:
		return PluginNotFoundError(e)
	case actionerror.ProcessInstanceNotFoundError:
		return ProcessInstanceNotFoundError(e)
	case actionerror.ProcessInstanceNotRunningError:
		return ProcessInstanceNotRunningError(e)
	case actionerror.ProcessNotFoundError:
		return ProcessNotFoundError(e)
	case actionerror.PropertyCombinationError:
		return PropertyCombinationError(e)
	case actionerror.RepositoryNameTakenError:
		return RepositoryNameTakenError(e)
	case actionerror.RepositoryNotRegisteredError:
		return RepositoryNotRegisteredError(e)
	case actionerror.RouteInDifferentSpaceError:
		return RouteInDifferentSpaceError(e)
	case actionerror.RoutePathWithTCPDomainError:
		return RoutePathWithTCPDomainError(e)
	case actionerror.SecurityGroupNotFoundError:
		return SecurityGroupNotFoundError(e)
	case actionerror.ServiceInstanceNotFoundError:
		return ServiceInstanceNotFoundError(e)
	case actionerror.ServiceInstanceNotSharedToSpaceError:
		return ServiceInstanceNotSharedToSpaceError{ServiceInstanceName: e.ServiceInstanceName}
	case actionerror.SharedServiceInstanceNotFoundError:
		return SharedServiceInstanceNotFoundError(e)
	case actionerror.SpaceNotFoundError:
		return SpaceNotFoundError{Name: e.Name}
	case actionerror.StackNotFoundError:
		return StackNotFoundError(e)
	case actionerror.StagingTimeoutError:
		return StagingTimeoutError(e)
	case actionerror.TaskWorkersUnavailableError:
		return RunTaskError{Message: "Task workers are unavailable."}
	case actionerror.TCPRouteOptionsNotProvidedError:
		return TCPRouteOptionsNotProvidedError{}
	case actionerror.TriggerLegacyPushError:
		return TriggerLegacyPushError{DomainHostRelated: e.DomainHostRelated}
	case actionerror.UploadFailedError:
		return UploadFailedError{Err: ConvertToTranslatableError(e.Err)}
	case actionerror.CommandLineOptionsAndManifestConflictError:
		return CommandLineOptionsAndManifestConflictError{
			ManifestAttribute:  e.ManifestAttribute,
			CommandLineOptions: e.CommandLineOptions,
		}

	// Generic CC Errors
	case ccerror.APINotFoundError:
		return APINotFoundError(e)
	case ccerror.RequestError:
		return APIRequestError(e)
	case ccerror.SSLValidationHostnameError:
		return SSLCertError(e)
	case ccerror.UnverifiedServerError:
		return InvalidSSLCertError(e)

	// Specific CC Errors
	case ccerror.JobFailedError:
		return JobFailedError(e)
	case ccerror.JobTimeoutError:
		return JobTimeoutError{JobGUID: e.JobGUID}
	case ccerror.UnprocessableEntityError:
		if strings.Contains(e.Message, "Task must have a droplet. Specify droplet or assign current droplet to app.") {
			return RunTaskError{Message: "App is not staged."}
		}

	// JSON Errors
	case *json.SyntaxError:
		return JSONSyntaxError{Err: e}

	// Manifest Errors
	case manifest.ManifestCreationError:
		return ManifestCreationError(e)
	case manifest.InheritanceFieldError:
		return TriggerLegacyPushError{InheritanceRelated: true}
	case manifest.GlobalFieldsError:
		return TriggerLegacyPushError{GlobalRelated: e.Fields}

	// Plugin Execution Errors
	case pluginerror.RawHTTPStatusError:
		return DownloadPluginHTTPError{Message: e.Status}
	case pluginerror.SSLValidationHostnameError:
		return DownloadPluginHTTPError(e)
	case pluginerror.UnverifiedServerError:
		return DownloadPluginHTTPError{Message: e.Error()}

	// SSH Errors
	case ssherror.UnableToAuthenticateError:
		return SSHUnableToAuthenticateError{}

	// UAA Errors
	case uaa.BadCredentialsError:
		return BadCredentialsError{}
	case uaa.InsufficientScopeError:
		return UnauthorizedToPerformActionError{}
	case uaa.InvalidAuthTokenError:
		return InvalidRefreshTokenError{}
	}

	return err
}
