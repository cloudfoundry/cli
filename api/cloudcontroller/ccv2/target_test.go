package ccv2_test

import (
	"net/http"
	"strings"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Target", func() {
	var (
		serverAPIURL string

		client *Client
	)

	BeforeEach(func() {
		serverAPIURL = server.URL()[8:]
		client = NewClient("CF CLI API Target Test", "Unknown")
	})

	Describe("TargetCF", func() {
		BeforeEach(func() {
			response := `{
					"name":"",
					"build":"",
					"support":"http://support.cloudfoundry.com",
					"version":0,
					"description":"",
					"authorization_endpoint":"https://login.APISERVER",
					"token_endpoint":"https://uaa.APISERVER",
					"min_cli_version":null,
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

		Context("when passed a valid API URL", func() {
			Context("when the api has unverified SSL", func() {
				Context("when setting the skip ssl flat", func() {
					It("sets all the endpoints on the client", func() {
						_, err := client.TargetCF(TargetSettings{
							SkipSSLValidation: true,
							URL:               server.URL(),
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(client.API()).To(MatchRegexp("https://%s", serverAPIURL))
						Expect(client.APIVersion()).To(Equal("2.59.0"))
						Expect(client.AuthorizationEndpoint()).To(MatchRegexp("https://login.%s", serverAPIURL))
						Expect(client.DopplerEndpoint()).To(MatchRegexp("wss://doppler.%s", serverAPIURL))
						Expect(client.LoggregatorEndpoint()).To(MatchRegexp("wss://loggregator.%s", serverAPIURL))
						Expect(client.RoutingEndpoint()).To(MatchRegexp("https://%s/routing", serverAPIURL))
						Expect(client.TokenEndpoint()).To(MatchRegexp("https://uaa.%s", serverAPIURL))
					})
				})

				It("sets the http endpoint and warns user", func() {
					warnings, err := client.TargetCF(TargetSettings{
						SkipSSLValidation: true,
						URL:               server.URL(),
					})
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ContainElement("this is a warning"))
				})
			})
		})
	})
})
