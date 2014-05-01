package net_test

import (
	"github.com/cloudfoundry/cli/cf/net"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("WarningsCollector", func() {
	var (
		ui *testterm.FakeUI
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
	})

	It("prints warnings in all warning producers", func() {
		warning_producer_one := testnet.NewWarningProducer([]string{"Hello", "Darling"})
		warning_producer_two := testnet.NewWarningProducer([]string{"Goodbye", "Sweetie"})
		warning_producer_three := testnet.NewWarningProducer(nil)
		warnings_collector := net.NewWarningsCollector(ui, warning_producer_one, warning_producer_two, warning_producer_three)

		warnings_collector.PrintWarnings()

		Expect(ui.WarnOutputs).To(ContainSubstrings(
			[]string{"Hello"},
			[]string{"Darling"},
			[]string{"Goodbye"},
			[]string{"Sweetie"},
		))
	})
})
