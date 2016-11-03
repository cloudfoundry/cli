package common_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actors/v3actions"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/commands/v3/common"

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

		Entry("v3actions.ApplicationNotFoundError -> ApplicationNotFoundError", v3actions.ApplicationNotFoundError{
			Name: "some-app",
		}, ApplicationNotFoundError{
			Name: "some-app",
		}),

		Entry("default case -> original error", err, err),
	)
})
