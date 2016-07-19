package flags_test

import (
	"strings"

	"code.cloudfoundry.org/cli/cf/flags"
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
			Expect(outputs).To(ContainSubstring("--flag-a"))
		})
	})

	Context("when given a flag with a longname and usage", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("flag-a", "", "Usage for flag-a")
		})

		It("prints the longname with two hyphens followed by the usage", func() {
			outputs := fc.ShowUsage(0)
			Expect(outputs).To(ContainSubstring("--flag-a      Usage for flag-a"))
		})
	})

	Context("when given a flag with a longname and a shortname and usage", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("flag-a", "a", "Usage for flag-a")
		})

		It("prints the longname with two hyphens followed by the shortname followed by the usage", func() {
			outputs := fc.ShowUsage(0)
			Expect(outputs).To(ContainSubstring("--flag-a, -a      Usage for flag-a"))
		})
	})

	Context("when given a flag with a longname and a shortname", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("flag-a", "a", "")
		})

		It("prints the longname with two hyphens followed by the shortname with one hyphen", func() {
			outputs := fc.ShowUsage(0)
			Expect(outputs).To(ContainSubstring("--flag-a, -a"))
		})
	})

	Context("when given a flag with a shortname", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("", "a", "")
		})

		It("prints the shortname with one hyphen", func() {
			outputs := fc.ShowUsage(0)
			Expect(outputs).To(ContainSubstring("-a"))
		})
	})

	Context("when given a flag with a shortname and usage", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("", "a", "Usage for a")
		})

		It("prints the shortname with one hyphen followed by the usage", func() {
			outputs := fc.ShowUsage(0)
			Expect(outputs).To(MatchRegexp("^-a      Usage for a"))
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
			Expect(outputs).To(ContainSubstring("--flag-c, -c      Usage for flag-c"))
		})
	})

	Context("when given a non-zero integer for padding", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("flag-a", "", "")
		})

		It("prefixes the flag name with the number of spaces requested", func() {
			outputs := fc.ShowUsage(5)
			Expect(outputs).To(ContainSubstring("     --flag-a"))
		})
	})

	Context("when showing usage for multiple flags", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewIntFlag("foo-a", "a", "Usage for foo-a")
			fc.NewStringFlag("someflag-b", "", "Usage for someflag-b")
			fc.NewBoolFlag("foo-c", "c", "Usage for foo-c")
			fc.NewBoolFlag("", "d", "Usage for d")
		})

		It("aligns the text by padding string with spaces", func() {
			outputs := fc.ShowUsage(0)
			Expect(outputs).To(ContainSubstring("-d                Usage for d"))
			Expect(outputs).To(ContainSubstring("--foo-a, -a       Usage for foo-a"))
			Expect(outputs).To(ContainSubstring("--foo-c, -c       Usage for foo-c"))
			Expect(outputs).To(ContainSubstring("--someflag-b      Usage for someflag-b"))
		})

		It("prints the flags in order", func() {
			for i := 0; i < 10; i++ {
				outputs := fc.ShowUsage(0)

				outputLines := strings.Split(outputs, "\n")

				Expect(outputLines).To(Equal([]string{
					"-d                Usage for d",
					"--foo-a, -a       Usage for foo-a",
					"--foo-c, -c       Usage for foo-c",
					"--someflag-b      Usage for someflag-b",
				}))
			}
		})

		Context("hidden flag", func() {
			BeforeEach(func() {
				fs := make(map[string]flags.FlagSet)
				fs["hostname"] = &flags.StringFlag{Name: "hostname", ShortName: "n", Usage: "Hostname used to identify the HTTP route", Hidden: true}
				fs["path"] = &flags.StringFlag{Name: "path", Usage: "Path used to identify the HTTP route"}
				fc = flags.NewFlagContext(fs)
			})

			It("prints the flags in order", func() {
				output := fc.ShowUsage(0)

				outputLines := strings.Split(output, "\n")

				Expect(outputLines).To(Equal([]string{
					"--path      Path used to identify the HTTP route",
				}))
			})
		})
	})
})
