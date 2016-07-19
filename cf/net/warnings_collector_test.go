package net_test

import (
	"os"

	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/net/netfakes"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("WarningsCollector", func() {
	var (
		ui                 *terminalfakes.FakeUI
		oldRaiseErrorValue string
		warningsCollector  net.WarningsCollector
	)

	BeforeEach(func() {
		ui = new(terminalfakes.FakeUI)
	})

	Describe("PrintWarnings", func() {
		BeforeEach(func() {
			oldRaiseErrorValue = os.Getenv("CF_RAISE_ERROR_ON_WARNINGS")
		})

		AfterEach(func() {
			os.Setenv("CF_RAISE_ERROR_ON_WARNINGS", oldRaiseErrorValue)
		})

		Context("when the CF_RAISE_ERROR_ON_WARNINGS environment variable is set", func() {
			BeforeEach(func() {
				os.Setenv("CF_RAISE_ERROR_ON_WARNINGS", "true")
			})

			Context("when there are warnings", func() {
				BeforeEach(func() {
					warning_producer_one := new(netfakes.FakeWarningProducer)
					warning_producer_one.WarningsReturns([]string{"something"})
					warningsCollector = net.NewWarningsCollector(ui, warning_producer_one)
				})

				It("returns an error", func() {
					err := warningsCollector.PrintWarnings()
					Expect(err).To(HaveOccurred())
				})
			})

			Context("when there are no warnings", func() {
				BeforeEach(func() {
					warningsCollector = net.NewWarningsCollector(ui)
				})

				It("does not return an error", func() {
					err := warningsCollector.PrintWarnings()
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when the CF_RAISE_ERROR_ON_WARNINGS environment variable is not set", func() {
			BeforeEach(func() {
				os.Setenv("CF_RAISE_ERROR_ON_WARNINGS", "")
			})

			It("does not return an error", func() {
				warning_producer_one := new(netfakes.FakeWarningProducer)
				warning_producer_one.WarningsReturns([]string{"Hello", "Darling"})
				warningsCollector := net.NewWarningsCollector(ui, warning_producer_one)

				err := warningsCollector.PrintWarnings()
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not print out duplicate warnings", func() {
				warning_producer_one := new(netfakes.FakeWarningProducer)
				warning_producer_one.WarningsReturns([]string{"Hello Darling"})
				warning_producer_two := new(netfakes.FakeWarningProducer)
				warning_producer_two.WarningsReturns([]string{"Hello Darling"})
				warningsCollector := net.NewWarningsCollector(ui, warning_producer_one, warning_producer_two)

				warningsCollector.PrintWarnings()
				Expect(ui.WarnCallCount()).To(Equal(1))
				Expect(ui.WarnArgsForCall(0)).To(ContainSubstring("Hello Darling"))
			})

			It("does not print out Endpoint deprecated warnings", func() {
				warning_producer_one := new(netfakes.FakeWarningProducer)
				warning_producer_one.WarningsReturns([]string{"Endpoint deprecated"})
				warning_producer_two := new(netfakes.FakeWarningProducer)
				warning_producer_two.WarningsReturns([]string{"A warning"})
				warningsCollector := net.NewWarningsCollector(ui, warning_producer_one, warning_producer_two)

				warningsCollector.PrintWarnings()
				Expect(ui.WarnCallCount()).To(Equal(1))
				Expect(ui.WarnArgsForCall(0)).To(ContainSubstring("A warning"))
			})
		})
	})
})
