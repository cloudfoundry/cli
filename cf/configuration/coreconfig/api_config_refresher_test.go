package coreconfig_test

import (
	. "github.com/cloudfoundry/cli/cf/configuration/coreconfig"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ApiConfigRefresher", func() {
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
})
