package ccv3_test

import (
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
