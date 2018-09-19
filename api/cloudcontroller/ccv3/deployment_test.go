package ccv3_test

import (
	"fmt"
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Task", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetDeployments", func() {
		var (
			deployments []Deployment
			warnings    Warnings
			executeErr  error
		)

		JustBeforeEach(func() {
			deployments, warnings, executeErr = client.GetDeployments(Query{Key: AppGUIDFilter, Values: []string{"some-app-guid"}}, Query{Key: OrderBy, Values: []string{"-created_at"}}, Query{Key: PerPage, Values: []string{"1"}})
		})

		var response string
		var response2 string
		BeforeEach(func() {
			response = fmt.Sprintf(`{
	"pagination": {
		"next": {
			"href": "%s/v3/deployments?app_guids=some-app-guid&order_by=-created_at&page=2&per_page=1"
		}
	},
	"resources": [
		{
      		"guid": "newest-deployment-guid",
      		"created_at": "2018-05-25T22:42:10Z"
   	}
	]
}`, server.URL())
			response2 = `{
  	"pagination": {
		"next": null
	},
	"resources": [
		{
      		"guid": "oldest-deployment-guid",
      		"created_at": "2018-04-25T22:42:10Z"
   	}
	]
}`
		})

		Context("when the deployment exists", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/deployments", "app_guids=some-app-guid&order_by=-created_at&per_page=1"),
						RespondWith(http.StatusAccepted, response),
					),
				)

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/deployments", "app_guids=some-app-guid&order_by=-created_at&page=2&per_page=1"),
						RespondWith(http.StatusAccepted, response2, http.Header{"X-Cf-Warnings": {"warning"}}),
					),
				)

			})

			It("returns the deployment guid of the most recent deployment", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning"))
				Expect(deployments).To(ConsistOf(
					Deployment{GUID: "newest-deployment-guid", CreatedAt: "2018-05-25T22:42:10Z"},
					Deployment{GUID: "oldest-deployment-guid", CreatedAt: "2018-04-25T22:42:10Z"},
				))
			})
		})
	})

	Describe("CancelDeployment", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = client.CancelDeployment("some-deployment-guid")
		})

		Context("when the deployment exists", func() {
			Context("when cancelling the deployment succeeds", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/deployments/some-deployment-guid/actions/cancel"),
							RespondWith(http.StatusAccepted, "", http.Header{"X-Cf-Warnings": {"warning"}}),
						),
					)
				})

				It("cancels the deployment with no errors and returns all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning"))
				})
			})
		})
	})

	Describe("CreateApplicationDeployment", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = client.CreateApplicationDeployment("some-app-guid")
		})

		Context("when the application exists", func() {
			var response string
			BeforeEach(func() {
				response = `{
  "guid": "some-deployment-guid",
  "created_at": "2018-04-25T22:42:10Z",
  "relationships": {
    "app": {
      "data": {
        "guid": "some-app-guid"
      }
    }
  }
}`
			})

			Context("when creating the deployment succeeds", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/deployments"),
							VerifyJSON(`{"relationships":{"app":{"data":{"guid":"some-app-guid"}}}}`),
							RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"warning"}}),
						),
					)
				})

				It("creates the deployment with no errors and returns all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning"))
				})
			})
		})
	})
})
