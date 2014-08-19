package net_test

import (
	"os"

	"github.com/cloudfoundry/cli/cf/net"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("WarningsCollector", func() {
	var (
		ui                 *testterm.FakeUI
		oldRaiseErrorValue string
		warningsCollector  net.WarningsCollector
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
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
					warning_producer_one := testnet.NewWarningProducer([]string{"Hello", "Darling"})
					warning_producer_two := testnet.NewWarningProducer([]string{"Goodbye", "Sweetie"})
					warning_producer_three := testnet.NewWarningProducer(nil)
					warningsCollector = net.NewWarningsCollector(ui, warning_producer_one, warning_producer_two, warning_producer_three)
				})

				It("panics with an error that contains all the warnings", func() {
					Expect(warningsCollector.PrintWarnings).To(Panic())
				})
			})

			Context("when there are no warnings", func() {
				BeforeEach(func() {
					warningsCollector = net.NewWarningsCollector(ui)
				})

				It("does not panic", func() {
					Expect(warningsCollector.PrintWarnings).NotTo(Panic())
				})

			})
		})

		Context("when the CF_RAISE_ERROR_ON_WARNINGS environment variable is not set", func() {
			BeforeEach(func() {
				os.Setenv("CF_RAISE_ERROR_ON_WARNINGS", "")
			})

			It("does not panic", func() {
				warning_producer_one := testnet.NewWarningProducer([]string{"Hello", "Darling"})
				warning_producer_two := testnet.NewWarningProducer([]string{"Goodbye", "Sweetie"})
				warning_producer_three := testnet.NewWarningProducer(nil)
				warningsCollector := net.NewWarningsCollector(ui, warning_producer_one, warning_producer_two, warning_producer_three)

				Expect(warningsCollector.PrintWarnings).NotTo(Panic())
			})
		})
	})

})
