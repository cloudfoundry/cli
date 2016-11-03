package ccv3_test

import (
	"fmt"
	"net/http"
	"net/url"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Application", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetApplications", func() {
		Context("when apps exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
  "next_url": "%s/v3/apps?space_guids=some-space-guid&names=some-app-name&page=2&per_page=2",
  "resources": [
    {
      "name": "app-name-1",
      "guid": "app-guid-1"
    },
    {
      "name": "app-name-2",
      "guid": "app-guid-2"
    }
  ]
}`, server.URL())
				response2 := `{
	"next_url": null,
	"resources": [
	  {
      "name": "app-name-3",
		  "guid": "app-guid-3"
		}
	]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps", "space_guids=some-space-guid&names=some-app-name"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps", "space_guids=some-space-guid&names=some-app-name&page=2&per_page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the queried apps", func() {
				apps, warnings, err := client.GetApplications(url.Values{
					"space_guids": []string{"some-space-guid"},
					"names":       []string{"some-app-name"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(apps).To(ConsistOf([]Application{
					{Name: "app-name-1", GUID: "app-guid-1"},
					{Name: "app-name-2", GUID: "app-guid-2"},
					{Name: "app-name-3", GUID: "app-guid-3"},
				}))
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))
			})
		})

		Context("when the cloud controller returns an error", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps", "space_guids=some-space-guid&names=some-app-name"),
						RespondWith(http.StatusUnprocessableEntity,
							`{ "errors": [{}] }`,
							http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := client.GetApplications(url.Values{
					"space_guids": []string{"some-space-guid"},
					"names":       []string{"some-app-name"},
				})
				Expect(err).To(MatchError(UnexpectedResponseError{
					ResponseCode: 422,
					CCErrorResponse: CCErrorResponse{
						Errors: []CCError{
							CCError{},
						}},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
