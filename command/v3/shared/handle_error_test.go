package shared_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/command"
	. "code.cloudfoundry.org/cli/command/v3/shared"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleError", func() {
	err := errors.New("some-error")

	unprocessableEntityError := cloudcontroller.UnprocessableEntityError{
		Message: "another message",
	}

	DescribeTable("error translations",
		func(passedInErr error, expectedErr error) {
			actualErr := HandleError(passedInErr)
			Expect(actualErr).To(MatchError(expectedErr))
		},

		Entry("cloudcontroller.RequestError -> APIRequestError", cloudcontroller.RequestError{
			Err: err,
		}, command.APIRequestError{
			Err: err,
		}),

		Entry("cloudcontroller.UnverifiedServerError -> InvalidSSLCertError", cloudcontroller.UnverifiedServerError{
			URL: "some-url",
		}, command.InvalidSSLCertError{
			API: "some-url",
		}),

		Entry("cloudcontroller.SSLValidationHostnameError -> SSLCertErrorError", cloudcontroller.SSLValidationHostnameError{
			Message: "some-message",
		}, command.SSLCertErrorError{
			Message: "some-message",
		}),

		Entry("cloudcontroller.UnprocessableEntityError with droplet message -> RunTaskError", cloudcontroller.UnprocessableEntityError{
			Message: "The request is semantically invalid: Task must have a droplet. Specify droplet or assign current droplet to app.",
		}, RunTaskError{
			Message: "App is not staged.",
		}),

		Entry("cloudcontroller.UnprocessableEntityError without droplet message -> original error", unprocessableEntityError, unprocessableEntityError),

		Entry("cloudcontroller.APINotFoundError -> APINotFoundError", cloudcontroller.APINotFoundError{
			URL: "some-url",
		}, command.APINotFoundError{
			URL: "some-url",
		}),

		Entry("v3action.ApplicationNotFoundError -> ApplicationNotFoundError", v3action.ApplicationNotFoundError{
			Name: "some-app",
		}, command.ApplicationNotFoundError{
			Name: "some-app",
		}),

		Entry("v3action.TaskWorkersUnavailableError -> RunTaskError", v3action.TaskWorkersUnavailableError{
			Message: "fooo: Banana Pants",
		}, RunTaskError{Message: "Task workers are unavailable."}),

		Entry("default case -> original error", err, err),
	)
})
