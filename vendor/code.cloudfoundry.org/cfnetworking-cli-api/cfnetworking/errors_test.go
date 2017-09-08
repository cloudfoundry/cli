package cfnetworking_test

import (
	"fmt"
	"net/http"

	. "code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking"
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetworkingfakes"
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/networkerror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error Wrapper", func() {
	const errorMessage = "I am an error"

	DescribeTable("Make",
		func(statusCode int, expectedError error) {
			fakeConnection := new(cfnetworkingfakes.FakeConnection)
			fakeConnection.MakeReturns(networkerror.RawHTTPStatusError{
				StatusCode:  statusCode,
				RawResponse: []byte(fmt.Sprintf(`{"error":"%s"}`, errorMessage)),
			})

			errorWrapper := NewErrorWrapper().Wrap(fakeConnection)
			err := errorWrapper.Make(nil, nil)
			Expect(err).To(MatchError(expectedError))
		},
		Entry("400 -> BadRequestError", http.StatusBadRequest, networkerror.BadRequestError{Message: errorMessage}),
		Entry("401 -> UnauthorizedError", http.StatusUnauthorized, networkerror.UnauthorizedError{Message: errorMessage}),
		Entry("403 -> ForbiddenError", http.StatusForbidden, networkerror.ForbiddenError{Message: errorMessage}),
		Entry("406 -> NotAcceptable", http.StatusNotAcceptable, networkerror.NotAcceptableError{Message: errorMessage}),
		Entry("409 -> ConflictError", http.StatusConflict, networkerror.ConflictError{Message: errorMessage}),
	)
})
