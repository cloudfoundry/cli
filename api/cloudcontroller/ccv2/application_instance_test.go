package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Application Instance", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetApplicationApplicationInstances", func() {
		var (
			appGUID string

			instances  map[int]ApplicationInstance
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			appGUID = "some-app-guid"
		})

		JustBeforeEach(func() {
			instances, warnings, executeErr = client.GetApplicationApplicationInstances(appGUID)
		})

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
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(instances).To(Equal(
					map[int]ApplicationInstance{
						0: {
							ID:      0,
							State:   constant.ApplicationInstanceRunning,
							Since:   1403140717.984577,
							Details: "some detail",
						},
						1: {
							ID:      1,
							State:   constant.ApplicationInstanceCrashed,
							Since:   2514251828.984577,
							Details: "more details",
						},
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
				Expect(executeErr).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The app could not be found: some-app-guid",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})
})
