package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Application", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("UpdateAppFeature", func() {
		var (
			warnings   Warnings
			executeErr error
			appGUID    = "some-app-guid"
			enabled    = false
		)
		Context("Updating SSH", func() {
			JustBeforeEach(func() {
				warnings, executeErr = client.UpdateAppFeature(appGUID, enabled, "ssh")
			})

			When("the app exists", func() {
				BeforeEach(func() {
					response1 := fmt.Sprintf(`{
   "name": "ssh",
   "description": "Enable SSHing into the app.",
   "enabled": false
}`)

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, fmt.Sprintf("/v3/apps/%s/features/ssh", appGUID)),
							VerifyBody([]byte(fmt.Sprintf(`{"enabled":%t}`, enabled))),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns all warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})

			When("the cloud controller returns errors and warnings", func() {
				BeforeEach(func() {
					updateResponse := `{
		 "errors": [
		   {
		     "code": 10008,
		     "detail": "The request is semantically invalid: command presence",
		     "title": "CF-UnprocessableEntity"
		   },
		   {
		     "code": 10010,
		     "detail": "Org not found",
		     "title": "CF-ResourceNotFound"
		   }
		 ]
		}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, fmt.Sprintf("/v3/apps/%s/features/ssh", appGUID)),
							RespondWith(http.StatusTeapot, updateResponse, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
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
								Detail: "Org not found",
								Title:  "CF-ResourceNotFound",
							},
						},
					}))
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})
		})
	})

	Describe("GetAppFeature", func() {
		var (
			warnings           Warnings
			executeErr         error
			appGUID            = "some-app-guid"
			applicationFeature ApplicationFeature
		)

		Context("Getting SSH", func() {
			JustBeforeEach(func() {
				applicationFeature, warnings, executeErr = client.GetAppFeature(appGUID, "ssh")
			})

			When("the app exists", func() {
				BeforeEach(func() {
					getResponse := fmt.Sprintf(`{
   "name": "ssh",
   "description": "Enable SSHing into the app.",
   "enabled": false
}`)

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, fmt.Sprintf("/v3/apps/%s/features/ssh", appGUID)),
							RespondWith(http.StatusOK, getResponse, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns an app feature and all warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(warnings).To(ConsistOf("this is a warning"))
					Expect(applicationFeature.Name).To(Equal("ssh"))
					Expect(applicationFeature.Enabled).To(Equal(false))
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
		    "detail": "Org not found",
		    "title": "CF-ResourceNotFound"
		  }
		]
		}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, fmt.Sprintf("/v3/apps/%s/features/ssh", appGUID)),
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
								Detail: "Org not found",
								Title:  "CF-ResourceNotFound",
							},
						},
					}))
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})
		})
	})
})
