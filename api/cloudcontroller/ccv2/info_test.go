package ccv2_test

import (
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Info", func() {
	var (
		serverAPIURL string

		client *Client
	)

	BeforeEach(func() {
		serverAPIURL = server.URL()[8:]
		client = NewTestClient()
	})

	Describe("Info", func() {
		BeforeEach(func() {
			response := `{
					"name":"faceman test server",
					"build":"",
					"support":"http://support.cloudfoundry.com",
					"version":0,
					"description":"",
					"authorization_endpoint":"https://login.APISERVER",
					"token_endpoint":"https://uaa.APISERVER",
					"min_cli_version":"6.22.1",
					"min_recommended_cli_version":null,
					"api_version":"2.59.0",
					"app_ssh_endpoint":"ssh.APISERVER",
					"app_ssh_host_key_fingerprint":"a6:d1:08:0b:b0:cb:9b:5f:c4:ba:44:2a:97:26:19:8a",
					"routing_endpoint": "https://APISERVER/routing",
					"app_ssh_oauth_client":"ssh-proxy",
					"logging_endpoint":"wss://loggregator.APISERVER",
					"doppler_logging_endpoint":"wss://doppler.APISERVER"
				}`
			response = strings.Replace(response, "APISERVER", serverAPIURL, -1)
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/info"),
					RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)
		})

		It("returns back the CC Information", func() {
			info, _, err := client.Info()
			Expect(err).NotTo(HaveOccurred())

			Expect(info.APIVersion).To(Equal("2.59.0"))
			Expect(info.AuthorizationEndpoint).To(MatchRegexp("https://login.%s", serverAPIURL))
			Expect(info.DopplerEndpoint).To(MatchRegexp("wss://doppler.%s", serverAPIURL))
			Expect(info.LoggregatorEndpoint).To(MatchRegexp("wss://loggregator.%s", serverAPIURL))
			Expect(info.MinimumCLIVersion).To(Equal("6.22.1"))
			Expect(info.MinimumRecommendedCLIVersion).To(BeEmpty())
			Expect(info.Name).To(Equal("faceman test server"))
			Expect(info.RoutingEndpoint).To(MatchRegexp("https://%s/routing", serverAPIURL))
			Expect(info.TokenEndpoint).To(MatchRegexp("https://uaa.%s", serverAPIURL))
		})

		It("sets the http endpoint and warns user", func() {
			_, warnings, err := client.Info()
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElement("this is a warning"))
		})
	})

	Context("when the API response gives a bad API endpoint", func() {
		BeforeEach(func() {
			response := `i am banana`
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/info"),
					RespondWith(http.StatusNotFound, response),
				),
			)
		})

		It("returns back an APINotFoundError", func() {
			_, _, err := client.Info()
			Expect(err).To(MatchError(cloudcontroller.APINotFoundError{URL: server.URL()}))
		})
	})
})
