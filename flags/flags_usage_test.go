package flags_test

import (
	"github.com/cloudfoundry/cli/flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ShowUsage", func() {
	var fc flags.FlagContext

	Context("when given a flag with a longname", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("flag-a", "", "")
		})

		It("prints the longname with two hyphens", func() {
			outputs := fc.ShowUsage(0)
			Expect(outputs).To(ContainSubstring("--flag-a\n"))
		})
	})

	Context("when given a flag with a longname and usage", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("flag-a", "", "Usage for flag-a")
		})

		It("prints the longname with two hyphens followed by the usage", func() {
			outputs := fc.ShowUsage(0)
			Expect(outputs).To(ContainSubstring("--flag-a      Usage for flag-a\n"))
		})
	})

	Context("when given a flag with a longname and a shortname and usage", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("flag-a", "a", "Usage for flag-a")
		})

		It("prints the longname with two hyphens followed by the shortname followed by the usage", func() {
			outputs := fc.ShowUsage(0)
			Expect(outputs).To(ContainSubstring("--flag-a, -a      Usage for flag-a\n"))
		})
	})

	Context("when given a flag with a longname and a shortname", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("flag-a", "a", "")
		})

		It("prints the longname with two hyphens followed by the shortname with one hyphen", func() {
			outputs := fc.ShowUsage(0)
			Expect(outputs).To(ContainSubstring("--flag-a, -a\n"))
		})
	})

	Context("when given a flag with a shortname", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("", "a", "")
		})

		It("prints the shortname with one hyphen", func() {
			outputs := fc.ShowUsage(0)
			Expect(outputs).To(ContainSubstring("-a\n"))
		})
	})

	Context("when given a flag with a shortname and usage", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("", "a", "Usage for a")
		})

		It("prints the shortname with one hyphen followed by the usage", func() {
			outputs := fc.ShowUsage(0)
			Expect(outputs).To(MatchRegexp("^-a      Usage for a\n"))
		})
	})

	Context("when showing usage for multiple flags", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("flag-a", "a", "Usage for flag-a")
			fc.NewStringFlag("flag-b", "", "")
			fc.NewBoolFlag("flag-c", "c", "Usage for flag-c")
		})

		It("prints each flag on its own line", func() {
			outputs := fc.ShowUsage(0)
			Expect(outputs).To(ContainSubstring("--flag-a, -a      Usage for flag-a\n"))
			Expect(outputs).To(ContainSubstring("--flag-b\n"))
			Expect(outputs).To(ContainSubstring("--flag-c, -c      Usage for flag-c\n"))
		})
	})

	Context("when given a non-zero integer for padding", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("flag-a", "", "")
		})

		It("prefixes the flag name with the number of spaces requested", func() {
			outputs := fc.ShowUsage(5)
			Expect(outputs).To(ContainSubstring("     --flag-a\n"))
		})
	})

	Context("when showing usage for multiple flags", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("flag-a", "a", "Usage for flag-a")
			fc.NewStringFlag("flag-b", "", "Usage for flag-b")
			fc.NewBoolFlag("flag-c", "c", "Usage for flag-c")
		})

		It("aligns the text by padding string with spaces", func() {
			outputs := fc.ShowUsage(0)
			Expect(outputs).To(ContainSubstring("--flag-a, -a      Usage for flag-a"))
			Expect(outputs).To(ContainSubstring("--flag-b          Usage for flag-b"))
			Expect(outputs).To(ContainSubstring("--flag-c, -c      Usage for flag-c"))
		})
	})
})
