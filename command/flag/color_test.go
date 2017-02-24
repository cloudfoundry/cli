package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Color", func() {
	var color Color

	Describe("Complete", func() {
		DescribeTable("returns list of completions",
			func(prefix string, matches []flags.Completion) {
				completions := color.Complete(prefix)
				Expect(completions).To(Equal(matches))
			},

			Entry("completes to 'true' when passed 't'", "t",
				[]flags.Completion{{Item: "true"}}),
			Entry("completes to 'false' when passed 'f'", "f",
				[]flags.Completion{{Item: "false"}}),
			Entry("completes to 'true' when passed 'tR'", "tR",
				[]flags.Completion{{Item: "true"}}),
			Entry("completes to 'false' when passed 'Fa'", "Fa",
				[]flags.Completion{{Item: "false"}}),
			Entry("returns 'true' and 'false' when passed nothing", "",
				[]flags.Completion{{Item: "true"}, {Item: "false"}}),
			Entry("completes to nothing when passed 'wut'", "wut",
				[]flags.Completion{}),
		)
	})

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			color = Color{}
		})

		It("accepts true", func() {
			err := color.UnmarshalFlag("true")
			Expect(err).ToNot(HaveOccurred())
			Expect(color.Color).To(BeTrue())
		})

		It("accepts false", func() {
			err := color.UnmarshalFlag("FalsE")
			Expect(err).ToNot(HaveOccurred())
			Expect(color.Color).To(BeFalse())
		})

		It("errors on anything else", func() {
			err := color.UnmarshalFlag("I AM A BANANANANANANANANAE")
			Expect(err).To(MatchError(&flags.Error{
				Type:    flags.ErrRequired,
				Message: `COLOR must be "true" or "false"`,
			}))
		})
	})
})
