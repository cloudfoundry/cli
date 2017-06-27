package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("SecurityGroupLifecycle", func() {
	var group SecurityGroupLifecycle

	Describe("Complete", func() {
		DescribeTable("returns list of completions",
			func(prefix string, matches []flags.Completion) {
				completions := group.Complete(prefix)
				Expect(completions).To(Equal(matches))
			},

			Entry("completes to 'staging' when passed 's'", "s",
				[]flags.Completion{{Item: "staging"}}),
			Entry("completes to 'running' when passed 'r'", "r",
				[]flags.Completion{{Item: "running"}}),
			Entry("completes to 'staging' when passed 'sT'", "sT",
				[]flags.Completion{{Item: "staging"}}),
			Entry("completes to 'running' when passed 'Ru'", "Ru",
				[]flags.Completion{{Item: "running"}}),
			Entry("returns 'staging' and 'running' when passed nothing", "",
				[]flags.Completion{{Item: "staging"}, {Item: "running"}}),
			Entry("completes to nothing when passed 'wut'", "wut",
				[]flags.Completion{}),
		)
	})
})
