package flags_test

import (
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Flags", func() {
	Describe("FlagContext", func() {
		Describe("Parsing and retriving values", func() {
			var (
				fCtx       flags.FlagContext
				cmdFlagMap map[string]flags.FlagSet
			)

			BeforeEach(func() {
				cmdFlagMap = make(map[string]flags.FlagSet)

				cmdFlagMap["name"] = &cliFlags.StringFlag{Name: "name", ShortName: "n", Usage: "test string flag"}
				cmdFlagMap["skip"] = &cliFlags.BoolFlag{Name: "skip", Usage: "test bool flag"}
				cmdFlagMap["instance"] = &cliFlags.IntFlag{Name: "instance", Usage: "test int flag"}
				cmdFlagMap["float"] = &cliFlags.Float64Flag{Name: "float", Usage: "test float64 flag"}
				cmdFlagMap["skip2"] = &cliFlags.BoolFlag{Name: "skip2", Usage: "test bool flag"}
				cmdFlagMap["slice"] = &cliFlags.StringSliceFlag{Name: "slice", Usage: "test stringSlice flag"}

				fCtx = flags.NewFlagContext(cmdFlagMap)
			})

			It("accepts flags with either single '-' or double '-' ", func() {
				err := fCtx.Parse("--name", "blue")
				Ω(err).ToNot(HaveOccurred())

				err = fCtx.Parse("-name", "")
				Ω(err).ToNot(HaveOccurred())
			})

			It("sets a flag with it's full name", func() {
				err := fCtx.Parse("-name", "blue")
				Ω(err).ToNot(HaveOccurred())
				Ω(fCtx.IsSet("name")).To(BeTrue())
				Ω(fCtx.IsSet("n")).To(BeTrue())
				Ω(fCtx.String("name")).To(Equal("blue"))
				Ω(fCtx.String("n")).To(Equal("blue"))
			})

			It("sets a flag with it's short name", func() {
				err := fCtx.Parse("-n", "red")
				Ω(err).ToNot(HaveOccurred())
				Ω(fCtx.IsSet("name")).To(BeTrue())
				Ω(fCtx.IsSet("n")).To(BeTrue())
				Ω(fCtx.String("name")).To(Equal("red"))
				Ω(fCtx.String("n")).To(Equal("red"))
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
				err := fCtx.Parse("--skip=false", "-skip2", "true", "-name=johndoe")
				Ω(err).ToNot(HaveOccurred())

				Ω(len(fCtx.Args())).To(Equal(0), "Length of Args() should be 0")
				Ω(fCtx.Bool("skip")).To(Equal(false), "skip should be false")
				Ω(fCtx.Bool("skip2")).To(Equal(true), "skip2 should be true")
				Ω(fCtx.Bool("name")).To(Equal(false), "name should be false")
				Ω(fCtx.String("name")).To(Equal("johndoe"), "name should be johndoe")
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

			It("sets StringSlice(<flag>) to return provided value when a stringSlice flag is provided", func() {
				err := fCtx.Parse("-slice", "value1", "-slice", "value2")
				Ω(err).ToNot(HaveOccurred())

				Ω(fCtx.StringSlice("slice")[0]).To(Equal("value1"), "slice[0] should be 'value1'")
				Ω(fCtx.StringSlice("slice")[1]).To(Equal("value2"), "slice[1] should be 'value2'")
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

			It("sets Float64(<flag>) to return provided value when a float64 flag is provided", func() {
				err := fCtx.Parse("-float", "10.5")
				Ω(err).ToNot(HaveOccurred())

				Ω(fCtx.Float64("float")).To(Equal(10.5))
				Ω(fCtx.IsSet("float")).To(Equal(true))

				Ω(fCtx.Float64("non-exist-flag")).To(Equal(float64(0)))
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

			Context("Default Flag Value", func() {

				BeforeEach(func() {
					cmdFlagMap = make(map[string]flags.FlagSet)

					cmdFlagMap["defaultStringFlag"] = &cliFlags.StringFlag{Name: "defaultStringFlag", Value: "Set by default"}
					cmdFlagMap["defaultBoolFlag"] = &cliFlags.BoolFlag{Name: "defaultBoolFlag", Value: true}
					cmdFlagMap["defaultIntFlag"] = &cliFlags.IntFlag{Name: "defaultIntFlag", Value: 100}
					cmdFlagMap["defaultStringAryFlag"] = &cliFlags.StringSliceFlag{Name: "defaultStringAryFlag", Value: []string{"abc", "def"}}
					cmdFlagMap["defaultFloat64Flag"] = &cliFlags.Float64Flag{Name: "defaultFloat64Flag", Value: 100.5}
					cmdFlagMap["noDefaultStringFlag"] = &cliFlags.StringFlag{Name: "noDefaultStringFlag"}

					fCtx = flags.NewFlagContext(cmdFlagMap)
				})

				It("sets flag with default value if 'Value' is provided", func() {
					err := fCtx.Parse()
					Ω(err).ToNot(HaveOccurred())

					Ω(fCtx.String("defaultStringFlag")).To(Equal("Set by default"))
					Ω(fCtx.IsSet("defaultStringFlag")).To(BeTrue())

					Ω(fCtx.Bool("defaultBoolFlag")).To(BeTrue())
					Ω(fCtx.IsSet("defaultBoolFlag")).To(BeTrue())

					Ω(fCtx.Int("defaultIntFlag")).To(Equal(100))
					Ω(fCtx.IsSet("defaultIntFlag")).To(BeTrue())

					Ω(fCtx.Float64("defaultFloat64Flag")).To(Equal(100.5))
					Ω(fCtx.IsSet("defaultFloat64Flag")).To(BeTrue())

					Ω(fCtx.StringSlice("defaultStringAryFlag")).To(Equal([]string{"abc", "def"}))
					Ω(fCtx.IsSet("defaultStringAryFlag")).To(BeTrue())

					Ω(fCtx.String("noDefaultStringFlag")).To(Equal(""))
					Ω(fCtx.IsSet("noDefaultStringFlag")).To(BeFalse())
				})

				It("overrides default value if argument is provided, except StringSlice Flag", func() {
					err := fCtx.Parse("-defaultStringFlag=foo", "-defaultBoolFlag=false", "-defaultIntFlag=200", "-defaultStringAryFlag=foo", "-defaultStringAryFlag=bar", "-noDefaultStringFlag=baz")
					Ω(err).ToNot(HaveOccurred())

					Ω(fCtx.String("defaultStringFlag")).To(Equal("foo"))
					Ω(fCtx.IsSet("defaultStringFlag")).To(BeTrue())

					Ω(fCtx.Bool("defaultBoolFlag")).To(BeFalse())
					Ω(fCtx.IsSet("defaultBoolFlag")).To(BeTrue())

					Ω(fCtx.Int("defaultIntFlag")).To(Equal(200))
					Ω(fCtx.IsSet("defaultIntFlag")).To(BeTrue())

					Ω(fCtx.String("noDefaultStringFlag")).To(Equal("baz"))
					Ω(fCtx.IsSet("noDefaultStringFlag")).To(BeTrue())
				})

				It("appends argument value to StringSliceFlag to the default values", func() {
					err := fCtx.Parse("-defaultStringAryFlag=foo", "-defaultStringAryFlag=bar")
					Ω(err).ToNot(HaveOccurred())

					Ω(fCtx.StringSlice("defaultStringAryFlag")).To(Equal([]string{"abc", "def", "foo", "bar"}))
					Ω(fCtx.IsSet("defaultStringAryFlag")).To(BeTrue())
				})

			})

			Context("SkipFlagParsing", func() {
				It("skips flag parsing and treats all arguments as values", func() {
					fCtx.SkipFlagParsing(true)
					err := fCtx.Parse("value1", "--name", "foo")
					Ω(err).ToNot(HaveOccurred())

					Ω(fCtx.IsSet("name")).To(Equal(false))

					Ω(len(fCtx.Args())).To(Equal(3))
					Ω(fCtx.Args()[0]).To(Equal("value1"))
					Ω(fCtx.Args()[1]).To(Equal("--name"))
					Ω(fCtx.Args()[2]).To(Equal("foo"))
				})
			})

		})

	})
})
