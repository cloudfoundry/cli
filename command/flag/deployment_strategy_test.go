package flag_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeploymentStrategy", func() {
	var strategy DeploymentStrategy

	Describe("Complete", func() {
		DescribeTable("returns list of completions",
			func(prefix string, matches []flags.Completion) {
				completions := strategy.Complete(prefix)
				Expect(completions).To(Equal(matches))
			},
			Entry("returns 'rolling' when passed 'r'", "r",
				[]flags.Completion{{Item: "rolling"}}),
		)
	})

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			strategy = DeploymentStrategy{}
		})

		DescribeTable("downcases and sets strategy",
			func(settingType string, expectedType constant.DeploymentStrategy) {
				err := strategy.UnmarshalFlag(settingType)
				Expect(err).ToNot(HaveOccurred())
				Expect(strategy.Name).To(Equal(expectedType))
			},
			Entry("sets 'rolling' when passed 'rolling'", "rolling", constant.DeploymentStrategyRolling),
			Entry("sets 'rolling' when passed 'rOlliNg'", "rOlliNg", constant.DeploymentStrategyRolling),
			Entry("sets 'rolling' when passed 'ROLLING'", "ROLLING", constant.DeploymentStrategyRolling),
		)

		When("passed anything else", func() {
			It("returns an error", func() {
				err := strategy.UnmarshalFlag("banana")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrInvalidChoice,
					Message: `STRATEGY must be "rolling" or not set`,
				}))
				Expect(strategy.Name).To(BeEmpty())
			})
		})
	})
})
