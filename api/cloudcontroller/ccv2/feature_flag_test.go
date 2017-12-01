package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Feature Flag", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetConfigFeatureFlags", func() {
		var (
			featureFlags []FeatureFlag
			warnings     Warnings
			err          error
		)

		JustBeforeEach(func() {
			featureFlags, warnings, err = client.GetConfigFeatureFlags()
		})

		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				response := `[
				  {
						"name": "feature-flag-1",
						"enabled": true
					},
				  {
						"name": "feature-flag-2",
						"enabled": false
					}
				]`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/config/feature_flags"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning"}}),
					))
			})

			It("returns the feature flags and all warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(featureFlags).To(Equal([]FeatureFlag{
					{Name: "feature-flag-1", Enabled: true},
					{Name: "feature-flag-2", Enabled: false},
				}))
				Expect(warnings).To(ConsistOf("warning"))
			})
		})

		Context("when an error is encountered", func() {
			BeforeEach(func() {
				response := `{
					"code": 10001,
					"description": "Some Error",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/config/feature_flags"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})
})
