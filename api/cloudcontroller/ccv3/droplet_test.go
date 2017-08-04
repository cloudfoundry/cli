package ccv3_test

import (
	"fmt"
	"net/http"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Droplet", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetApplicationDroplets", func() {
		Context("when the application exists", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
					"pagination": {
						"next": {
							"href": "%s/v3/apps/some-app-guid/droplets?current=true&per_page=2&page=2"
						}
					},
					"resources": [
						{
							"stack": "some-stack",
							"buildpacks": [{
								"name": "some-buildpack",
								"detect_output": "detected-buildpack"
							}]
						},
						{
							"stack": "some-stack2",
							"buildpacks": [{
								"name": "some-buildpack2",
								"detect_output": "detected-buildpack2"
							}]
						}
					]
				}`, server.URL())
				response2 := `{
					"pagination": {
						"next": null
					},
					"resources": [
						{
							"stack": "some-stack3",
							"buildpacks": [{
								"name": "some-buildpack3",
								"detect_output": "detected-buildpack3"
							}]
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/droplets", "current=true&per_page=2"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/droplets", "current=true&per_page=2&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					),
				)
			})

			It("returns the current droplet for the given app and all warnings", func() {
				droplets, warnings, err := client.GetApplicationDroplets("some-app-guid", url.Values{"per_page": []string{"2"}, "current": []string{"true"}})
				Expect(err).ToNot(HaveOccurred())
				Expect(droplets).To(HaveLen(3))

				Expect(droplets[0]).To(Equal(Droplet{
					Stack: "some-stack",
					Buildpacks: []Buildpack{
						{
							Name:         "some-buildpack",
							DetectOutput: "detected-buildpack",
						},
					},
				}))
				Expect(droplets[1]).To(Equal(Droplet{
					Stack: "some-stack2",
					Buildpacks: []Buildpack{
						{
							Name:         "some-buildpack2",
							DetectOutput: "detected-buildpack2",
						},
					},
				}))
				Expect(droplets[2]).To(Equal(Droplet{
					Stack: "some-stack3",
					Buildpacks: []Buildpack{
						{
							Name:         "some-buildpack3",
							DetectOutput: "detected-buildpack3",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		Context("when cloud controller returns an error", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10010,
							"detail": "App not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/droplets"),
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns the error", func() {
				_, _, err := client.GetApplicationDroplets("some-app-guid", url.Values{})
				Expect(err).To(MatchError(ccerror.ApplicationNotFoundError{}))
			})
		})
	})
})
