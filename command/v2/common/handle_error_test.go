package common_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/command/v2/common"

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

		Entry("cloudcontroller.APINotFoundError -> APINotFoundError", cloudcontroller.APINotFoundError{
			URL: "some-url",
		}, APINotFoundError{
			URL: "some-url",
		}),

		Entry("v2action.ApplicationNotFoundError -> ApplicationNotFoundError", v2action.ApplicationNotFoundError{
			Name: "some-app",
		}, ApplicationNotFoundError{
			Name: "some-app",
		}),

		Entry("v2action.ServiceInstanceNotFoundError -> ServiceInstanceNotFoundError", v2action.ServiceInstanceNotFoundError{
			Name: "some-service-instance",
		}, ServiceInstanceNotFoundError{
			Name: "some-service-instance",
		}),

		Entry("default case -> original error", err, err),
	)
})
