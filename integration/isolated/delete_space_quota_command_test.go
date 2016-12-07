package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-space-quota command", func() {
	var quotaName string

	BeforeEach(func() {
		setupCF(ReadOnlyOrg, ReadOnlySpace)
		quotaName = helpers.QuotaName()
		session := helpers.CF("create-space-quota", quotaName)
		Eventually(session).Should(Exit(0))
	})

	It("deletes a space quota", func() {
		session := helpers.CF("delete-space-quota", quotaName, "-f")
		Eventually(session).Should(Say("Deleting space quota %s", quotaName))
		Eventually(session).Should(Exit(0))

		session = helpers.CF("space-quota", quotaName)
		Eventually(session).Should(Exit(1))
	})
})
