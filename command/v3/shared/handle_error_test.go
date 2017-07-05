package shared_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command"
	. "code.cloudfoundry.org/cli/command/v3/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleError", func() {
	err := errors.New("some-error")
	unprocessableEntityError := ccerror.UnprocessableEntityError{Message: "another message"}

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

		Entry("ccerror.APINotFoundError -> APINotFoundError",
			ccerror.APINotFoundError{URL: "some-url"},
			command.APINotFoundError{URL: "some-url"}),

		Entry("v3action.ApplicationNotFoundError -> ApplicationNotFoundError",
			v3action.ApplicationNotFoundError{Name: "some-app"},
			command.ApplicationNotFoundError{Name: "some-app"}),

		Entry("v3action.TaskWorkersUnavailableError -> RunTaskError",
			v3action.TaskWorkersUnavailableError{Message: "fooo: Banana Pants"},
			RunTaskError{Message: "Task workers are unavailable."}),

		Entry("sharedaction.NotLoggedInError -> NotLoggedInError",
			sharedaction.NotLoggedInError{BinaryName: "faceman"},
			command.NotLoggedInError{BinaryName: "faceman"}),

		Entry("sharedaction.NoOrganizationTargetedError -> NoOrganizationTargetedError",
			sharedaction.NoOrganizationTargetedError{BinaryName: "faceman"},
			command.NoOrganizationTargetedError{BinaryName: "faceman"}),

		Entry("sharedaction.NoSpaceTargetedError -> NoSpaceTargetedError",
			sharedaction.NoSpaceTargetedError{BinaryName: "faceman"},
			command.NoSpaceTargetedError{BinaryName: "faceman"}),

		Entry("v3action.OrganizationNotFoundError -> OrgNotFoundError",
			v3action.OrganizationNotFoundError{Name: "some-org"},
			OrganizationNotFoundError{Name: "some-org"}),

		Entry("v3action.AssignDropletError -> AssignDropletError",
			v3action.AssignDropletError{Message: "some-message"},
			AssignDropletError{Message: "some-message"}),

		Entry("default case -> original error",
			err,
			err),
	)
})
