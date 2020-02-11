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

var _ = Describe("Space Features", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("UpdateSpaceFeature", func() {
		var (
			warnings   Warnings
			executeErr error
			spaceGUID  = "some-space-guid"
			enabled    = false
		)

		Context("Updating SSH", func() {
			JustBeforeEach(func() {
				warnings, executeErr = client.UpdateSpaceFeature(spaceGUID, enabled, "ssh")
			})

			When("the space exists", func() {
				BeforeEach(func() {
					response := fmt.Sprintf(`{
   "name": "ssh",
   "description": "Enable SSHing into apps in the space.",
   "enabled": false
}`)

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, fmt.Sprintf("/v3/spaces/%s/features/ssh", spaceGUID)),
							VerifyBody([]byte(fmt.Sprintf(`{"enabled":%t}`, enabled))),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("updates the space feature to the desired value", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})

			When("the cloud controller returns errors and warnings", func() {
				BeforeEach(func() {
					updateResponse := `{
		 "errors": [
		   {
		     "code": 10010,
		     "detail": "Org not found",
		     "title": "CF-ResourceNotFound"
		   }
		 ]
		}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, fmt.Sprintf("/v3/spaces/%s/features/ssh", spaceGUID)),
							RespondWith(http.StatusTeapot, updateResponse, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(MatchError(ccerror.V3UnexpectedResponseError{
						ResponseCode: http.StatusTeapot,
						V3ErrorResponse: ccerror.V3ErrorResponse{
							Errors: []ccerror.V3Error{
								{
									Code:   10010,
									Detail: "Org not found",
									Title:  "CF-ResourceNotFound",
								},
							},
						},
					}))
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})
		})
	})
})
