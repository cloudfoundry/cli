package shared_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/command"
	. "code.cloudfoundry.org/cli/command/v2/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleError", func() {
	err := errors.New("some-error")

	DescribeTable("error translations",
		func(passedInErr error, expectedErr error) {
			actualErr := HandleError(passedInErr)
			Expect(actualErr).To(MatchError(expectedErr))
		},

		Entry("ccerror.RequestError -> APIRequestError",
			ccerror.RequestError{Err: err},
			command.APIRequestError{Err: err}),

		Entry("ccerror.UnverifiedServerError -> InvalidSSLCertError",
			ccerror.UnverifiedServerError{URL: "some-url"},
			command.InvalidSSLCertError{API: "some-url"}),

		Entry("ccerror.SSLValidationHostnameError -> SSLCertErrorError",
			ccerror.SSLValidationHostnameError{Message: "some-message"},
			command.SSLCertErrorError{Message: "some-message"}),

		Entry("ccerror.APINotFoundError -> APINotFoundError",
			ccerror.APINotFoundError{URL: "some-url"},
			command.APINotFoundError{URL: "some-url"}),

		Entry("v2action.ApplicationNotFoundError -> ApplicationNotFoundError",
			v2action.ApplicationNotFoundError{Name: "some-app"},
			command.ApplicationNotFoundError{Name: "some-app"}),

		Entry("v2action.SecurityGroupNotFoundError -> SecurityGroupNotFoundError",
			v2action.SecurityGroupNotFoundError{Name: "some-security-group"},
			SecurityGroupNotFoundError{Name: "some-security-group"}),

		Entry("v2action.ServiceInstanceNotFoundError -> ServiceInstanceNotFoundError",
			v2action.ServiceInstanceNotFoundError{Name: "some-service-instance"},
			command.ServiceInstanceNotFoundError{Name: "some-service-instance"}),

		Entry("ccerror.JobFailedError -> JobFailedError",
			ccerror.JobFailedError{JobGUID: "some-job-guid", Message: "some-message"},
			JobFailedError{JobGUID: "some-job-guid", Message: "some-message"}),

		Entry("ccerror.JobTimeoutError -> JobTimeoutError",
			ccerror.JobTimeoutError{JobGUID: "some-job-guid"},
			JobTimeoutError{JobGUID: "some-job-guid"}),

		Entry("v2action.OrganizationNotFoundError -> OrgNotFoundError",
			v2action.OrganizationNotFoundError{Name: "some-org"},
			OrganizationNotFoundError{Name: "some-org"}),

		Entry("v2action.SpaceNotFoundError -> SpaceNotFoundError",
			v2action.SpaceNotFoundError{Name: "some-space"},
			SpaceNotFoundError{Name: "some-space"}),

		Entry("sharedaction.NotLoggedInError -> NotLoggedInError",
			sharedaction.NotLoggedInError{BinaryName: "faceman"},
			command.NotLoggedInError{BinaryName: "faceman"}),

		Entry("sharedaction.NoOrganizationTargetedError -> NoOrganizationTargetedError",
			sharedaction.NoOrganizationTargetedError{BinaryName: "faceman"},
			command.NoOrganizationTargetedError{BinaryName: "faceman"}),

		Entry("sharedaction.NoSpaceTargetedError -> NoSpaceTargetedError",
			sharedaction.NoSpaceTargetedError{BinaryName: "faceman"},
			command.NoSpaceTargetedError{BinaryName: "faceman"}),

		Entry("v2action.HTTPHealthCheckInvalidError -> HTTPHealthCheckInvalidError",
			v2action.HTTPHealthCheckInvalidError{},
			HTTPHealthCheckInvalidError{},
		),

		Entry("v2action.RouteInDifferentSpaceError -> RouteInDifferentSpaceError",
			v2action.RouteInDifferentSpaceError{Route: "some-route"},
			RouteInDifferentSpaceError{Route: "some-route"},
		),

		Entry("v2action.FileChangedError -> FileChangedError",
			v2action.FileChangedError{Filename: "some-filename"},
			FileChangedError{Filename: "some-filename"},
		),

		Entry("v2action.EmptyDirectoryError -> EmptyDirectoryError",
			v2action.EmptyDirectoryError{Path: "some-filename"},
			EmptyDirectoryError{Path: "some-filename"},
		),

		Entry("uaa.BadCredentialsError -> command.BadCredentialsError",
			uaa.BadCredentialsError{},
			command.BadCredentialsError{},
		),

		Entry("uaa.InvalidAuthTokenError -> InvalidRefreshTokenError",
			uaa.InvalidAuthTokenError{},
			InvalidRefreshTokenError{},
		),

		Entry("pushaction.NoDomainsFoundError -> NoDomainsFoundError",
			pushaction.NoDomainsFoundError{OrganizationGUID: "some-guid"},
			NoDomainsFoundError{},
		),

		Entry("pushaction.UploadFailedError -> UploadFailedError",
			pushaction.UploadFailedError{Err: pushaction.NoDomainsFoundError{}},
			UploadFailedError{Err: NoDomainsFoundError{}},
		),

		Entry("default case -> original error",
			err,
			err),
	)

	It("returns nil for a nil error", func() {
		nilErr := HandleError(nil)
		Expect(nilErr).To(BeNil())
	})
})
