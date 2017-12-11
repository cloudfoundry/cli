package translatableerror_test

import (
	"bytes"
	"errors"
	"text/template"

	. "code.cloudfoundry.org/cli/command/translatableerror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Translatable Errors", func() {
	translateFunc := func(s string, vars ...interface{}) string {
		formattedTemplate, err := template.New("test template").Parse(s)
		Expect(err).ToNot(HaveOccurred())
		formattedTemplate.Option("missingkey=error")

		var buffer bytes.Buffer
		if len(vars) > 0 {
			err = formattedTemplate.Execute(&buffer, vars[0])
			Expect(err).ToNot(HaveOccurred())

			return buffer.String()
		} else {
			return s
		}
	}

	DescribeTable("translates error",
		func(e error) {
			err, ok := e.(TranslatableError)
			Expect(ok).To(BeTrue())
			err.Translate(translateFunc)
		},

		Entry("AddPluginRepositoryError", AddPluginRepositoryError{}),
		Entry("APINotFoundError", APINotFoundError{}),
		Entry("APIRequestError", APIRequestError{}),
		Entry("ApplicationNotFoundError", ApplicationNotFoundError{}),
		Entry("AppNotFoundInManifestError", AppNotFoundInManifestError{}),
		Entry("ArgumentCombinationError", ArgumentCombinationError{}),
		Entry("AssignDropletError", AssignDropletError{}),
		Entry("BadCredentialsError", BadCredentialsError{}),
		Entry("CFNetworkingEndpointNotFoundError", CFNetworkingEndpointNotFoundError{}),
		Entry("CommandLineArgsWithMultipleAppsError", CommandLineArgsWithMultipleAppsError{}),
		Entry("CommandLineOptionsAndManifestConflictError", CommandLineOptionsAndManifestConflictError{}),
		Entry("DockerPasswordNotSetError", DockerPasswordNotSetError{}),
		Entry("DownloadPluginHTTPError", DownloadPluginHTTPError{}),
		Entry("EmptyDirectoryError", EmptyDirectoryError{}),
		Entry("FetchingPluginInfoFromRepositoriesError", FetchingPluginInfoFromRepositoriesError{}),
		Entry("FileChangedError", FileChangedError{}),
		Entry("FileNotFoundError", FileNotFoundError{}),
		Entry("GettingPluginRepositoryError", GettingPluginRepositoryError{}),
		Entry("HealthCheckTypeUnsupportedError", HealthCheckTypeUnsupportedError{SupportedTypes: []string{"some-type", "another-type"}}),
		Entry("HostAndPathNotAllowedWithTCPDomainError", HostAndPathNotAllowedWithTCPDomainError{}),
		Entry("HostnameWithTCPDomainError", HostnameWithTCPDomainError{}),
		Entry("HTTPHealthCheckInvalidError", HTTPHealthCheckInvalidError{}),
		Entry("InvalidChecksumError", InvalidChecksumError{}),
		Entry("InvalidRouteError", InvalidRouteError{}),
		Entry("InvalidSSLCertError", InvalidSSLCertError{}),
		Entry("IsolationSegmentNotFoundError", IsolationSegmentNotFoundError{}),
		Entry("JobFailedError", JobFailedError{}),
		Entry("JobTimeoutError", JobTimeoutError{}),
		Entry("JSONSyntaxError", JSONSyntaxError{Err: errors.New("some-error")}),
		Entry("LifecycleMinimumAPIVersionNotMetError", LifecycleMinimumAPIVersionNotMetError{}),
		Entry("ManifestCreationError", ManifestCreationError{}),
		Entry("ManifestFileNotFoundInDirectoryError", ManifestFileNotFoundInDirectoryError{}),
		Entry("MinimumAPIVersionNotMetError", MinimumAPIVersionNotMetError{}),
		Entry("NetworkPolicyProtocolOrPortNotProvidedError", NetworkPolicyProtocolOrPortNotProvidedError{}),
		Entry("NoAPISetError", NoAPISetError{}),
		Entry("NoCompatibleBinaryError", NoCompatibleBinaryError{}),
		Entry("NoDomainsFoundError", NoDomainsFoundError{}),
		Entry("NoMatchingDomainError", NoMatchingDomainError{}),
		Entry("NoOrganizationTargetedError", NoOrganizationTargetedError{}),
		Entry("NoPluginRepositoriesError", NoPluginRepositoriesError{}),
		Entry("NoSpaceTargetedError", NoSpaceTargetedError{}),
		Entry("NotLoggedInError", NotLoggedInError{}),
		Entry("OrgNotFoundError", OrganizationNotFoundError{}),
		Entry("ParseArgumentError", ParseArgumentError{}),
		Entry("PluginAlreadyInstalledError", PluginAlreadyInstalledError{}),
		Entry("PluginBinaryRemoveFailedError", PluginBinaryRemoveFailedError{}),
		Entry("PluginBinaryUninstallError", PluginBinaryUninstallError{}),
		Entry("PluginCommandsConflictError", PluginCommandsConflictError{}),
		Entry("PluginInvalidError", PluginInvalidError{Err: errors.New("invalid error")}),
		Entry("PluginInvalidError", PluginInvalidError{}),
		Entry("PluginNotFoundError", PluginNotFoundError{}),
		Entry("PluginNotFoundInRepositoryError", PluginNotFoundInRepositoryError{}),
		Entry("PluginNotFoundOnDiskOrInAnyRepositoryError", PluginNotFoundOnDiskOrInAnyRepositoryError{}),
		Entry("PortNotAllowedWithHTTPDomainError", PortNotAllowedWithHTTPDomainError{}),
		Entry("ProcessInstanceNotFoundError", ProcessInstanceNotFoundError{ProcessType: "some-process", InstanceIndex: 1}),
		Entry("ProcessInstanceNotRunningError", ProcessInstanceNotRunningError{ProcessType: "some-process", InstanceIndex: 1}),
		Entry("PropertyCombinationError", PropertyCombinationError{Properties: []string{"property-1", "property-2"}}),
		Entry("RepositoryNameTakenError", RepositoryNameTakenError{}),
		Entry("RequiredArgumentError", RequiredArgumentError{}),
		Entry("RequiredFlagsError", RequiredFlagsError{}),
		Entry("RequiredNameForPushError", RequiredNameForPushError{}),
		Entry("RouteInDifferentSpaceError", RouteInDifferentSpaceError{}),
		Entry("RoutePathWithTCPDomainError", RoutePathWithTCPDomainError{}),
		Entry("RunTaskError", RunTaskError{}),
		Entry("SecurityGroupNotFoundError", SecurityGroupNotFoundError{}),
		Entry("ServiceInstanceNotFoundError", ServiceInstanceNotFoundError{}),
		Entry("SharedServiceInstanceNotFoundError", SharedServiceInstanceNotFoundError{}),
		Entry("SpaceNotFoundError", SpaceNotFoundError{}),
		Entry("SSHUnableToAuthenticateError", SSHUnableToAuthenticateError{}),
		Entry("SSLCertError", SSLCertError{}),
		Entry("StackNotFoundError with name", SpaceNotFoundError{Name: "steve"}),
		Entry("StackNotFoundError without name", SpaceNotFoundError{}),
		Entry("StagingFailedError", StagingFailedError{}),
		Entry("StagingFailedNoAppDetectedError", StagingFailedNoAppDetectedError{}),
		Entry("StagingTimeoutError", StagingTimeoutError{}),
		Entry("StartupTimeoutError", StartupTimeoutError{}),
		Entry("ThreeRequiredArgumentsError", ThreeRequiredArgumentsError{}),
		Entry("TriggerLegacyPushError", TriggerLegacyPushError{}),
		Entry("UnsuccessfulStartError", UnsuccessfulStartError{}),
		Entry("UnsupportedURLSchemeError", UnsupportedURLSchemeError{}),
		Entry("UploadFailedError", UploadFailedError{Err: JobFailedError{}}),
		Entry("V3APIDoesNotExistError", V3APIDoesNotExistError{}),
	)

	Describe("PluginInvalidError", func() {
		Context("when the wrapped error is nil", func() {
			It("does not concatenate the nil error in the returned Error()", func() {
				err := PluginInvalidError{}
				Expect(err.Error()).To(Equal("File is not a valid cf CLI plugin binary."))
			})
		})

		Context("when the wrapped error is not nil", func() {
			It("does prepends the error message in the returned Error()", func() {
				err := PluginInvalidError{Err: errors.New("ello")}
				Expect(err.Error()).To(Equal("ello\nFile is not a valid cf CLI plugin binary."))
			})
		})
	})
})
