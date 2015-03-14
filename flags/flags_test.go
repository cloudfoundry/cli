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

			It("sets Bool(<flag>) to return true when a bool flag is provided", func() {
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

			It("sets Int(<flag>) to return provided value when a int flag is provided", func() {
				err := fCtx.Parse("--instance", "10")
				Ω(err).ToNot(HaveOccurred())

				Ω(fCtx.Int("instance")).To(Equal(10))
				Ω(fCtx.IsSet("instance")).To(Equal(true))

				Ω(fCtx.Int("non-exist-flag")).To(Equal(0))
				Ω(fCtx.IsSet("non-exist-flag")).To(Equal(false))
			})

			// It("sets flags regardless of order", func() {
			// 	err := fCtx.Parse("--skip", "-name", "doe", "-instance", "10")
			// 	Ω(err).ToNot(HaveOccurred())
			// 	Ω(fCtx.String("name")).To(Equal("doe"))
			// 	Ω(fCtx.Bool("skip")).To(Equal(true), "skip should be true")
			// 	Ω(fCtx.Int("instance")).To(Equal(10))

			// 	fCtx = NewFlagContext(cmdFlagMap)
			// 	err = fCtx.Parse("-instance", "20", "APP-NAME", "APP-ROUTE", "-name", "smith", "--skip")
			// 	Ω(err).ToNot(HaveOccurred())
			// 	Ω(fCtx.String("name")).To(Equal("smith"))
			// 	Ω(fCtx.Bool("skip")).To(Equal(true), "skip should be true")
			// 	Ω(fCtx.Int("instance")).To(Equal(20))
			// })
		})

	})
})
