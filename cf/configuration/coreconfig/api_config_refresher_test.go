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
				ccInfo = &CCInfo{}
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
				ccInfo = &CCInfo{}
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
