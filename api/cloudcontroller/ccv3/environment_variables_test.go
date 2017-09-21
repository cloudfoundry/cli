package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("EnvironmentVariables", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetApplicationEnvironmentVariables", func() {
		var (
			fetchedEnvVars EnvironmentVariables
			warnings       Warnings
			err            error
		)

		JustBeforeEach(func() {
			fetchedEnvVars, warnings, err = client.GetApplicationEnvironmentVariables("some-app-guid")
		})

		Context("when the request errors", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/environment_variables"),
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
				responseBody := `{
	"var": {
		"DEBUG": "false"
	}
}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/environment_variables"),
						RespondWith(http.StatusOK, responseBody, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(fetchedEnvVars).To(Equal(EnvironmentVariables{Variables: map[string]types.FilteredString{
					"DEBUG": {Value: "false", IsSet: true},
				}}))
			})

		})
	})

	Describe("PatchApplicationEnvironmentVariables", func() {
		var (
			envVars        EnvironmentVariables
			patchedEnvVars EnvironmentVariables

			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			patchedEnvVars, warnings, err = client.PatchApplicationEnvironmentVariables("some-app-guid", envVars)
		})

		Context("when the request errors", func() {
			BeforeEach(func() {
				envVars = EnvironmentVariables{Variables: map[string]types.FilteredString{"my-var": {Value: "my-val", IsSet: true}}}
				expectedBody := map[string]interface{}{
					"var": map[string]types.FilteredString{
						"my-var": {Value: "my-val", IsSet: true},
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
			Context("when env variable is being set", func() {
				BeforeEach(func() {
					envVars = EnvironmentVariables{Variables: map[string]types.FilteredString{"my-var": {Value: "my-val", IsSet: true}}}
					expectedBody := map[string]interface{}{
						"var": map[string]types.FilteredString{
							"my-var": {Value: "my-val", IsSet: true},
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
					Expect(patchedEnvVars).To(Equal(EnvironmentVariables{Variables: map[string]types.FilteredString{
						"DEBUG":  {Value: "false", IsSet: true},
						"my-var": {Value: "my-val", IsSet: true},
					}}))
				})
			})

			Context("when env variable is being unset", func() {
				BeforeEach(func() {
					envVars = EnvironmentVariables{Variables: map[string]types.FilteredString{"my-var": {Value: "", IsSet: false}}}
					expectedBody := map[string]interface{}{
						"var": map[string]types.FilteredString{
							"my-var": {Value: "", IsSet: false},
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
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
					Expect(patchedEnvVars).To(Equal(EnvironmentVariables{Variables: map[string]types.FilteredString{
						"DEBUG": {Value: "false", IsSet: true},
					}}))
				})
			})
		})
	})
})
