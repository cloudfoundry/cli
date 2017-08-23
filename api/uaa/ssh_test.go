package uaa_test

import (
	"fmt"
	"net/http"

	. "code.cloudfoundry.org/cli/api/uaa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("SSH", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestUAAClientAndStore()
	})

	Describe("GetSSHPasscode", func() {
		Context("when no errors occur", func() {
			var expectedCode string

			BeforeEach(func() {
				expectedCode = "c0d3"
				locationHeader := http.Header{}
				locationHeader.Add("Location", fmt.Sprintf("http://localhost/redirect/cf?code=%s&state=", expectedCode))
				uaaServer.AppendHandlers(
					CombineHandlers(
						verifyRequestHost(TestUAAResource),
						VerifyRequest(http.MethodGet, "/oauth/authorize", "response_type=code&client_id=ssh-proxy"),
						RespondWith(http.StatusFound, nil, locationHeader),
					))
			})

			It("returns a ssh passcode", func() {
				code, err := client.GetSSHPasscode("4c3sst0k3n", "ssh-proxy")
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(expectedCode))
			})
		})

		Context("when an error occurs", func() {
			var response string

			BeforeEach(func() {
				response = `{
					"error": "some-error",
					"error_description": "some-description"
				}`
				uaaServer.AppendHandlers(
					CombineHandlers(
						verifyRequestHost(TestUAAResource),
						VerifyRequest(http.MethodGet, "/oauth/authorize", "response_type=code&client_id=ssh-proxy"),
						RespondWith(http.StatusBadRequest, response),
					))
			})

			It("returns an error", func() {
				_, err := client.GetSSHPasscode("4c3sst0k3n", "ssh-proxy")
				Expect(err).To(MatchError(RawHTTPStatusError{
					StatusCode:  http.StatusBadRequest,
					RawResponse: []byte(response),
				}))
			})
		})
	})
})
