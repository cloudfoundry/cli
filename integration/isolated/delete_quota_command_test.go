package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-quota command", func() {
	var quotaName string

	BeforeEach(func() {
		setupCF(ReadOnlyOrg, ReadOnlySpace)
		quotaName = helpers.QuotaName()
		session := helpers.CF("create-quota", quotaName)
		Eventually(session).Should(Exit(0))
	})

	It("deletes a quota", func() {
		session := helpers.CF("delete-quota", quotaName, "-f")
		Eventually(session).Should(Say("Deleting quota %s", quotaName))
		Eventually(session).Should(Exit(0))

		session = helpers.CF("quota", quotaName)
		Eventually(session).Should(Say("%s.+not found", quotaName))
		Eventually(session).Should(Exit(1))
	})
})
