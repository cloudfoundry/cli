package strategy_test

import (
	"cf/api/resources"
	. "cf/api/strategy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EndpointStrategy", func() {
	var (
		strategy EndpointStrategy
		err      error
	)

	Describe("events", func() {
		Context("when targeting a v2.0.0 cloud foundry", func() {
			BeforeEach(func() {
				strategy, err = NewEndpointStrategy("2.0.0")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an appropriate endpoint", func() {
				Expect(strategy.EventsURL("the-guid", 20)).To(Equal("/v2/apps/the-guid/events?results-per-page=20"))
			})

			It("returns an old EventResource", func() {
				Expect(strategy.EventsResource()).To(BeAssignableToTypeOf(resources.EventResourceOldV2{}))
			})
		})

		Context("when targeting a v2.2.0 cloud foundry", func() {
			BeforeEach(func() {
				strategy, err = NewEndpointStrategy("2.2.1")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an appropriate endpoint", func() {
				Expect(strategy.EventsURL("guids-r-us", 42)).To(Equal("/v2/events?order-direction=desc&q=actee%3Aguids-r-us&results-per-page=42"))
			})

			It("returns a new EventResource", func() {
				Expect(strategy.EventsResource()).To(BeAssignableToTypeOf(resources.EventResourceNewV2{}))
			})
		})
	})
})
