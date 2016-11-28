package common_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actors/v3actions"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/command/v3/common"

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
		}, APIRequestError{
			Err: err,
		}),

		Entry("cloudcontroller.UnverifiedServerError -> InvalidSSLCertError", cloudcontroller.UnverifiedServerError{
			URL: "some-url",
		}, InvalidSSLCertError{
			API: "some-url",
		}),

		Entry("cloudcontroller.UnprocessableEntityError with droplet message -> RunTaskError", cloudcontroller.UnprocessableEntityError{
			Message: "The request is semantically invalid: Task must have a droplet. Specify droplet or assign current droplet to app.",
		}, RunTaskError{
			Message: "App is not staged.",
		}),

		Entry("cloudcontroller.UnprocessableEntityError without droplet message -> original error", unprocessableEntityError, unprocessableEntityError),

		Entry("cloudcontroller.APINotFoundError -> APINotFoundError", cloudcontroller.APINotFoundError{
			URL: "some-url",
		}, APINotFoundError{
			URL: "some-url",
		}),

		Entry("v3actions.ApplicationNotFoundError -> ApplicationNotFoundError", v3actions.ApplicationNotFoundError{
			Name: "some-app",
		}, ApplicationNotFoundError{
			Name: "some-app",
		}),

		Entry("v3actions.TaskWorkersUnavailableError -> RunTaskError", v3actions.TaskWorkersUnavailableError{
			Message: "fooo: Banana Pants",
		}, RunTaskError{Message: "Task workers are unavailable."}),

		Entry("default case -> original error", err, err),
	)
})
