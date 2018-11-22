package translatableerror_test

import (
	"encoding/json"
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
	"code.cloudfoundry.org/cli/api/uaa"
	. "code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/clissh/ssherror"
	"code.cloudfoundry.org/cli/util/download"
	"code.cloudfoundry.org/cli/util/manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConvertToTranslatableError", func() {
	err := errors.New("some-error")
	genericErr := errors.New("some-generic-error")
	jsonErr := new(json.SyntaxError)
	unprocessableEntityError := ccerror.UnprocessableEntityError{Message: "another message"}

	DescribeTable("error translations",
		func(passedInErr error, expectedErr error) {
			actualErr := ConvertToTranslatableError(passedInErr)
			Expect(actualErr).To(MatchError(expectedErr))
		},

		// Action Errors
		Entry("actionerror.AddPluginRepositoryError -> AddPluginRepositoryError",
			actionerror.AddPluginRepositoryError{Name: "some-repo", URL: "some-URL", Message: "404"},
			AddPluginRepositoryError{Name: "some-repo", URL: "some-URL", Message: "404"}),

		Entry("actionerror.ApplicationNotFoundError -> ApplicationNotFoundError",
			actionerror.ApplicationNotFoundError{Name: "some-app"},
			ApplicationNotFoundError{Name: "some-app"}),

		Entry("actionerror.ApplicationNotStartedError -> ApplicationNotStartedError",
			actionerror.ApplicationNotStartedError{Name: "some-app"},
			ApplicationNotStartedError{Name: "some-app"}),

		Entry("actionerror.AppNotFoundInManifestError -> AppNotFoundInManifestError",
			actionerror.AppNotFoundInManifestError{Name: "some-app"},
			AppNotFoundInManifestError{Name: "some-app"}),

		Entry("actionerror.AssignDropletError -> AssignDropletError",
			actionerror.AssignDropletError{Message: "some-message"},
			AssignDropletError{Message: "some-message"}),

		Entry("actionerror.ServicePlanNotFoundError -> ServicePlanNotFoundError",
			actionerror.ServicePlanNotFoundError{PlanName: "some-plan", ServiceName: "some-service"},
			ServicePlanNotFoundError{PlanName: "some-plan", ServiceName: "some-service"}),

		Entry("actionerror.BuildpackNotFoundError -> BuildpackNotFoundError",
			actionerror.BuildpackNotFoundError{},
			BuildpackNotFoundError{}),

		Entry("actionerror.BuildpackStackChangeError-> BuildpackStackChangeError",
			actionerror.BuildpackStackChangeError{},
			BuildpackStackChangeError{}),

		Entry("actionerror.CommandLineOptionsWithMultipleAppsError -> CommandLineArgsWithMultipleAppsError",
			actionerror.CommandLineOptionsWithMultipleAppsError{},
			CommandLineArgsWithMultipleAppsError{}),

		Entry("actionerror.DockerPasswordNotSetError -> DockerPasswordNotSetError",
			actionerror.DockerPasswordNotSetError{},
			DockerPasswordNotSetError{}),

		Entry("actionerror.DomainNotFoundError -> DomainNotFoundError",
			actionerror.DomainNotFoundError{Name: "some-domain-name", GUID: "some-domain-guid"},
			DomainNotFoundError{Name: "some-domain-name", GUID: "some-domain-guid"}),

		Entry("actionerror.EmptyBuildpacksError -> EmptyBuildpacksError",
			manifest.EmptyBuildpacksError{},
			EmptyBuildpacksError{},
		),

		Entry("actionerror.EmptyDirectoryError -> EmptyDirectoryError",
			actionerror.EmptyDirectoryError{Path: "some-filename"},
			EmptyDirectoryError{Path: "some-filename"}),

		Entry("actionerror.EmptyBuildpackDirectoryError -> EmptyBuildpackDirectoryError",
			actionerror.EmptyBuildpackDirectoryError{Path: "some-path"},
			EmptyBuildpackDirectoryError{Path: "some-path"}),

		Entry("actionerror.FileChangedError -> FileChangedError",
			actionerror.FileChangedError{Filename: "some-filename"},
			FileChangedError{Filename: "some-filename"}),

		Entry("actionerror.GettingPluginRepositoryError -> GettingPluginRepositoryError",
			actionerror.GettingPluginRepositoryError{Name: "some-repo", Message: "404"},
			GettingPluginRepositoryError{Name: "some-repo", Message: "404"}),

		Entry("actionerror.HostnameWithTCPDomainError -> HostnameWithTCPDomainError",
			actionerror.HostnameWithTCPDomainError{},
			HostnameWithTCPDomainError{}),

		Entry("actionerror.HTTPHealthCheckInvalidError -> HTTPHealthCheckInvalidError",
			actionerror.HTTPHealthCheckInvalidError{},
			HTTPHealthCheckInvalidError{}),

		Entry("actionerror.InvalidBuildpacksError -> InvalidBuildpacksError",
			actionerror.InvalidBuildpacksError{},
			InvalidBuildpacksError{}),

		Entry("actionerror.InvalidHTTPRouteSettings -> PortNotAllowedWithHTTPDomainError",
			actionerror.InvalidHTTPRouteSettings{Domain: "some-domain"},
			PortNotAllowedWithHTTPDomainError{Domain: "some-domain"}),

		Entry("actionerror.InvalidRouteError -> InvalidRouteError",
			actionerror.InvalidRouteError{Route: "some-invalid-route"},
			InvalidRouteError{Route: "some-invalid-route"}),

		Entry("actionerror.InvalidTCPRouteSettings -> HostAndPathNotAllowedWithTCPDomainError",
			actionerror.InvalidTCPRouteSettings{Domain: "some-domain"},
			HostAndPathNotAllowedWithTCPDomainError{Domain: "some-domain"}),

		Entry("actionerror.MissingNameError -> RequiredNameForPushError",
			actionerror.MissingNameError{},
			RequiredNameForPushError{}),

		Entry("actionerror.MultipleBuildpacksFoundError -> MultipleBuildpacksFoundError",
			actionerror.MultipleBuildpacksFoundError{BuildpackName: "some-bp-name"},
			MultipleBuildpacksFoundError{BuildpackName: "some-bp-name"}),

		Entry("actionerror.NoCompatibleBinaryError -> NoCompatibleBinaryError",
			actionerror.NoCompatibleBinaryError{},
			NoCompatibleBinaryError{}),

		Entry("actionerror.NoDomainsFoundError -> NoDomainsFoundError",
			actionerror.NoDomainsFoundError{OrganizationGUID: "some-guid"},
			NoDomainsFoundError{}),

		Entry("actionerror.NoHostnameAndSharedDomainError -> NoHostnameAndSharedDomainError",
			actionerror.NoHostnameAndSharedDomainError{},
			NoHostnameAndSharedDomainError{}),

		Entry("actionerror.NoMatchingDomainError -> NoMatchingDomainError",
			actionerror.NoMatchingDomainError{Route: "some-route.com"},
			NoMatchingDomainError{Route: "some-route.com"}),

		Entry("actionerror.NonexistentAppPathError -> FileNotFoundError",
			actionerror.NonexistentAppPathError{Path: "some-path"},
			FileNotFoundError{Path: "some-path"}),

		Entry("actionerror.NoOrganizationTargetedError -> NoOrganizationTargetedError",
			actionerror.NoOrganizationTargetedError{BinaryName: "faceman"},
			NoOrganizationTargetedError{BinaryName: "faceman"}),

		Entry("actionerror.NoSpaceTargetedError -> NoSpaceTargetedError",
			actionerror.NoSpaceTargetedError{BinaryName: "faceman"},
			NoSpaceTargetedError{BinaryName: "faceman"}),

		Entry("actionerror.NotLoggedInError -> NotLoggedInError",
			actionerror.NotLoggedInError{BinaryName: "faceman"},
			NotLoggedInError{BinaryName: "faceman"}),

		Entry("actionerror.OrganizationNotFoundError -> OrgNotFoundError",
			actionerror.OrganizationNotFoundError{Name: "some-org"},
			OrganizationNotFoundError{Name: "some-org"}),

		Entry("actionerror.OrganizationQuotaNotFoundForNameError -> OrganizationQuotaNotFoundForNameError",
			actionerror.OrganizationQuotaNotFoundForNameError{Name: "some-quota"},
			OrganizationQuotaNotFoundForNameError{Name: "some-quota"}),

		Entry("actionerror.PasswordGrantTypeLogoutRequiredError -> PasswordGrantTypeLogoutRequiredError",
			actionerror.PasswordGrantTypeLogoutRequiredError{},
			PasswordGrantTypeLogoutRequiredError{}),

		Entry("actionerror.PluginCommandConflictError -> PluginCommandConflictError",
			actionerror.PluginCommandsConflictError{
				PluginName:     "some-plugin",
				PluginVersion:  "1.1.1",
				CommandNames:   []string{"some-command", "some-other-command"},
				CommandAliases: []string{"sc", "soc"},
			},
			PluginCommandsConflictError{
				PluginName:     "some-plugin",
				PluginVersion:  "1.1.1",
				CommandNames:   []string{"some-command", "some-other-command"},
				CommandAliases: []string{"sc", "soc"},
			}),

		Entry("actionerror.PluginInvalidError -> PluginInvalidError",
			actionerror.PluginInvalidError{},
			PluginInvalidError{}),

		Entry("actionerror.PluginInvalidError -> PluginInvalidError",
			actionerror.PluginInvalidError{Err: genericErr},
			PluginInvalidError{Err: genericErr}),

		Entry("actionerror.PluginNotFoundError -> PluginNotFoundError",
			actionerror.PluginNotFoundError{PluginName: "some-plugin"},
			PluginNotFoundError{PluginName: "some-plugin"}),

		Entry("actionerror.ProcessInstanceNotFoundError -> ProcessInstanceNotFoundError",
			actionerror.ProcessInstanceNotFoundError{ProcessType: "some-process-type", InstanceIndex: 42},
			ProcessInstanceNotFoundError{ProcessType: "some-process-type", InstanceIndex: 42}),

		Entry("actionerror.ProcessInstanceNotRunningError -> ProcessInstanceNotRunningError",
			actionerror.ProcessInstanceNotRunningError{ProcessType: "some-process-type", InstanceIndex: 42},
			ProcessInstanceNotRunningError{ProcessType: "some-process-type", InstanceIndex: 42}),

		Entry("actionerror.ProcessNotFoundError -> ProcessNotFoundError",
			actionerror.ProcessNotFoundError{ProcessType: "some-process-type"},
			ProcessNotFoundError{ProcessType: "some-process-type"}),

		Entry("actionerror.PropertyCombinationError -> PropertyCombinationError",
			actionerror.PropertyCombinationError{Properties: []string{"property-1", "property-2"}},
			PropertyCombinationError{Properties: []string{"property-1", "property-2"}}),

		Entry("actionerror.RepositoryNameTakenError -> RepositoryNameTakenError",
			actionerror.RepositoryNameTakenError{Name: "some-repo"},
			RepositoryNameTakenError{Name: "some-repo"}),

		Entry("actionerror.RepositoryNotRegisteredError -> RepositoryNotRegisteredError",
			actionerror.RepositoryNotRegisteredError{Name: "some-repo"},
			RepositoryNotRegisteredError{Name: "some-repo"}),

		Entry("actionerror.RouteInDifferentSpaceError -> RouteInDifferentSpaceError",
			actionerror.RouteInDifferentSpaceError{Route: "some-route"},
			RouteInDifferentSpaceError{Route: "some-route"}),

		Entry("actionerror.RoutePathWithTCPDomainError -> RoutePathWithTCPDomainError",
			actionerror.RoutePathWithTCPDomainError{},
			RoutePathWithTCPDomainError{}),

		Entry("actionerror.RouterGroupNotFoundError -> RouterGroupNotFoundError",
			actionerror.RouterGroupNotFoundError{Name: "some-group"},
			RouterGroupNotFoundError{Name: "some-group"},
		),

		Entry("actionerror.SecurityGroupNotFoundError -> SecurityGroupNotFoundError",
			actionerror.SecurityGroupNotFoundError{Name: "some-security-group"},
			SecurityGroupNotFoundError{Name: "some-security-group"}),

		Entry("actionerror.ServiceInstanceNotFoundError -> ServiceInstanceNotFoundError",
			actionerror.ServiceInstanceNotFoundError{Name: "some-service-instance"},
			ServiceInstanceNotFoundError{Name: "some-service-instance"}),

		Entry("actionerror.ServiceInstanceNotShareableError -> ServiceInstanceNotShareableError",
			actionerror.ServiceInstanceNotShareableError{
				FeatureFlagEnabled:          true,
				ServiceBrokerSharingEnabled: false},
			ServiceInstanceNotShareableError{
				FeatureFlagEnabled:          true,
				ServiceBrokerSharingEnabled: false}),

		Entry("actionerror.SharedServiceInstanceNotFoundError -> SharedServiceInstanceNotFoundError",
			actionerror.SharedServiceInstanceNotFoundError{},
			SharedServiceInstanceNotFoundError{}),

		Entry("actionerror.SpaceNotFoundError -> SpaceNotFoundError",
			actionerror.SpaceNotFoundError{Name: "some-space"},
			SpaceNotFoundError{Name: "some-space"}),

		Entry("actionerror.SpaceQuotaNotFoundByNameError -> SpaceQuotaNotFoundByNameError",
			actionerror.SpaceQuotaNotFoundByNameError{Name: "some-space"},
			SpaceQuotaNotFoundByNameError{Name: "some-space"}),

		Entry("actionerror.StackNotFoundError -> StackNotFoundError",
			actionerror.StackNotFoundError{Name: "some-stack-name", GUID: "some-stack-guid"},
			StackNotFoundError{Name: "some-stack-name", GUID: "some-stack-guid"}),

		Entry("actionerror.TaskWorkersUnavailableError -> RunTaskError",
			actionerror.TaskWorkersUnavailableError{Message: "fooo: Banana Pants"},
			RunTaskError{Message: "Task workers are unavailable."}),

		Entry("actionerror.TCPRouteOptionsNotProvidedError-> TCPRouteOptionsNotProvidedError",
			actionerror.TCPRouteOptionsNotProvidedError{},
			TCPRouteOptionsNotProvidedError{}),

		Entry("actionerror.TriggerLegacyPushError -> TriggerLegacyPushError",
			actionerror.TriggerLegacyPushError{DomainHostRelated: []string{"domain", "host"}},
			TriggerLegacyPushError{DomainHostRelated: []string{"domain", "host"}}),

		Entry("actionerror.UploadFailedError -> UploadFailedError",
			actionerror.UploadFailedError{Err: actionerror.NoDomainsFoundError{}},
			UploadFailedError{Err: NoDomainsFoundError{}}),

		Entry("v3action.StagingTimeoutError -> StagingTimeoutError",
			actionerror.StagingTimeoutError{AppName: "some-app", Timeout: time.Nanosecond},
			StagingTimeoutError{AppName: "some-app", Timeout: time.Nanosecond}),

		Entry("actionerror.CommandLineOptionsAndManifestConflictError -> CommandLineOptionsAndManifestConflictError",
			actionerror.CommandLineOptionsAndManifestConflictError{
				ManifestAttribute:  "some-attribute",
				CommandLineOptions: []string{"option-1", "option-2"},
			},
			CommandLineOptionsAndManifestConflictError{
				ManifestAttribute:  "some-attribute",
				CommandLineOptions: []string{"option-1", "option-2"},
			}),

		Entry("actionerror.ServiceInstanceNotSharedToSpaceError -> ServiceInstanceNotSharedToSpaceError",
			actionerror.ServiceInstanceNotSharedToSpaceError{ServiceInstanceName: "some-service-instance-name"},
			ServiceInstanceNotSharedToSpaceError{ServiceInstanceName: "some-service-instance-name"}),

		// CC Errors
		Entry("ccerror.APINotFoundError -> APINotFoundError",
			ccerror.APINotFoundError{URL: "some-url"},
			APINotFoundError{URL: "some-url"}),

		Entry("ccerror.JobFailedError -> JobFailedError",
			ccerror.JobFailedError{JobGUID: "some-job-guid", Message: "some-message"},
			JobFailedError{JobGUID: "some-job-guid", Message: "some-message"}),

		Entry("ccerror.JobTimeoutError -> JobTimeoutError",
			ccerror.JobTimeoutError{JobGUID: "some-job-guid"},
			JobTimeoutError{JobGUID: "some-job-guid"}),

		Entry("ccerror.MultiError -> MultiError",
			ccerror.MultiError{ResponseCode: 418, Errors: []ccerror.V3Error{
				{
					Code:   282010,
					Detail: "detail 1",
					Title:  "title-1",
				},
				{
					Code:   10242013,
					Detail: "detail 2",
					Title:  "title-2",
				},
			}},
			MultiError{Messages: []string{"detail 1", "detail 2"}},
		),

		Entry("ccerror.RequestError -> APIRequestError",
			ccerror.RequestError{Err: err},
			APIRequestError{Err: err}),

		Entry("ccerror.SSLValidationHostnameError -> SSLCertErrorError",
			ccerror.SSLValidationHostnameError{Message: "some-message"},
			SSLCertError{Message: "some-message"}),

		Entry("ccerror.UnverifiedServerError -> InvalidSSLCertError",
			ccerror.UnverifiedServerError{URL: "some-url"},
			InvalidSSLCertError{URL: "some-url"}),

		Entry("ccerror.UnprocessableEntityError with droplet message -> RunTaskError",
			ccerror.UnprocessableEntityError{Message: "The request is semantically invalid: Task must have a droplet. Specify droplet or assign current droplet to app."},
			RunTaskError{Message: "App is not staged."}),

		// This changed in CF254
		Entry("ccerror.UnprocessableEntityError with droplet message -> RunTaskError",
			ccerror.UnprocessableEntityError{Message: "Task must have a droplet. Specify droplet or assign current droplet to app."},
			RunTaskError{Message: "App is not staged."}),

		Entry("ccerror.UnprocessableEntityError without droplet message -> original error",
			unprocessableEntityError,
			unprocessableEntityError),

		Entry("download.RawHTTPStatusError -> HTTPStatusError",
			download.RawHTTPStatusError{Status: "some status"},
			HTTPStatusError{Status: "some status"},
		),

		Entry("json.SyntaxError -> JSONSyntaxError",
			jsonErr,
			JSONSyntaxError{Err: jsonErr},
		),

		// Manifest Errors
		Entry("manifest.ManifestCreationError -> ManifestCreationError",
			manifest.ManifestCreationError{Err: errors.New("some-error")},
			ManifestCreationError{Err: errors.New("some-error")}),

		Entry("manifest.InheritanceFieldError -> TriggerLegacyPushError",
			manifest.InheritanceFieldError{},
			TriggerLegacyPushError{InheritanceRelated: true}),

		Entry("manifest.GlobalFieldError -> TriggerLegacyPushError",
			manifest.GlobalFieldsError{Fields: []string{"some-field"}},
			TriggerLegacyPushError{GlobalRelated: []string{"some-field"}}),

		Entry("manifest.InterpolationError -> InterpolationError",
			manifest.InterpolationError{Err: errors.New("an-error")},
			InterpolationError{Err: errors.New("an-error")}),

		// Plugin Errors
		Entry("pluginerror.RawHTTPStatusError -> DownloadPluginHTTPError",
			pluginerror.RawHTTPStatusError{Status: "some status"},
			DownloadPluginHTTPError{Message: "some status"},
		),
		Entry("pluginerror.SSLValidationHostnameError -> DownloadPluginHTTPError",
			pluginerror.SSLValidationHostnameError{Message: "some message"},
			DownloadPluginHTTPError{Message: "some message"},
		),
		Entry("pluginerror.UnverifiedServerError -> DownloadPluginHTTPError",
			pluginerror.UnverifiedServerError{URL: "some URL"},
			DownloadPluginHTTPError{Message: "x509: certificate signed by unknown authority"},
		),

		// SSH Error
		Entry("ssherror.UnableToAuthenticateError -> UnableToAuthenticateError",
			ssherror.UnableToAuthenticateError{},
			SSHUnableToAuthenticateError{}),

		// UAA Errors
		Entry("uaa.BadCredentialsError -> BadCredentialsError",
			uaa.UnauthorizedError{},
			UnauthorizedError{}),

		Entry("uaa.InsufficientScopeError -> UnauthorizedToPerformActionError",
			uaa.InsufficientScopeError{},
			UnauthorizedToPerformActionError{}),

		Entry("uaa.InvalidAuthTokenError -> InvalidRefreshTokenError",
			uaa.InvalidAuthTokenError{},
			InvalidRefreshTokenError{}),

		Entry("default case -> original error",
			err,
			err),
	)

	It("returns nil for a nil error", func() {
		nilErr := ConvertToTranslatableError(nil)
		Expect(nilErr).To(BeNil())
	})
})
