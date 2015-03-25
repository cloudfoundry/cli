package flags_test

import (
	. "github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Flags", func() {

	Describe("FlagContext", func() {

		Describe("Parsing and retriving values", func() {

			var (
				fCtx       FlagContext
				cmdFlagMap map[string]FlagSet
			)

			BeforeEach(func() {
				cmdFlagMap = make(map[string]FlagSet)
				cmdFlagMap["name"] = &cliFlags.StringFlag{Name: "name", Value: "", Usage: "test string flag"}
				cmdFlagMap["skip"] = &cliFlags.BoolFlag{Name: "skip", Value: false, Usage: "test bool flag"}
				cmdFlagMap["instance"] = &cliFlags.IntFlag{Name: "instance", Value: 0, Usage: "test int flag"}
				cmdFlagMap["skip2"] = &cliFlags.BoolFlag{Name: "skip2", Value: false, Usage: "test bool flag"}

				fCtx = NewFlagContext(cmdFlagMap)
			})

			It("accepts flags with either single '-' or double '-' ", func() {
				err := fCtx.Parse("--name", "blue")
				Ω(err).ToNot(HaveOccurred())

				err = fCtx.Parse("-name", "")
				Ω(err).ToNot(HaveOccurred())
			})

			It("checks if a flag is defined in the FlagContext", func() {
				err := fCtx.Parse("-not_defined", "")
				Ω(err).To(HaveOccurred())

				err = fCtx.Parse("-name", "blue")
				Ω(err).ToNot(HaveOccurred())

				err = fCtx.Parse("--skip", "")
				Ω(err).ToNot(HaveOccurred())
			})

			It("sets Bool(<flag>) to return value if bool flag is provided with value true/false", func() {
				err := fCtx.Parse("--skip=false", "-skip2", "true", "-name", "johndoe")
				Ω(err).ToNot(HaveOccurred())

				Ω(fCtx.Bool("skip")).To(Equal(false), "skip should be false")
				Ω(fCtx.Bool("skip2")).To(Equal(true), "skip2 should be true")
				Ω(fCtx.Bool("name")).To(Equal(false), "name should be false")
				Ω(fCtx.Bool("non-exisit-flag")).To(Equal(false))
			})

			It("sets Bool(<flag>) to return true if bool flag is provided with invalid value", func() {
				err := fCtx.Parse("--skip=Not_Valid", "skip2", "FALSE", "-name", "johndoe")
				Ω(err).ToNot(HaveOccurred())

				Ω(fCtx.Bool("skip")).To(Equal(true), "skip should be true")
				Ω(fCtx.Bool("skip2")).To(Equal(false), "skip2 should be false")
			})

			It("sets Bool(<flag>) to return true when a bool flag is provided without value", func() {
				err := fCtx.Parse("--skip", "-name", "johndoe")
				Ω(err).ToNot(HaveOccurred())

				Ω(fCtx.Bool("skip")).To(Equal(true), "skip should be true")
				Ω(fCtx.Bool("name")).To(Equal(false), "name should be false")
				Ω(fCtx.Bool("non-exisit-flag")).To(Equal(false))
			})

			It("sets String(<flag>) to return provided value when a string flag is provided", func() {
				err := fCtx.Parse("--skip", "-name", "doe")
				Ω(err).ToNot(HaveOccurred())

				Ω(fCtx.String("name")).To(Equal("doe"))
				Ω(fCtx.Bool("skip")).To(Equal(true), "skip should be true")
			})

			It("errors when a non-boolean flag is provided without a value", func() {
				err := fCtx.Parse("-name")
				Ω(err).To(HaveOccurred())
				Ω(err.Error()).To(ContainSubstring("No value provided for flag"))
				Ω(fCtx.String("name")).To(Equal(""))
			})

			It("sets Int(<flag>) to return provided value when a int flag is provided", func() {
				err := fCtx.Parse("--instance", "10")
				Ω(err).ToNot(HaveOccurred())

				Ω(fCtx.Int("instance")).To(Equal(10))
				Ω(fCtx.IsSet("instance")).To(Equal(true))

				Ω(fCtx.Int("non-exist-flag")).To(Equal(0))
				Ω(fCtx.IsSet("non-exist-flag")).To(Equal(false))
			})

			It("returns any non-flag arguments in Args()", func() {
				err := fCtx.Parse("Arg-1", "--instance", "10", "--skip", "Arg-2")
				Ω(err).ToNot(HaveOccurred())

				Ω(len(fCtx.Args())).To(Equal(2))
				Ω(fCtx.Args()[0]).To(Equal("Arg-1"))
				Ω(fCtx.Args()[1]).To(Equal("Arg-2"))
			})

			It("accepts flag/value in the forms of '-flag=value' and '-flag value'", func() {
				err := fCtx.Parse("-instance", "10", "--name=foo", "--skip", "Arg-1")
				Ω(err).ToNot(HaveOccurred())

				Ω(fCtx.IsSet("instance")).To(Equal(true))
				Ω(fCtx.Int("instance")).To(Equal(10))

				Ω(fCtx.IsSet("name")).To(Equal(true))
				Ω(fCtx.String("name")).To(Equal("foo"))

				Ω(fCtx.IsSet("skip")).To(Equal(true))

				Ω(len(fCtx.Args())).To(Equal(1))
				Ω(fCtx.Args()[0]).To(Equal("Arg-1"))
			})

		})

	})
})
