package strategy_test

import (
	"code.cloudfoundry.org/cli/cf/api/resources"
	. "code.cloudfoundry.org/cli/cf/api/strategy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EndpointStrategy", func() {
	var strategy EndpointStrategy

	Describe("events", func() {
		Context("when the version string can't be parsed", func() {
			BeforeEach(func() {
				strategy = NewEndpointStrategy("")
			})

			It("uses the oldest possible strategy", func() {
				Expect(strategy.EventsURL("the-guid", 20)).To(Equal("/v2/apps/the-guid/events?results-per-page=20"))
				Expect(strategy.EventsResource()).To(BeAssignableToTypeOf(resources.EventResourceOldV2{}))
			})
		})

		Context("when targeting a pre-2.1.0 cloud controller", func() {
			BeforeEach(func() {
				strategy = NewEndpointStrategy("2.0.0")
			})

			It("returns an appropriate endpoint", func() {
				Expect(strategy.EventsURL("the-guid", 20)).To(Equal("/v2/apps/the-guid/events?results-per-page=20"))
			})

			It("returns an old EventResource", func() {
				Expect(strategy.EventsResource()).To(BeAssignableToTypeOf(resources.EventResourceOldV2{}))
			})
		})

		Context("when targeting a 2.1.0 cloud controller", func() {
			BeforeEach(func() {
				strategy = NewEndpointStrategy("2.1.0")
			})

			It("returns an appropriate endpoint", func() {
				Expect(strategy.EventsURL("guids-r-us", 42)).To(Equal("/v2/events?order-direction=desc&q=actee%3Aguids-r-us&results-per-page=42"))
			})

			It("returns a new EventResource", func() {
				Expect(strategy.EventsResource()).To(BeAssignableToTypeOf(resources.EventResourceNewV2{}))
			})
		})
	})

	Describe("domains", func() {
		Context("when targeting a pre-2.1.0 cloud controller", func() {
			BeforeEach(func() {
				strategy = NewEndpointStrategy("2.0.0")
			})

			It("uses the general domains endpoint", func() {
				Expect(strategy.PrivateDomainsURL()).To(Equal("/v2/domains"))
			})
		})

		Context("when targeting a v2.1.0 cloud controller", func() {
			BeforeEach(func() {
				strategy = NewEndpointStrategy("2.1.0")
			})

			It("uses the private domains endpoint", func() {
				Expect(strategy.PrivateDomainsURL()).To(Equal("/v2/private_domains"))
			})
		})
	})
})
