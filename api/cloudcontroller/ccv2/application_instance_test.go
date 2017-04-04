package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Application Instance", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetApplicationInstancesByApplication", func() {
		Context("when the app is found", func() {
			BeforeEach(func() {

				response := `{
					"0": {
						"state": "RUNNING",
						"since": 1403140717.984577,
						"details": "some detail"
					},
					"1": {
						"state": "CRASHED",
						"since": 2514251828.984577,
						"details": "more details"
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/apps/some-app-guid/instances"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the app instances and warnings", func() {
				instances, warnings, err := client.GetApplicationInstancesByApplication("some-app-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(instances).To(HaveLen(2))

				Expect(instances[0]).To(Equal(ApplicationInstance{
					ID:      0,
					State:   ApplicationInstanceRunning,
					Since:   1403140717.984577,
					Details: "some detail",
				},
				))

				Expect(instances[1]).To(Equal(ApplicationInstance{
					ID:      1,
					State:   ApplicationInstanceCrashed,
					Since:   2514251828.984577,
					Details: "more details",
				},
				))
			})
		})

		Context("when the client returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 100004,
					"description": "The app could not be found: some-app-guid",
					"error_code": "CF-AppNotFound"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/apps/some-app-guid/instances"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := client.GetApplicationInstancesByApplication("some-app-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The app could not be found: some-app-guid",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})
})
