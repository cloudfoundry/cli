package ccv3_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Buildpacks", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("GetBuildpacks", func() {
		var (
			query Query

			buildpacks []Buildpack
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			buildpacks, warnings, executeErr = client.GetBuildpacks(query)
		})

		When("buildpacks exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
					"pagination": {
						"next": {
							"href": "%s/v3/buildpacks?names=some-buildpack-name&page=2&per_page=2"
						}
					},
					"resources": [
						{
							"guid": "guid1",
							"name": "ruby_buildpack",
							"state": "AWAITING_UPLOAD",
							"stack": "windows64",
							"position": 1,
							"enabled": true,
							"locked": false
						},
						{
							"guid": "guid2",
							"name": "staticfile_buildpack",
							"state": "AWAITING_UPLOAD",
							"stack": "cflinuxfs3",
							"position": 2,
							"enabled": false,
							"locked": true
						}
					]
				}`, server.URL())
				response2 := `{
					"pagination": {
						"next": null
					},
					"resources": [
						{
							"guid": "guid3",
							"name": "go_buildpack",
							"state": "AWAITING_UPLOAD",
							"stack": "cflinuxfs2",
							"position": 3,
							"enabled": true,
							"locked": false
						}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/buildpacks", "names=some-buildpack-name"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/buildpacks", "names=some-buildpack-name&page=2&per_page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)

				query = Query{
					Key:    NameFilter,
					Values: []string{"some-buildpack-name"},
				}
			})

			It("returns the queried buildpacks and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(buildpacks).To(ConsistOf(
					Buildpack{
						Name:     "ruby_buildpack",
						GUID:     "guid1",
						Position: 1,
						Enabled:  true,
						Locked:   false,
						Stack:    "windows64",
						State:    "AWAITING_UPLOAD",
					},
					Buildpack{
						Name:     "staticfile_buildpack",
						GUID:     "guid2",
						Position: 2,
						Enabled:  false,
						Locked:   true,
						Stack:    "cflinuxfs3",
						State:    "AWAITING_UPLOAD",
					},
					Buildpack{
						Name:     "go_buildpack",
						GUID:     "guid3",
						Position: 3,
						Enabled:  true,
						Locked:   false,
						Stack:    "cflinuxfs2",
						State:    "AWAITING_UPLOAD",
					},
				))
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := `{
  "errors": [
    {
      "code": 10008,
      "detail": "The request is semantically invalid: command presence",
      "title": "CF-UnprocessableEntity"
    },
    {
      "code": 10010,
      "detail": "buildpack not found",
      "title": "CF-buildpackNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/buildpacks"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "buildpack not found",
							Title:  "CF-buildpackNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("CreateBuildpack", func() {
		var (
			inputBuildpack Buildpack

			bp         Buildpack
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			bp, warnings, executeErr = client.CreateBuildpack(inputBuildpack)
		})

		When("the buildpack is successfully created", func() {
			BeforeEach(func() {
				inputBuildpack = Buildpack{
					Name:  "some-buildpack",
					Stack: "some-stack",
				}
				response := `{
    				"guid": "some-bp-guid",
    				"created_at": "2016-03-18T23:26:46Z",
    				"updated_at": "2016-10-17T20:00:42Z",
    				"name": "some-buildpack",
    				"state": "AWAITING_UPLOAD",
    				"filename": null,
    				"stack": "some-stack",
    				"position": 42,
    				"enabled": true,
    				"locked": false,
    				"links": {
    				  "self": {
    				    "href": "/v3/buildpacks/some-bp-guid"
    				  },
						"upload": {
							"href": "/v3/buildpacks/some-bp-guid/upload",
							"method": "POST"
						}
    				}
				}`

				expectedBody := map[string]interface{}{
					"name":  "some-buildpack",
					"stack": "some-stack",
				}
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/buildpacks"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the created buildpack and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				expectedBuildpack := Buildpack{
					GUID:     "some-bp-guid",
					Name:     "some-buildpack",
					Stack:    "some-stack",
					Enabled:  true,
					Filename: "",
					Locked:   false,
					State:    constant.BuildpackAwaitingUpload,
					Position: 42,
					Links: APILinks{
						"upload": APILink{
							Method: "POST",
							HREF:   "/v3/buildpacks/some-bp-guid/upload",
						},
						"self": APILink{
							HREF: "/v3/buildpacks/some-bp-guid",
						},
					},
				}
				Expect(bp).To(Equal(expectedBuildpack))
			})
		})

		When("cc returns back an error or warnings", func() {
			BeforeEach(func() {
				inputBuildpack = Buildpack{}
				response := ` {
  "errors": [
    {
      "code": 10008,
      "detail": "The request is semantically invalid: command presence",
      "title": "CF-UnprocessableEntity"
    },
    {
      "code": 10010,
      "detail": "Buildpack not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/buildpacks"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "Buildpack not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
