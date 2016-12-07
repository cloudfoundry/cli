package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Version", func() {
	DescribeTable("displays version",
		func(arg string) {
			session := helpers.CF(arg)
			Eventually(session).Should(Say("cf version [\\w0-9.+]+-[\\w0-9]+"))
			Eventually(session).Should(Exit(0))
		},

		Entry("when passed version", "version"),
		Entry("when passed -v", "-v"),
		Entry("when passed --version", "--version"),
	)
})
