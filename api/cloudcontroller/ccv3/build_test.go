package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Build", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("CreateBuild", func() {
		Context("when the build successfully is created", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-build-guid",
					"state": "STAGING",
					"droplet": {
						"guid": "some-droplet-guid"
					}
				}`

				expectedBody := map[string]interface{}{
					"package": map[string]interface{}{
						"guid": "some-package-guid",
					},
				}
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/builds"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the created build and warnings", func() {
				build, warnings, err := client.CreateBuild(Build{PackageGUID: "some-package-guid"})

				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(build).To(Equal(Build{
					GUID:        "some-build-guid",
					State:       constant.BuildStaging,
					DropletGUID: "some-droplet-guid",
				}))
			})
		})

		Context("when cc returns back an error or warnings", func() {
			BeforeEach(func() {
				response := ` {
  "errors": [
    {
      "code": 10008,
      "detail": "I can't even",
      "title": "CF-UnprocessableEntity"
    },
    {
      "code": 10010,
      "detail": "Package not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/builds"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := client.CreateBuild(Build{PackageGUID: "some-package-guid"})
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
						Errors: []ccerror.V3Error{
							{
								Code:   10008,
								Detail: "I can't even",
								Title:  "CF-UnprocessableEntity",
							},
							{
								Code:   10010,
								Detail: "Package not found",
								Title:  "CF-ResourceNotFound",
							},
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetBuild", func() {
		Context("when the build exist", func() {
			BeforeEach(func() {
				response := `{
					"created_at": "some-time",
					"guid": "some-build-guid",
					"state": "FAILED",
					"error": "some error",
					"droplet": {
						"guid": "some-droplet-guid"
					}
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/builds/some-build-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the queried build and all warnings", func() {
				build, warnings, err := client.GetBuild("some-build-guid")
				Expect(err).NotTo(HaveOccurred())

				expectedBuild := Build{
					CreatedAt:   "some-time",
					GUID:        "some-build-guid",
					State:       constant.BuildFailed,
					Error:       "some error",
					DropletGUID: "some-droplet-guid",
				}
				Expect(build).To(Equal(expectedBuild))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		Context("when the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := ` {
					"errors": [
						{
							"code": 10008,
							"detail": "I can't even",
							"title": "CF-UnprocessableEntity"
						},
						{
							"code": 10010,
							"detail": "Build not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/builds/some-build-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := client.GetBuild("some-build-guid")
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
						Errors: []ccerror.V3Error{
							{
								Code:   10008,
								Detail: "I can't even",
								Title:  "CF-UnprocessableEntity",
							},
							{
								Code:   10010,
								Detail: "Build not found",
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
