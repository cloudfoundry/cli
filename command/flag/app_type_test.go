package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppType", func() {
	var appType AppType

	Describe("Complete", func() {
		DescribeTable("returns list of completions",
			func(prefix string, matches []flags.Completion) {
				completions := appType.Complete(prefix)
				Expect(completions).To(Equal(matches))
			},

			Entry("completes to 'buildpack' when passed 'b'", "b",
				[]flags.Completion{{Item: "buildpack"}}),
			Entry("completes to 'docker' when passed 'd'", "d",
				[]flags.Completion{{Item: "docker"}}),
			Entry("completes to 'buildpack' when passed 'bU'", "bU",
				[]flags.Completion{{Item: "buildpack"}}),
			Entry("completes to 'docker' when passed 'Do'", "Do",
				[]flags.Completion{{Item: "docker"}}),
			Entry("returns 'buildpack' and 'docker' when passed nothing", "",
				[]flags.Completion{{Item: "buildpack"}, {Item: "docker"}}),
			Entry("completes to nothing when passed 'wut'", "wut",
				[]flags.Completion{}),
		)
	})
})
