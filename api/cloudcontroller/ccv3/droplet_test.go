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

var _ = Describe("Droplet", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetApplicationCurrentDroplet", func() {
		Context("when the application exists", func() {
			BeforeEach(func() {
				response := fmt.Sprintf(`{
					"stack": "some-stack",
					"buildpacks": [{
						"name": "some-buildpack",
						"detect_output": "detected-buildpack"
					}]
				}`, server.URL())
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/droplets/current"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the current droplet for the given app and all warnings", func() {
				droplet, warnings, err := client.GetApplicationCurrentDroplet("some-app-guid")
				Expect(err).ToNot(HaveOccurred())

				Expect(droplet).To(Equal(Droplet{
					Stack: "some-stack",
					Buildpacks: []Buildpack{
						{
							Name:         "some-buildpack",
							DetectOutput: "detected-buildpack",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1"))
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
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/droplets/current"),
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns the error", func() {
				_, _, err := client.GetApplicationCurrentDroplet("some-app-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: "App not found"}))
			})
		})
	})
})
