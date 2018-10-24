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
		client, _ = NewTestClient()
	})

	Describe("UpdateApplicationEnvironmentVariables", func() {
		var (
			envVars        EnvironmentVariables
			patchedEnvVars EnvironmentVariables

			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			patchedEnvVars, warnings, executeErr = client.UpdateApplicationEnvironmentVariables("some-app-guid", envVars)
		})

		When("the request errors", func() {
			BeforeEach(func() {
				envVars = EnvironmentVariables{"my-var": {Value: "my-val", IsSet: true}}

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
				Expect(executeErr).To(MatchError(ccerror.V3UnexpectedResponseError{ResponseCode: http.StatusTeapot}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("the request succeeds", func() {
			When("env variable is being set", func() {
				BeforeEach(func() {
					envVars = EnvironmentVariables{
						"my-var":    {Value: "my-val", IsSet: true},
						"delete-me": {},
					}

					expectedBody := map[string]interface{}{
						"var": map[string]interface{}{
							"my-var":    "my-val",
							"delete-me": nil,
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
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
					Expect(patchedEnvVars).To(Equal(EnvironmentVariables{
						"DEBUG":  {Value: "false", IsSet: true},
						"my-var": {Value: "my-val", IsSet: true},
					}))
				})
			})

			When("env variable is being unset", func() {
				BeforeEach(func() {
					envVars = EnvironmentVariables{
						"my-var": {Value: "", IsSet: false},
					}

					expectedBody := map[string]interface{}{
						"var": map[string]interface{}{
							"my-var": nil,
						},
					}

					responseBody := `{
						"var": {
							"DEBUG": "false"
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
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
					Expect(patchedEnvVars).To(Equal(EnvironmentVariables{
						"DEBUG": {Value: "false", IsSet: true},
					}))
				})
			})
		})
	})
})
