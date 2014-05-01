package terminal_test

import (
	. "github.com/cloudfoundry/cli/cf/terminal"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Table", func() {
	var (
		ui    *testterm.FakeUI
		table Table
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		table = NewTable(ui, []string{"watashi", "no", "atama!"})
	})

	It("prints the header", func() {
		table.Print()
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"watashi", "no", "atama!"},
		))
	})

	It("prints format string literals as strings", func() {
		table.Add([]string{"cloak %s", "and", "dagger"})
		table.Print()

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"cloak %s", "and", "dagger"},
		))
	})

	It("prints all the rows you give it", func() {
		table.Add([]string{"something", "and", "nothing"})
		table.Print()
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"something", "and", "nothing"},
		))
	})

	Describe("adding rows to be printed later", func() {
		It("prints them when you call Print()", func() {
			table.Add([]string{"a", "b", "c"})
			table.Add([]string{"passed", "to", "print"})
			table.Print()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"a", "b", "c"},
			))
		})

		It("flushes previously added rows and then outputs passed rows", func() {
			table.Add([]string{"a", "b", "c"})
			table.Add([]string{"passed", "to", "print"})
			table.Print()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"watashi", "no", "atama!"},
				[]string{"a", "b", "c"},
				[]string{"passed", "to", "print"},
			))
		})

		It("flushes the buffer of rows when you call print", func() {
			table.Add([]string{"a", "b", "c"})
			table.Add([]string{"passed", "to", "print"})
			table.Print()
			ui.ClearOutputs()

			table.Print()
			Expect(ui.Outputs).To(BeEmpty())
		})
	})

	Describe("aligning columns", func() {
		It("aligns rows to the header when the header is longest", func() {
			table.Add([]string{"a", "b", "c"})
			table.Print()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"watashi   no   atama!"},
				[]string{"a         b    c"},
			))
		})

		It("aligns rows to the longest row provided", func() {
			table.Add([]string{"x", "y", "z"})
			table.Add([]string{"something", "something", "darkside"})
			table.Print()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"watashi     no          atama!"},
				[]string{"x           y           z"},
				[]string{"something   something   darkside"},
			))
		})
	})
})
