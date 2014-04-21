package terminal_test

import (
	. "cf/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "testhelpers/matchers"
	testterm "testhelpers/terminal"
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
		table.Print([][]string{})
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"watashi", "no", "atama!"},
		))
	})

	It("prints format string literals as strings", func() {
		table.Print([][]string{{"cloak %s", "and", "dagger"}})

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"cloak %s", "and", "dagger"},
		))
	})

	It("prints all the rows you give it", func() {
		table.Print([][]string{{"something", "and", "nothing"}})
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"something", "and", "nothing"},
		))
	})

	Describe("adding rows to be printed later", func() {
		It("prints them when you call Print()", func() {
			table.Add([]string{"a", "b", "c"})
			table.Print([][]string{{"passed", "to", "print"}})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"a", "b", "c"},
			))
		})

		It("flushes previously added rows and then outputs passed rows", func() {
			table.Add([]string{"a", "b", "c"})
			table.Print([][]string{{"passed", "to", "print"}})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"watashi", "no", "atama!"},
				[]string{"a", "b", "c"},
				[]string{"passed", "to", "print"},
			))
		})

		It("flushes the buffer of rows when you call print", func() {
			table.Add([]string{"a", "b", "c"})
			table.Print([][]string{{"passed", "to", "print"}})
			ui.ClearOutputs()

			table.Print([][]string{{"passed", "to", "print"}})
			Expect(ui.Outputs).ToNot(ContainSubstrings(
				[]string{"a", "b", "c"},
			))
		})
	})

	Describe("aligning columns", func() {
		It("aligns rows to the header when the header is longest", func() {
			table.Add([]string{"a", "b", "c"})
			table.Print([][]string{{"d", "e", "f"}})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"watashi   no   atama!"},
				[]string{"a         b    c"},
			))
		})

		It("aligns rows to the row when the row is longest", func() {
			table.Add([]string{"something", "something", "darkside"})
			table.Print([][]string{{"x", "y", "z"}})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"watashi     no          atama!"},
				[]string{"something   something   darkside"},
			))
		})

		It("aligns rows to the longest row provided", func() {
			table.Add([]string{"x", "y", "z"})
			table.Print([][]string{{"something", "something", "darkside"}})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"watashi     no          atama!"},
				[]string{"x           y           z"},
				[]string{"something   something   darkside"},
			))
		})
	})
})
