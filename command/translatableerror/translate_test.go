package translatableerror_test

import (
	"bytes"
	"text/template"

	. "code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/ui"

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
			err, ok := e.(ui.TranslatableError)
			Expect(ok).To(BeTrue())
			err.Translate(translateFunc)
		},

		Entry("APINotFoundError", APINotFoundError{}),
		Entry("APIRequestError", APIRequestError{}),
		Entry("ApplicationNotFoundError", ApplicationNotFoundError{}),
		Entry("ArgumentCombinationError", ArgumentCombinationError{}),
		Entry("BadCredentialsError", BadCredentialsError{}),
		Entry("EmptyDirectoryError", EmptyDirectoryError{}),
		Entry("FileChangedError", FileChangedError{}),
		Entry("HealthCheckTypeUnsupportedError", HealthCheckTypeUnsupportedError{SupportedTypes: []string{"some-type", "another-type"}}),
		Entry("HTTPHealthCheckInvalidError", HTTPHealthCheckInvalidError{}),
		Entry("InvalidSSLCertError", InvalidSSLCertError{}),
		Entry("JobFailedError", JobFailedError{}),
		Entry("JobTimeoutError", JobTimeoutError{}),
		Entry("LifecycleMinimumAPIVersionNotMetError", LifecycleMinimumAPIVersionNotMetError{}),
		Entry("MinimumAPIVersionNotMetError", MinimumAPIVersionNotMetError{}),
		Entry("NoAPISetError", NoAPISetError{}),
		Entry("NoDomainsFoundError", NoDomainsFoundError{}),
		Entry("NoOrganizationTargetedError", NoOrganizationTargetedError{}),
		Entry("NoSpaceTargetedError", NoSpaceTargetedError{}),
		Entry("NotLoggedInError", NotLoggedInError{}),
		Entry("OrgNotFoundError", OrganizationNotFoundError{}),
		Entry("ParseArgumentError", ParseArgumentError{}),
		Entry("RequiredArgumentError", RequiredArgumentError{}),
		Entry("RouteInDifferentSpaceError", RouteInDifferentSpaceError{}),
		Entry("SecurityGroupNotFoundError", SecurityGroupNotFoundError{}),
		Entry("ServiceInstanceNotFoundError", ServiceInstanceNotFoundError{}),
		Entry("SpaceNotFoundError", SpaceNotFoundError{}),
		Entry("SSLCertErrorError", SSLCertErrorError{}),
		Entry("StagingFailedError", StagingFailedError{}),
		Entry("StagingFailedNoAppDetectedError", StagingFailedNoAppDetectedError{}),
		Entry("StagingTimeoutError", StagingTimeoutError{}),
		Entry("StartupTimeoutError", StartupTimeoutError{}),
		Entry("ThreeRequiredArgumentsError", ThreeRequiredArgumentsError{}),
		Entry("UnsuccessfulStartError", UnsuccessfulStartError{}),
		Entry("UnsupportedURLSchemeError", UnsupportedURLSchemeError{}),
		Entry("UploadFailedError", UploadFailedError{Err: JobFailedError{}}),
	)
})
