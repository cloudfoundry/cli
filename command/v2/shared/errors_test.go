package shared_test

import (
	"bytes"
	"text/template"

	. "code.cloudfoundry.org/cli/command/v2/shared"
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

		// Actor errors.
		Entry("FileChangedError", FileChangedError{}),
		Entry("HTTPHealthCheckInvalidError", HTTPHealthCheckInvalidError{}),
		Entry("JobFailedError", JobFailedError{}),
		Entry("JobTimeoutError", JobTimeoutError{}),
		Entry("NoDomainsFoundError", NoDomainsFoundError{}),
		Entry("RouteInDifferentSpaceError", RouteInDifferentSpaceError{}),
		Entry("StagingFailedError", StagingFailedError{}),
		Entry("StagingFailedNoAppDetectedError", StagingFailedNoAppDetectedError{}),
		Entry("StagingTimeoutError", StagingTimeoutError{}),
		Entry("StartupTimeoutError", StartupTimeoutError{}),
		Entry("UploadFailedError", UploadFailedError{Err: JobFailedError{}}),

		// Command errors.
		Entry("NoOrgTargetedError", NoOrganizationTargetedError{}),
		Entry("OrgNotFoundError", OrganizationNotFoundError{}),
		Entry("SecurityGroupNotBoundError", SecurityGroupNotBoundError{}),
		Entry("SecurityGroupNotFoundError", SecurityGroupNotFoundError{}),
		Entry("SpaceNotFoundError", SpaceNotFoundError{}),
		Entry("UnsuccessfulStartError", UnsuccessfulStartError{}),
	)
})
