package terminal_test

import (
	. "cf/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "testhelpers/matchers"
	testterm "testhelpers/terminal"
)

var _ = Describe("Table", func() {
	It("prints format string literals as strings", func() {
		ui := &testterm.FakeUI{}
		table := NewTable(ui, []string{"watashi no atama!"})

		table.Print([][]string{{"cloak %s and dagger"}})

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"watashi no atama!"},
			[]string{"cloak %s and dagger"},
		))
	})
})
