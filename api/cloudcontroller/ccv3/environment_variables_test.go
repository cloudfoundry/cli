package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("EnvironmentVariables", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("UpdateApplicationEnvironmentVariables", func() {
		var (
			envVars  EnvironmentVariables
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			envVars, warnings, err = client.UpdateApplicationEnvironmentVariables("some-app-guid", EnvironmentVariables{Variables: map[string]string{"my-var": "my-val"}})
		})

		Context("when the request errors", func() {
			BeforeEach(func() {
				expectedBody := map[string]interface{}{
					"var": map[string]string{
						"my-var": "my-val",
					},
				}
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/apps/some-app-guid/environment_variables"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusTeapot, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{ResponseCode: http.StatusTeapot}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		Context("when the request succeeds", func() {
			BeforeEach(func() {
				expectedBody := map[string]interface{}{
					"var": map[string]string{
						"my-var": "my-val",
					},
				}

				responseBody := `{
	"var": {
		"DEBUG": "false",
		"my-var": "my-val"
	}
}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/apps/some-app-guid/environment_variables"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, responseBody, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(envVars).To(Equal(EnvironmentVariables{Variables: map[string]string{
					"DEBUG":  "false",
					"my-var": "my-val",
				}}))
			})
		})
	})
})
