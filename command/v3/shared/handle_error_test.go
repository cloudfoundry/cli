package shared_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/translatableerror"
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
			translatableerror.APIRequestError{Err: err}),

		Entry("ccerror.UnverifiedServerError -> InvalidSSLCertError",
			ccerror.UnverifiedServerError{URL: "some-url"},
			translatableerror.InvalidSSLCertError{API: "some-url"}),

		Entry("ccerror.SSLValidationHostnameError -> SSLCertErrorError",
			ccerror.SSLValidationHostnameError{Message: "some-message"},
			translatableerror.SSLCertError{Message: "some-message"}),

		Entry("ccerror.UnprocessableEntityError with droplet message -> RunTaskError",
			ccerror.UnprocessableEntityError{Message: "The request is semantically invalid: Task must have a droplet. Specify droplet or assign current droplet to app."},
			translatableerror.RunTaskError{Message: "App is not staged."}),

		// This changed in CF254
		Entry("ccerror.UnprocessableEntityError with droplet message -> RunTaskError",
			ccerror.UnprocessableEntityError{Message: "Task must have a droplet. Specify droplet or assign current droplet to app."},
			translatableerror.RunTaskError{Message: "App is not staged."}),

		Entry("ccerror.UnprocessableEntityError without droplet message -> original error",
			unprocessableEntityError,
			unprocessableEntityError),

		Entry("ccerror.APINotFoundError -> APINotFoundError",
			ccerror.APINotFoundError{URL: "some-url"},
			translatableerror.APINotFoundError{URL: "some-url"}),

		Entry("v3action.ApplicationNotFoundError -> ApplicationNotFoundError",
			v3action.ApplicationNotFoundError{Name: "some-app"},
			translatableerror.ApplicationNotFoundError{Name: "some-app"}),

		Entry("v3action.TaskWorkersUnavailableError -> RunTaskError",
			v3action.TaskWorkersUnavailableError{Message: "fooo: Banana Pants"},
			translatableerror.RunTaskError{Message: "Task workers are unavailable."}),

		Entry("sharedaction.NotLoggedInError -> NotLoggedInError",
			sharedaction.NotLoggedInError{BinaryName: "faceman"},
			translatableerror.NotLoggedInError{BinaryName: "faceman"}),

		Entry("sharedaction.NoOrganizationTargetedError -> NoOrganizationTargetedError",
			sharedaction.NoOrganizationTargetedError{BinaryName: "faceman"},
			translatableerror.NoOrganizationTargetedError{BinaryName: "faceman"}),

		Entry("sharedaction.NoSpaceTargetedError -> NoSpaceTargetedError",
			sharedaction.NoSpaceTargetedError{BinaryName: "faceman"},
			translatableerror.NoSpaceTargetedError{BinaryName: "faceman"}),

		Entry("v3action.AssignDropletError -> AssignDropletError",
			v3action.AssignDropletError{Message: "some-message"},
			translatableerror.AssignDropletError{Message: "some-message"}),

		Entry("v3action.OrganizationNotFoundError -> OrgNotFoundError",
			v3action.OrganizationNotFoundError{Name: "some-org"},
			translatableerror.OrganizationNotFoundError{Name: "some-org"}),

		Entry("v3action.ProcessNotFoundError -> ProcessNotFoundError",
			v3action.ProcessNotFoundError{ProcessType: "some-process-type"},
			translatableerror.ProcessNotFoundError{ProcessType: "some-process-type"}),

		Entry("v3action.ProcessInstanceNotFoundError -> ProcessInstanceNotFoundError",
			v3action.ProcessInstanceNotFoundError{ProcessType: "some-process-type", InstanceIndex: 42},
			translatableerror.ProcessInstanceNotFoundError{ProcessType: "some-process-type", InstanceIndex: 42}),

		Entry("v3action.StagingTimeoutError -> StagingTimeoutError",
			v3action.StagingTimeoutError{AppName: "some-app", Timeout: time.Nanosecond},
			translatableerror.StagingTimeoutError{AppName: "some-app", Timeout: time.Nanosecond}),

		Entry("v3action.EmptyDirectoryError -> EmptyDirectoryError",
			v3action.EmptyDirectoryError{Path: "some-path"},
			translatableerror.EmptyDirectoryError{Path: "some-path"}),

		Entry("default case -> original error",
			err,
			err),
	)
})
