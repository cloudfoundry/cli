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
		table.Add("cloak %s", "and", "dagger")
		table.Print()

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"cloak %s", "and", "dagger"},
		))
	})

	It("prints all the rows you give it", func() {
		table.Add("something", "and", "nothing")
		table.Print()
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"something", "and", "nothing"},
		))
	})

	Describe("adding rows to be printed later", func() {
		It("prints them when you call Print()", func() {
			table.Add("a", "b", "c")
			table.Add("passed", "to", "print")
			table.Print()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"a", "b", "c"},
			))
		})

		It("flushes previously added rows and then outputs passed rows", func() {
			table.Add("a", "b", "c")
			table.Add("passed", "to", "print")
			table.Print()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"watashi", "no", "atama!"},
				[]string{"a", "b", "c"},
				[]string{"passed", "to", "print"},
			))
		})

		It("flushes the buffer of rows when you call print", func() {
			table.Add("a", "b", "c")
			table.Add("passed", "to", "print")
			table.Print()
			ui.ClearOutputs()

			table.Print()
			Expect(ui.Outputs).To(BeEmpty())
		})
	})

	Describe("aligning columns", func() {
		It("aligns rows to the header when the header is longest", func() {
			table.Add("a", "b", "c")
			table.Print()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"watashi   no   atama!"},
				[]string{"a         b    c"},
			))
		})

		It("aligns rows to the longest row provided", func() {
			table.Add("x", "y", "z")
			table.Add("something", "something", "darkside")
			table.Print()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"watashi     no          atama!"},
				[]string{"x           y           z"},
				[]string{"something   something   darkside"},
			))
		})

		It("aligns rows to the longest row provided when there are multibyte characters present", func() {
			table.Add("x", "ÿ", "z")
			table.Add("something", "something", "darkside")
			table.Print()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"watashi     no          atama!"},
				[]string{"x           ÿ           z"},
				[]string{"something   something   darkside"},
			))
		})

		It("supports multi-byte Japanese runes", func() {
			table = NewTable(ui, []string{"", "", "", "", "", ""})
			table.Add("名前", "要求された状態", "インスタンス", "メモリー", "ディスク", "URL")
			table.Add("app-name", "stopped", "0/1", "1G", "1G", "app-name.example.com")
			table.Print()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"名前       要求された状態   インスタンス   メモリー   ディスク   URL"},
				[]string{"app-name   stopped          0/1            1G         1G         app-name.example.com"},
			))
		})

		It("supports multi-byte French runes", func() {
			table = NewTable(ui, []string{"", "", "", "", "", ""})
			table.Add("nom", "état demandé", "instances", "mémoire", "disque", "adresses URL")
			table.Add("app-name", "stopped", "0/1", "1G", "1G", "app-name.example.com")
			table.Print()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"nom        état demandé   instances   mémoire   disque   adresses URL"},
				[]string{"app-name   stopped        0/1         1G        1G       app-name.example.com"},
			))
		})
	})
})
