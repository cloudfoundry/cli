package ccv2_test

import (
	"net/http"
	"strings"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/ccv2fakes"

	. "github.com/onsi/ginkgo/v2"
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
	})

	Describe("TargetCF", func() {
		var response string
		var fakeWrapper1 *ccv2fakes.FakeConnectionWrapper
		var fakeWrapper2 *ccv2fakes.FakeConnectionWrapper

		BeforeEach(func() {
			response = `{
					"name":"",
					"build":"",
					"support":"http://support.cloudfoundry.com",
					"version":0,
					"description":"",
					"authorization_endpoint":"https://login.APISERVER",
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
			fakeWrapper1 = new(ccv2fakes.FakeConnectionWrapper)
			fakeWrapper1.WrapReturns(fakeWrapper1)
			fakeWrapper2 = new(ccv2fakes.FakeConnectionWrapper)
			fakeWrapper2.WrapReturns(fakeWrapper2)

			client = NewClient(Config{
				AppName:    "CF CLI API Target Test",
				AppVersion: "Unknown",
				Wrappers:   []ConnectionWrapper{fakeWrapper1, fakeWrapper2},
			})

		})
		When("using a older API that does not have the log cache url", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/info"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/"),
						RespondWith(http.StatusOK, `{ "links": "someurl" : "cool beans"}}`),
					),
				)

			})
			It("should string parse the API url to add the log cache url", func() {
				_, err := client.TargetCF(TargetSettings{
					SkipSSLValidation: true,
					URL:               "api.fun.com",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(client.LogCacheEndpoint()).To(Equal("log-cache.fun.com"))
			})
		})
		BeforeEach(func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/info"),
					RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/"),
					RespondWith(http.StatusOK, `{ "links": {"log_cache": {"href": "api.coolbeans.log-cache"}}}`),
				),
			)
		})

		When("client has wrappers", func() {

			BeforeEach(func() {
				fakeWrapper1 = new(ccv2fakes.FakeConnectionWrapper)
				fakeWrapper1.WrapReturns(fakeWrapper1)
				fakeWrapper2 = new(ccv2fakes.FakeConnectionWrapper)
				fakeWrapper2.WrapReturns(fakeWrapper2)

				client = NewClient(Config{
					AppName:    "CF CLI API Target Test",
					AppVersion: "Unknown",
					Wrappers:   []ConnectionWrapper{fakeWrapper1, fakeWrapper2},
				})
			})

			It("calls wrap on all the wrappers", func() {
				_, err := client.TargetCF(TargetSettings{
					SkipSSLValidation: true,
					URL:               server.URL(),
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeWrapper1.WrapCallCount()).To(Equal(1))
				Expect(fakeWrapper2.WrapCallCount()).To(Equal(1))
				Expect(fakeWrapper2.WrapArgsForCall(0)).To(Equal(fakeWrapper1))
			})
		})

		When("passed a valid API URL", func() {
			BeforeEach(func() {
				client = NewClient(Config{AppName: "CF CLI API Target Test", AppVersion: "Unknown"})
			})

			When("the api has unverified SSL", func() {
				When("setting the skip ssl flat", func() {
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
						Expect(client.RoutingEndpoint()).To(MatchRegexp("https://%s/routing", serverAPIURL))
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
