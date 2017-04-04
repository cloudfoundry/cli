package ccerror_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("V3UnexpectedResponseError", func() {
	Describe("Error", func() {
		It("returns all of the errors joined with newlines", func() {
			err := V3UnexpectedResponseError{
				ResponseCode: http.StatusTeapot,
				V3ErrorResponse: V3ErrorResponse{
					Errors: []V3Error{
						{
							Code:   282010,
							Detail: "detail 1",
							Title:  "title-1",
						},
						{
							Code:   10242013,
							Detail: "detail 2",
							Title:  "title-2",
						},
					},
				},
				RequestIDs: []string{
					"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95",
					"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95::7445d9db-c31e-410d-8dc5-9f79ec3fc26f",
				},
			}

			Expect(err.Error()).To(Equal(`Unexpected Response
Response Code: 418
Request ID:    6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95
Request ID:    6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95::7445d9db-c31e-410d-8dc5-9f79ec3fc26f
Code: 282010, Title: title-1, Detail: detail 1
Code: 10242013, Title: title-2, Detail: detail 2`))
		})
	})
})
