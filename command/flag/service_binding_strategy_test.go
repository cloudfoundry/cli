package flag_test

import (
	. "code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/resources"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceBindingStrategy", func() {
	var sbs ServiceBindingStrategy

	Describe("Complete", func() {
		DescribeTable("returns list of completions",
			func(prefix string, matches []flags.Completion) {
				completions := sbs.Complete(prefix)
				Expect(completions).To(Equal(matches))
			},
			Entry("returns 'single' when passed 's'", "s",
				[]flags.Completion{{Item: "single"}}),
			Entry("returns 'single' when passed 'S'", "S",
				[]flags.Completion{{Item: "single"}}),
			Entry("returns 'multiple' when passed 'm'", "m",
				[]flags.Completion{{Item: "multiple"}}),
			Entry("returns 'multiple' when passed 'M'", "M",
				[]flags.Completion{{Item: "multiple"}}),
			Entry("returns 'single' and 'multiple' when passed ''", "",
				[]flags.Completion{{Item: "single"}, {Item: "multiple"}}),
		)
	})

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			sbs = ServiceBindingStrategy{}
		})

		When("unmarshal has not been called", func() {
			It("is marked as not set", func() {
				Expect(sbs.IsSet).To(BeFalse())
			})
		})

		DescribeTable("downcases and sets strategy",
			func(input string, expected resources.BindingStrategyType) {
				err := sbs.UnmarshalFlag(input)
				Expect(err).ToNot(HaveOccurred())
				Expect(sbs.Strategy).To(Equal(expected))
				Expect(sbs.IsSet).To(BeTrue())
			},
			Entry("sets 'single' when passed 'single'", "single", resources.SingleBindingStrategy),
			Entry("sets 'single' when passed 'sInGlE'", "sInGlE", resources.SingleBindingStrategy),
			Entry("sets 'multiple' when passed 'multiple'", "multiple", resources.MultipleBindingStrategy),
			Entry("sets 'multiple' when passed 'MuLtIpLe'", "MuLtIpLe", resources.MultipleBindingStrategy),
		)

		When("passed anything else", func() {
			It("returns an error", func() {
				err := sbs.UnmarshalFlag("banana")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `STRATEGY must be "single" or "multiple"`,
				}))
				Expect(sbs.Strategy).To(BeEmpty())
			})
		})
	})
})
