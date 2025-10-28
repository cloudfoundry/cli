package ccv3_test

import (
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Info", func() {
	var (
		client      *Client
		respondWith http.HandlerFunc
		info        Info
		warnings    Warnings
		executeErr  error
	)
	BeforeEach(func() {
		respondWith = nil
	})

	JustBeforeEach(func() {
		client, _ = NewTestClient()

		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/v3/info"),
				respondWith,
			),
		)

		info, warnings, executeErr = client.GetInfo()
	})

	Describe("when all requests are successful", func() {
		BeforeEach(func() {
			response := strings.Replace(`{
				"name": "test-name",
				"build": "test-build",
				"osbapi_version": "1.0"
			}`, "SERVER_URL", server.URL(), -1)

			respondWith = RespondWith(
				http.StatusOK,
				response,
				http.Header{"X-Cf-Warnings": {"warning 1"}})
		})

		It("returns the CC Information", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(info.Name).To(Equal("test-name"))
			Expect(info.Build).To(Equal("test-build"))
			Expect(info.OSBAPIVersion).To(Equal("1.0"))
		})

		It("returns all warnings", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf("warning 1"))
		})
	})

	When("the cloud controller encounters an error", func() {
		When("the info response is invalid", func() {
			BeforeEach(func() {
				respondWith = RespondWith(
					http.StatusNotFound,
					`i am google, bow down`,
					http.Header{"X-Cf-Warnings": {"warning 2"}},
				)
			})

			It("returns an APINotFoundError and no warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.APINotFoundError{URL: server.URL()}))
				Expect(warnings).To(BeNil())
			})
		})

		When("the error occurs making a request to '/info'", func() {
			BeforeEach(func() {
				respondWith = RespondWith(
					http.StatusNotFound,
					`{"errors": [{}]}`,
					http.Header{"X-Cf-Warnings": {"this is a warning"}})
			})

			It("returns the same error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.ResourceNotFoundError{}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
