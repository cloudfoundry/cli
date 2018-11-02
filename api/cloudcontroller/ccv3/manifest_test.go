package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Application Manifest", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("GetApplicationManifest", func() {
		var (
			appGUID string

			rawManifest []byte
			warnings    Warnings
			executeErr  error

			expectedYAML []byte
		)

		BeforeEach(func() {
			appGUID = "some-app-guid"
		})

		JustBeforeEach(func() {
			rawManifest, warnings, executeErr = client.GetApplicationManifest(appGUID)
		})

		When("getting the manifest is successful", func() {
			BeforeEach(func() {
				expectedYAML = []byte("---\n- banana")
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/manifest"),
						VerifyHeaderKV("Accept", "application/x-yaml"),
						RespondWith(
							http.StatusOK,
							expectedYAML,
							http.Header{
								"Content-Type":  {"application/x-yaml"},
								"X-Cf-Warnings": {"this is a warning"},
							}),
					),
				)
			})

			It("the manifest and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(rawManifest).To(Equal(expectedYAML))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := `{
  "errors": [
    {
      "code": 10008,
      "detail": "The request is semantically invalid: command presence",
      "title": "CF-UnprocessableEntity"
    },
    {
      "code": 10010,
      "detail": "App not found",
      "title": "CF-AppNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/manifest"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "App not found",
							Title:  "CF-AppNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
