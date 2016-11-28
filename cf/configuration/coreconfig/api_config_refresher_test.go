package coreconfig_test

import (
	. "code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/util/testhelpers/configuration"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig/coreconfigfakes"
	"code.cloudfoundry.org/cli/cf/i18n"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("APIConfigRefresher", func() {
	Describe("LoggregatorEndpoint", func() {
		r := APIConfigRefresher{}

		var (
			ccInfo *CCInfo
		)

		Context("when the loggregator endpoint is specified in the CCInfo", func() {
			BeforeEach(func() {
				ccInfo = &CCInfo{
					LoggregatorEndpoint: "some-endpoint",
				}
			})

			It("returns the endpoint from the CCInfo", func() {
				endpoint := r.LoggregatorEndpoint(ccInfo, "a-url-that-doesn't-matter")

				Expect(endpoint).To(Equal("some-endpoint"))
			})
		})

		Context("when the loggregator endpoint is not specified in the CCInfo", func() {
			BeforeEach(func() {
				ccInfo = &CCInfo{
					LoggregatorEndpoint: "",
				}
			})

			It("extrapolates the loggregator URL based on the API URL (SSL API)", func() {
				endpoint := r.LoggregatorEndpoint(ccInfo, "https://127.0.0.1:443")
				Expect(endpoint).To(Equal("wss://loggregator.0.0.1:443"))
			})

			It("extrapolates the loggregator URL if there is a trailing slash", func() {
				endpoint := r.LoggregatorEndpoint(ccInfo, "https://127.0.0.1:443/")
				Expect(endpoint).To(Equal("wss://loggregator.0.0.1:443"))
			})

			It("extrapolates the loggregator URL based on the API URL (non-SSL API)", func() {
				endpoint := r.LoggregatorEndpoint(ccInfo, "http://127.0.0.1:80")
				Expect(endpoint).To(Equal("ws://loggregator.0.0.1:80"))
			})
		})
	})
	Describe("Refresh", func() {
		BeforeEach(func() {
			config := configuration.NewRepositoryWithDefaults()
			i18n.T = i18n.Init(config)
		})

		Context("when the cloud controller returns an insecure api endpoint", func() {
			var (
				r            APIConfigRefresher
				ccInfo       *CCInfo
				endpointRepo *coreconfigfakes.FakeEndpointRepository
			)

			BeforeEach(func() {
				ccInfo = &CCInfo{
					LoggregatorEndpoint: "some-endpoint",
				}
				endpointRepo = new(coreconfigfakes.FakeEndpointRepository)

				r = APIConfigRefresher{
					EndpointRepo: endpointRepo,
					Config:       new(coreconfigfakes.FakeReadWriter),
					Endpoint:     "api.some.endpoint.com",
				}
			})

			It("gives a warning", func() {
				endpointRepo.GetCCInfoReturns(ccInfo, "api.some.endpoint.com", nil)
				warning, err := r.Refresh()
				Expect(err).NotTo(HaveOccurred())
				Expect(warning.Warn()).To(Equal("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended\n"))
			})
		})

		Context("when the cloud controller returns a secure api endpoint", func() {
			var (
				r            APIConfigRefresher
				ccInfo       *CCInfo
				endpointRepo *coreconfigfakes.FakeEndpointRepository
			)

			BeforeEach(func() {
				ccInfo = &CCInfo{
					LoggregatorEndpoint: "some-endpoint",
				}
				endpointRepo = new(coreconfigfakes.FakeEndpointRepository)

				r = APIConfigRefresher{
					EndpointRepo: endpointRepo,
					Config:       new(coreconfigfakes.FakeReadWriter),
					Endpoint:     "api.some.endpoint.com",
				}
			})

			It("gives a warning", func() {
				endpointRepo.GetCCInfoReturns(ccInfo, "https://api.some.endpoint.com", nil)
				warning, err := r.Refresh()
				Expect(err).NotTo(HaveOccurred())
				Expect(warning).To(BeNil())
			})
		})
	})
})
