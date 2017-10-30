package shared_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
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

		Entry("actionerror.ApplicationNotStartedError -> ApplicationNotStartedError",
			actionerror.ApplicationNotStartedError{Name: "some-app"},
			translatableerror.ApplicationNotStartedError{Name: "some-app"}),

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

		Entry("actionerror.ApplicationNotFoundError -> ApplicationNotFoundError",
			actionerror.ApplicationNotFoundError{Name: "some-app"},
			translatableerror.ApplicationNotFoundError{Name: "some-app"}),

		Entry("actionerror.TaskWorkersUnavailableError -> RunTaskError",
			actionerror.TaskWorkersUnavailableError{Message: "fooo: Banana Pants"},
			translatableerror.RunTaskError{Message: "Task workers are unavailable."}),

		Entry("actionerror.NotLoggedInError -> NotLoggedInError",
			actionerror.NotLoggedInError{BinaryName: "faceman"},
			translatableerror.NotLoggedInError{BinaryName: "faceman"}),

		Entry("actionerror.NoOrganizationTargetedError -> NoOrganizationTargetedError",
			actionerror.NoOrganizationTargetedError{BinaryName: "faceman"},
			translatableerror.NoOrganizationTargetedError{BinaryName: "faceman"}),

		Entry("actionerror.NoSpaceTargetedError -> NoSpaceTargetedError",
			actionerror.NoSpaceTargetedError{BinaryName: "faceman"},
			translatableerror.NoSpaceTargetedError{BinaryName: "faceman"}),

		Entry("actionerror.AssignDropletError -> AssignDropletError",
			actionerror.AssignDropletError{Message: "some-message"},
			translatableerror.AssignDropletError{Message: "some-message"}),

		Entry("actionerror.OrganizationNotFoundError -> OrgNotFoundError",
			actionerror.OrganizationNotFoundError{Name: "some-org"},
			translatableerror.OrganizationNotFoundError{Name: "some-org"}),

		Entry("actionerror.ProcessNotFoundError -> ProcessNotFoundError",
			actionerror.ProcessNotFoundError{ProcessType: "some-process-type"},
			translatableerror.ProcessNotFoundError{ProcessType: "some-process-type"}),

		Entry("actionerror.ProcessInstanceNotFoundError -> ProcessInstanceNotFoundError",
			actionerror.ProcessInstanceNotFoundError{ProcessType: "some-process-type", InstanceIndex: 42},
			translatableerror.ProcessInstanceNotFoundError{ProcessType: "some-process-type", InstanceIndex: 42}),

		Entry("actionerror.ProcessInstanceNotRunningError -> ProcessInstanceNotRunningError",
			actionerror.ProcessInstanceNotRunningError{ProcessType: "some-process-type", InstanceIndex: 42},
			translatableerror.ProcessInstanceNotRunningError{ProcessType: "some-process-type", InstanceIndex: 42}),

		Entry("v3action.StagingTimeoutError -> StagingTimeoutError",
			actionerror.StagingTimeoutError{AppName: "some-app", Timeout: time.Nanosecond},
			translatableerror.StagingTimeoutError{AppName: "some-app", Timeout: time.Nanosecond}),

		Entry("actionerror.EmptyDirectoryError -> EmptyDirectoryError",
			actionerror.EmptyDirectoryError{Path: "some-path"},
			translatableerror.EmptyDirectoryError{Path: "some-path"}),

		Entry("default case -> original error",
			err,
			err),
	)
})
