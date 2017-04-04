package ccerror_test

import (
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("V2UnexpectedResponseError", func() {
	It("formats the error", func() {
		err := V2UnexpectedResponseError{
			ResponseCode: 123,
			V2ErrorResponse: V2ErrorResponse{
				Code:        456,
				Description: "some-error-description",
				ErrorCode:   "some-error-code",
			},
			RequestIDs: []string{
				"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95",
				"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95::7445d9db-c31e-410d-8dc5-9f79ec3fc26f",
			},
		}
		Expect(err.Error()).To(Equal(`Unexpected Response
Response code: 123
CC code:       456
CC error code: some-error-code
Request ID:    6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95
Request ID:    6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95::7445d9db-c31e-410d-8dc5-9f79ec3fc26f
Description:   some-error-description`))
	})
})
