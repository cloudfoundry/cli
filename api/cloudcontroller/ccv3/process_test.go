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

var _ = Describe("Process", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetApplicationProcesses", func() {
		Context("when the application exists", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
					"pagination": {
						"next": {
							"href": "%s/v3/apps/some-app-guid/processes?page=2"
						}
					},
					"resources": [
						{
							"guid": "process-1-guid",
							"type": "web"
						},
						{
							"guid": "process-2-guid",
							"type": "worker"
						}
					]
				}`, server.URL())
				response2 := `{
					"pagination": {
						"next": null
					},
					"resources": [
						{
							"guid": "process-3-guid",
							"type": "console"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/processes"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/processes", "page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					),
				)
			})

			It("returns a list of processes associated with the application and all warnings", func() {
				processes, warnings, err := client.GetApplicationProcesses("some-app-guid")
				Expect(err).ToNot(HaveOccurred())

				Expect(processes).To(ConsistOf(
					Process{
						GUID: "process-1-guid",
						Type: "web",
					},
					Process{
						GUID: "process-2-guid",
						Type: "worker",
					},
					Process{
						GUID: "process-3-guid",
						Type: "console",
					},
				))
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
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/processes"),
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns the error", func() {
				_, _, err := client.GetApplicationProcesses("some-app-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: "App not found"}))
			})
		})
	})
})
