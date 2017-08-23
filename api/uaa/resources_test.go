package uaa_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/uaa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("SetupResources", func() {
	var (
		client            *Client
		setupResourcesErr error
	)

	JustBeforeEach(func() {
		setupResourcesErr = client.SetupResources(server.URL())
	})

	BeforeEach(func() {
		client = NewClient(Config{
			AppName:           "CF CLI UAA API Test",
			AppVersion:        "Unknown",
			ClientID:          "client-id",
			ClientSecret:      "client-secret",
			SkipSSLValidation: true,
		})
	})

	Context("when the authentication server returns an error", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/login"),
					RespondWith(http.StatusNotFound, `{"errors": [{}]}`, nil),
				),
			)
		})

		It("returns the error", func() {
			Expect(setupResourcesErr).To(HaveOccurred())
		})
	})

	Context("when the request succeeds", func() {
		BeforeEach(func() {
			response := `{
				"links": {
					"uaa": "https://uaa.bosh-lite.com"
				}
			}`

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/login"),
					RespondWith(http.StatusOK, response, nil),
				),
			)
		})

		It("does not return an error", func() {
			Expect(setupResourcesErr).ToNot(HaveOccurred())
		})
	})
})
