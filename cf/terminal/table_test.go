package terminal_test

import (
	"bytes"
	"strings"

	. "code.cloudfoundry.org/cli/cf/terminal"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Table", func() {
	var (
		outputs *bytes.Buffer
		table   *Table
	)

	BeforeEach(func() {
		outputs = &bytes.Buffer{}
		table = NewTable([]string{"watashi", "no", "atama!"})
	})

	It("prints the header", func() {
		table.PrintTo(outputs)
		s := strings.Split(outputs.String(), "\n")
		Expect(s).To(ContainSubstrings(
			[]string{"watashi", "no", "atama!"},
		))
	})

	Describe("REGRESSION: #117404629, having a space in one of the middle headers", func() {
		BeforeEach(func() {
			outputs = &bytes.Buffer{}
			table = NewTable([]string{"watashi", "no ", "atama!"})
		})

		It("prints the table without error", func() {
			err := table.PrintTo(outputs)
			Expect(err).NotTo(HaveOccurred())

			s := strings.Split(outputs.String(), "\n")
			Expect(s).To(ContainSubstrings(
				[]string{"watashi", "no", "atama!"},
			))
		})

		It("prints the table with the extra whitespace from the header stripped", func() {
			err := table.PrintTo(outputs)
			Expect(err).NotTo(HaveOccurred())

			s := strings.Split(outputs.String(), "\n")
			Expect(s).To(ContainSubstrings(
				[]string{"watashi   no   atama!"},
			))
		})
	})

	It("prints format string literals as strings", func() {
		table.Add("cloak %s", "and", "dagger")
		table.PrintTo(outputs)
		s := strings.Split(outputs.String(), "\n")

		Expect(s).To(ContainSubstrings(
			[]string{"cloak %s", "and", "dagger"},
		))
	})

	It("prints all the rows you give it", func() {
		table.Add("something", "and", "nothing")
		table.PrintTo(outputs)
		s := strings.Split(outputs.String(), "\n")
		Expect(s).To(ContainSubstrings(
			[]string{"something", "and", "nothing"},
		))
	})

	Describe("adding rows to be printed later", func() {
		It("prints them when you call Print()", func() {
			table.Add("a", "b", "c")
			table.Add("passed", "to", "print")
			table.PrintTo(outputs)
			s := strings.Split(outputs.String(), "\n")

			Expect(s).To(ContainSubstrings(
				[]string{"a", "b", "c"},
			))
		})

		It("flushes previously added rows and then outputs passed rows", func() {
			table.Add("a", "b", "c")
			table.Add("passed", "to", "print")
			table.PrintTo(outputs)
			s := strings.Split(outputs.String(), "\n")

			Expect(s).To(ContainSubstrings(
				[]string{"watashi", "no", "atama!"},
				[]string{"a", "b", "c"},
				[]string{"passed", "to", "print"},
			))
		})
	})

	It("prints a newline for the headers, and nothing for rows", func() {
		table = NewTable([]string{})
		table.PrintTo(outputs)

		Expect(outputs.String()).To(Equal("\n"))
	})

	It("prints nothing with suppressed headers and no rows", func() {
		table.NoHeaders()
		table.PrintTo(outputs)

		Expect(outputs.String()).To(BeEmpty())
	})

	It("does not print the header when suppressed", func() {
		table.NoHeaders()
		table.Add("cloak", "and", "dagger")
		table.PrintTo(outputs)
		s := strings.Split(outputs.String(), "\n")

		Expect(s).To(Not(ContainSubstrings(
			[]string{"watashi", "no", "atama!"},
		)))
		Expect(s).To(ContainSubstrings(
			[]string{"cloak", "and", "dagger"},
		))
	})

	It("prints cell strings as specified by column transformers", func() {
		table.Add("cloak", "and", "dagger")
		table.SetTransformer(0, func(s string) string {
			return "<<" + s + ">>"
		})
		table.PrintTo(outputs)
		s := strings.Split(outputs.String(), "\n")

		Expect(s).To(ContainSubstrings(
			[]string{"<<cloak>>", "and", "dagger"},
		))
		Expect(s).To(Not(ContainSubstrings(
			[]string{"<<watashi>>"},
		)))
	})

	It("prints no more columns than headers", func() {
		table.Add("something", "and", "nothing", "ignored")
		table.PrintTo(outputs)
		s := strings.Split(outputs.String(), "\n")

		Expect(s).To(Not(ContainSubstrings(
			[]string{"ignored"},
		)))
	})

	It("avoids printing trailing whitespace for empty columns", func() {
		table.Add("something", "and")
		table.PrintTo(outputs)
		s := strings.Split(outputs.String(), "\n")

		Expect(s).To(ContainSubstrings(
			[]string{"watashi     no    atama!"},
			[]string{"something   and"},
		))
	})

	It("avoids printing trailing whitespace for whitespace columns", func() {
		table.Add("something", "    ")
		table.PrintTo(outputs)
		s := strings.Split(outputs.String(), "\n")

		Expect(s).To(ContainSubstrings(
			[]string{"watashi     no   atama!"},
			[]string{"something"},
		))
	})

	It("even avoids printing trailing whitespace for multi-line cells", func() {
		table.Add("a", "b\nd", "\nc")
		table.PrintTo(outputs)
		s := strings.Split(outputs.String(), "\n")

		Expect(s).To(ContainSubstrings(
			[]string{"watashi   no   atama!"},
			[]string{"a         b"},
			[]string{"          d    c"},
		))
	})

	It("prints multi-line cells on separate physical lines", func() {
		table.Add("a", "b\nd", "c")
		table.PrintTo(outputs)
		s := strings.Split(outputs.String(), "\n")

		Expect(s).To(ContainSubstrings(
			[]string{"watashi   no   atama!"},
			[]string{"a         b    c"},
			[]string{"          d"},
		))
	})

	Describe("aligning columns", func() {
		It("aligns rows to the header when the header is longest", func() {
			table.Add("a", "b", "c")
			table.PrintTo(outputs)
			s := strings.Split(outputs.String(), "\n")

			Expect(s).To(ContainSubstrings(
				[]string{"watashi   no   atama!"},
				[]string{"a         b    c"},
			))
		})

		It("aligns rows to the longest row provided", func() {
			table.Add("x", "y", "z")
			table.Add("something", "something", "darkside")
			table.PrintTo(outputs)
			s := strings.Split(outputs.String(), "\n")

			Expect(s).To(ContainSubstrings(
				[]string{"watashi     no          atama!"},
				[]string{"x           y           z"},
				[]string{"something   something   darkside"},
			))
		})

		It("aligns rows to the longest row provided when there are multibyte characters present", func() {
			table.Add("x", "ÿ", "z")
			table.Add("something", "something", "darkside")
			table.PrintTo(outputs)
			s := strings.Split(outputs.String(), "\n")

			Expect(s).To(ContainSubstrings(
				[]string{"watashi     no          atama!"},
				[]string{"x           ÿ           z"},
				[]string{"something   something   darkside"},
			))
		})

		It("supports multi-byte Japanese runes", func() {
			table = NewTable([]string{"", "", "", "", "", ""})
			table.Add("名前", "要求された状態", "インスタンス", "メモリー", "ディスク", "URL")
			table.Add("app-name", "stopped", "0/1", "1G", "1G", "app-name.example.com")
			table.PrintTo(outputs)
			s := strings.Split(outputs.String(), "\n")

			Expect(s).To(ContainSubstrings(
				[]string{"名前       要求された状態   インスタンス   メモリー   ディスク   URL"},
				[]string{"app-name   stopped          0/1            1G         1G         app-name.example.com"},
			))
		})

		It("supports multi-byte French runes", func() {
			table = NewTable([]string{"", "", "", "", "", ""})
			table.Add("nom", "état demandé", "instances", "mémoire", "disque", "adresses URL")
			table.Add("app-name", "stopped", "0/1", "1G", "1G", "app-name.example.com")
			table.PrintTo(outputs)
			s := strings.Split(outputs.String(), "\n")

			Expect(s).To(ContainSubstrings(
				[]string{"nom        état demandé   instances   mémoire   disque   adresses URL"},
				[]string{"app-name   stopped        0/1         1G        1G       app-name.example.com"},
			))
		})
	})
})
