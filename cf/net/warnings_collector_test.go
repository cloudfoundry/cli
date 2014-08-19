package net_test

import (
	"github.com/cloudfoundry/cli/cf/net"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("WarningsCollector", func() {
	var (
		ui                 *testterm.FakeUI
		oldRaiseErrorValue string
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

		Context("when the RAISE_ERROR_ON_WARNINGS environment variable is set", func() {
			BeforeEach(func() {
				os.Setenv("CF_RAISE_ERROR_ON_WARNINGS", "true")
			})

			It("panics with an error that contains all the warnings", func() {
				warning_producer_one := testnet.NewWarningProducer([]string{"Hello", "Darling"})
				warning_producer_two := testnet.NewWarningProducer([]string{"Goodbye", "Sweetie"})
				warning_producer_three := testnet.NewWarningProducer(nil)
				warnings_collector := net.NewWarningsCollector(ui, warning_producer_one, warning_producer_two, warning_producer_three)

				Expect(warnings_collector.PrintWarnings).To(Panic())
			})
		})

		Context("when the RAISE_ERROR_ON_WARNINGS environment variable is not set", func() {
			BeforeEach(func() {
				os.Setenv("CF_RAISE_ERROR_ON_WARNINGS", "")
			})

			It("does not panic", func() {
				warning_producer_one := testnet.NewWarningProducer([]string{"Hello", "Darling"})
				warning_producer_two := testnet.NewWarningProducer([]string{"Goodbye", "Sweetie"})
				warning_producer_three := testnet.NewWarningProducer(nil)
				warnings_collector := net.NewWarningsCollector(ui, warning_producer_one, warning_producer_two, warning_producer_three)

				Expect(warnings_collector.PrintWarnings).NotTo(Panic())
			})
		})
	})

})
