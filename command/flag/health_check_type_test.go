package flag_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HealthCheckType", func() {
	var healthCheck HealthCheckType

	Describe("Complete", func() {
		DescribeTable("returns list of completions",
			func(prefix string, matches []flags.Completion) {
				completions := healthCheck.Complete(prefix)
				Expect(completions).To(Equal(matches))
			},
			Entry("returns 'port' and 'process' when passed 'p'", "p",
				[]flags.Completion{{Item: "port"}, {Item: "process"}}),
			Entry("returns 'port' and 'process' when passed 'P'", "P",
				[]flags.Completion{{Item: "port"}, {Item: "process"}}),
			Entry("returns 'port' when passed 'poR'", "poR",
				[]flags.Completion{{Item: "port"}}),
			Entry("completes to 'http' when passed 'h'", "h",
				[]flags.Completion{{Item: "http"}}),
			Entry("completes to 'http', 'port', and 'process' when passed nothing", "",
				[]flags.Completion{{Item: "http"}, {Item: "port"}, {Item: "process"}}),
			Entry("completes to nothing when passed 'wut'", "wut",
				[]flags.Completion{}),
		)
	})

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			healthCheck = HealthCheckType{}
		})

		DescribeTable("downcases and sets type",
			func(settingType string, expectedType constant.HealthCheckType) {
				err := healthCheck.UnmarshalFlag(settingType)
				Expect(err).ToNot(HaveOccurred())
				Expect(healthCheck.Type).To(Equal(expectedType))
			},
			Entry("sets 'port' when passed 'port'", "port", constant.Port),
			Entry("sets 'port' when passed 'pOrt'", "pOrt", constant.Port),
			Entry("sets 'process' when passed 'process'", "process", constant.Process),
			Entry("sets 'http' when passed 'http'", "http", constant.HTTP),
		)

		When("passed anything else", func() {
			It("returns an error", func() {
				err := healthCheck.UnmarshalFlag("banana")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `HEALTH_CHECK_TYPE must be "port", "process", or "http"`,
				}))
				Expect(healthCheck.Type).To(BeEmpty())
			})
		})
	})
})
