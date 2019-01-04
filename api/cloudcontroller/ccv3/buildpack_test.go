package ccv3_test

import (
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
})
