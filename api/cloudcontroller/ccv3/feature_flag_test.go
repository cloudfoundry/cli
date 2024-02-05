package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Feature Flags", func() {
	var client *Client
	var executeErr error
	var warnings Warnings
	var featureFlags []resources.FeatureFlag

	BeforeEach(func() {
		client, _ = NewTestClient()
	})
	Describe("GetFeatureFlags", func() {

		JustBeforeEach(func() {
			featureFlags, warnings, executeErr = client.GetFeatureFlags()
		})

		When("the cloud controller returns without errors", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
					"pagination": {
						"next": {
							"href": "%s/v3/feature_flags?page=2&per_page=2"
						}
					},
					"resources": [
						{
						  "name": "flag1",
							"enabled": true,
							"updated_at": "2016-10-17T20:00:42Z",
							"custom_error_message": null,
							"links": {
								"self": {
									"href": "https://api.example.org/v3/feature_flags/flag1"
								}
							}
						},
						{
						  "name": "flag2",
							"enabled": false,
							"updated_at": "2016-10-17T20:00:42Z",
							"custom_error_message": null,
							"links": {
								"self": {
									"href": "https://api.example.org/v3/feature_flags/flag2"
								}
							}
						}
					]
				}`, server.URL())
				response2 := `{
					"pagination": {
						"next": null
					},
					"resources": [
						{
						  "name": "flag3",
							"enabled": true,
							"updated_at": "2016-10-17T20:00:42Z",
							"custom_error_message": "error message the user sees",
							"links": {
								"self": {
									"href": "https://api.example.org/v3/feature_flags/flag3"
								}
							}
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/feature_flags"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"page1 warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/feature_flags", "page=2&per_page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"page2 warning"}}),
					),
				)
			})

			It("returns feature flags and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("page1 warning", "page2 warning"))
				Expect(featureFlags).To(ConsistOf(
					resources.FeatureFlag{
						Name:    "flag1",
						Enabled: true,
					},
					resources.FeatureFlag{
						Name:    "flag2",
						Enabled: false,
					},
					resources.FeatureFlag{
						Name:    "flag3",
						Enabled: true,
					},
				))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				errResponse := `{
		   "errors": [
		      {
		         "detail": "Feature flag not found",
		         "title": "CF-ResourceNotFound",
		         "code": 10010
		      }
		   ]
		}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/feature_flags"),
						RespondWith(http.StatusNotFound, errResponse, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns a flag not found error", func() {
				Expect(executeErr).To(MatchError(ccerror.FeatureFlagNotFoundError{}))
				Expect(warnings).To(Equal(Warnings{"this is a warning"}))
			})

		})

	})

	Describe("GetFeatureFlag", func() {
		var (
			flagName     string
			flag         resources.FeatureFlag
			warnings     Warnings
			executeError error
		)

		JustBeforeEach(func() {
			flag, warnings, executeError = client.GetFeatureFlag(flagName)
		})

		When("The flag exists", func() {
			BeforeEach(func() {
				flagName = "flag1"
				response := `{
  "updated_at": "2016-06-08T16:41:39Z",
  "name": "flag1",
  "enabled": true,
  "custom_error_message": "This is wonky",
  "links": {
    "self": { "href": "https://example.com/v3/config/feature_flags/flag1" }
  }
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, fmt.Sprintf("/v3/feature_flags/%s", flagName)),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the requested flag", func() {
				Expect(executeError).ToNot(HaveOccurred())
				Expect(warnings).To(Equal(Warnings{"this is a warning"}))
				Expect(flag.Name).To(Equal(flagName))
				Expect(flag.Enabled).To(Equal(true))
			})
		})
		When("The flag does not exist", func() {
			BeforeEach(func() {
				flagName = "flag1"
				response := `{
		   "errors": [
		      {
		         "detail": "Feature flag not found",
		         "title": "CF-ResourceNotFound",
		         "code": 10010
		      }
		   ]
		}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, fmt.Sprintf("/v3/feature_flags/%s", flagName)),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns a flag not found error", func() {
				Expect(executeError).To(MatchError(ccerror.FeatureFlagNotFoundError{}))
				Expect(warnings).To(Equal(Warnings{"this is a warning"}))
			})
		})
		When("some other error occurs", func() {
			BeforeEach(func() {
				flagName = "flag1"
				response := `{
		   "errors": [
		     {
		        "code": 10008,
		        "detail": "The request is semantically invalid: command presence",
		        "title": "CF-UnprocessableEntity"
		      }
		   ]
		}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, fmt.Sprintf("/v3/feature_flags/%s", flagName)),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error", func() {
				Expect(executeError).To(MatchError(ccerror.V3UnexpectedResponseError{
					V3ErrorResponse: ccerror.V3ErrorResponse{
						Errors: []ccerror.V3Error{
							{
								Code:   10008,
								Detail: "The request is semantically invalid: command presence",
								Title:  "CF-UnprocessableEntity",
							},
						},
					},
					ResponseCode: http.StatusTeapot,
				},
				))
				Expect(warnings).To(Equal(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("UpdateFeatureFlag", func() {
		var (
			argFlag      resources.FeatureFlag
			expectedBody string
			ccFlag       resources.FeatureFlag
			warnings     Warnings
			executeError error
		)

		BeforeEach(func() {
			expectedBody = `{"enabled":false}`
			argFlag = resources.FeatureFlag{
				Name:    "flag1",
				Enabled: false,
			}
		})

		JustBeforeEach(func() {
			ccFlag, warnings, executeError = client.UpdateFeatureFlag(argFlag)
		})

		When("The flag exists", func() {
			BeforeEach(func() {
				response := `{
  "updated_at": "2016-06-08T16:41:39Z",
  "name": "flag1",
  "enabled": false,
  "custom_error_message": "This is wonky",
  "links": {
    "self": { "href": "https://example.com/v3/config/feature_flags/flag1" }
  }
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, fmt.Sprintf("/v3/feature_flags/%s", argFlag.Name)),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("Successfully updates the flag", func() {
				Expect(executeError).ToNot(HaveOccurred())
				Expect(warnings).To(Equal(Warnings{"this is a warning"}))
				Expect(ccFlag).To(Equal(argFlag))
			})
		})

		When("The flag does not exist", func() {
			BeforeEach(func() {
				response := `{
		   "errors": [
		      {
		         "detail": "Feature flag not found",
		         "title": "CF-ResourceNotFound",
		         "code": 10010
		      }
		   ]
		}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, fmt.Sprintf("/v3/feature_flags/%s", argFlag.Name)),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns a flag not found error", func() {
				Expect(executeError).To(MatchError(ccerror.FeatureFlagNotFoundError{}))
				Expect(warnings).To(Equal(Warnings{"this is a warning"}))
			})
		})
		When("some other error occurs", func() {
			BeforeEach(func() {
				response := `{
		   "errors": [
		     {
		        "code": 10008,
		        "detail": "The request is semantically invalid: command presence",
		        "title": "CF-UnprocessableEntity"
		      }
		   ]
		}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, fmt.Sprintf("/v3/feature_flags/%s", argFlag.Name)),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error", func() {
				Expect(executeError).To(MatchError(ccerror.V3UnexpectedResponseError{
					V3ErrorResponse: ccerror.V3ErrorResponse{
						Errors: []ccerror.V3Error{
							{
								Code:   10008,
								Detail: "The request is semantically invalid: command presence",
								Title:  "CF-UnprocessableEntity",
							},
						},
					},
					ResponseCode: http.StatusTeapot,
				},
				))
				Expect(warnings).To(Equal(Warnings{"this is a warning"}))
			})
		})
	})
})
