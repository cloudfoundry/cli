package ccv3_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
	"net/http"
)

var _ = Describe("Job URL", func() {
	var client *Client
	var fakedMultiErrResponse string
	var expectedMultiErr ccerror.MultiError

	BeforeEach(func() {
		client, _ = NewTestClient()

		expectedMultiErr = ccerror.MultiError{
			ResponseCode: http.StatusTeapot,
			Errors: []ccerror.V3Error{
				{
					Code:   1001,
					Detail: "Request invalid due to parse error: invalid request body",
					Title:  "CF-MessageParseError",
				},
				{
					Code:   10010,
					Detail: "App not found",
					Title:  "CF-ResourceNotFound",
				},
			},
		}
		fakedMultiErrResponse = `{
		 "errors": [
		   {
		     "code": 1001,
		     "detail": "Request invalid due to parse error: invalid request body",
		     "title": "CF-MessageParseError"
		   },
		   {
		     "code": 10010,
		     "detail": "App not found",
		     "title": "CF-ResourceNotFound"
		   }
		 ]
		}`

	})

	Describe("DeleteApplication", func() {
		var (
			jobLocation JobURL
			warnings    Warnings
			executeErr  error
		)

		JustBeforeEach(func() {
			jobLocation, warnings, executeErr = client.DeleteApplication("some-app-guid")
		})

		When("the application is deleted successfully", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/apps/some-app-guid"),
						RespondWith(http.StatusAccepted, ``,
							http.Header{
								"X-Cf-Warnings": {"some-warning"},
								"Location":      {"/v3/jobs/some-location"},
							},
						),
					),
				)
			})

			It("returns all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(jobLocation).To(Equal(JobURL("/v3/jobs/some-location")))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})

		When("deleting the application returns an error", func() {
			BeforeEach(func() {
				response := `{
  "errors": [
    {
      "code": 1001,
      "detail": "Request invalid due to parse error: invalid request body",
      "title": "CF-MessageParseError"
    },
    {
      "code": 10010,
      "detail": "App not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/apps/some-app-guid"),
						RespondWith(http.StatusBadRequest, response, http.Header{"X-Cf-Warnings": {"some-warning"}}),
					),
				)
			})

			It("returns all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusBadRequest,
					Errors: []ccerror.V3Error{
						{
							Code:   1001,
							Detail: "Request invalid due to parse error: invalid request body",
							Title:  "CF-MessageParseError",
						},
						{
							Code:   10010,
							Detail: "App not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})
	})

	Describe("UpdateApplicationApplyManifest", func() {
		var (
			manifestBody []byte

			jobURL     JobURL
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.UpdateApplicationApplyManifest(
				"some-app-guid",
				manifestBody,
			)
		})

		When("the manifest application is successful", func() {
			var expectedJobURL string

			BeforeEach(func() {
				manifestBody = []byte("fake-yaml-body")
				expectedJobURL = "i-am-a-job-url"

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/actions/apply_manifest"),
						VerifyHeaderKV("Content-type", "application/x-yaml"),
						VerifyBody(manifestBody),
						RespondWith(http.StatusAccepted, "", http.Header{
							"X-Cf-Warnings": {"this is a warning"},
							"Location":      {expectedJobURL},
						}),
					),
				)
			})

			It("returns the job URL and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(jobURL).To(Equal(JobURL(expectedJobURL)))
			})
		})

		When("the manifest application fails", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/actions/apply_manifest"),
						VerifyHeaderKV("Content-type", "application/x-yaml"),
						RespondWith(http.StatusTeapot, fakedMultiErrResponse, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(expectedMultiErr))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UpdateSpaceApplyManifest", func() {
		var (
			manifestBody []byte

			responseJobURL JobURL
			warnings       Warnings
			executeErr     error

			expectedJobURL string
		)

		JustBeforeEach(func() {
			responseJobURL, warnings, executeErr = client.UpdateSpaceApplyManifest(
				"some-space-guid",
				manifestBody,
			)
		})

		BeforeEach(func() {
			manifestBody = []byte("fake-manifest-yml-body")
			expectedJobURL = "apply-manifest-job-url"
		})

		When("applying the manifest to the space succeeds", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/spaces/some-space-guid/actions/apply_manifest"),
						VerifyHeaderKV("Content-type", "application/x-yaml"),
						VerifyBody(manifestBody),
						RespondWith(http.StatusAccepted, "", http.Header{
							"X-Cf-Warnings": {"some-ccv3-warning"},
							"Location":      {expectedJobURL},
						}),
					),
				)
			})

			It("returns the job URL and warnings", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(warnings).To(ConsistOf("some-ccv3-warning"))

				Expect(responseJobURL).To(Equal(JobURL(expectedJobURL)))
			})
		})

		When("applying the manifest to the space fails", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/spaces/some-space-guid/actions/apply_manifest"),
						VerifyHeaderKV("Content-type", "application/x-yaml"),
						RespondWith(
							http.StatusTeapot, fakedMultiErrResponse, http.Header{"X-Cf-Warnings": {"some warning"}}),
					),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError(expectedMultiErr))
			})
		})
	})
})
