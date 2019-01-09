package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = DescribeTable("error when extra arguments passed to command",
	func(cmd string) {
		helpers.SetupCF(ReadOnlyOrg, ReadOnlySpace)
		session := helpers.CF(cmd, "aaaa", "bbbbb", "ccccc", "ddddd", "eeeeee", "fffff")

		Eventually(session.Err).Should(Say(`Incorrect Usage: unexpected argument "[abcdef\s]+"`))
		Eventually(session).Should(Say("FAILED"))
		Eventually(session).Should(Exit(1))
	},
	Entry("for refactored commands", "orgs"),
	Entry("for non-refactored commands", "feature-flag"),
	Entry("for V3'd command", "push"),
)
