package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-space-quota command", func() {
	var (
		orgName   string
		spaceName string
		quotaName string
	)
	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.PrefixedRandomName("SPACE")

		setupCF(orgName, spaceName)
		quotaName = helpers.QuotaName()
		session := helpers.CF("create-space-quota", quotaName)
		Eventually(session).Should(Exit(0))
	})

	It("sets the space quota on a space", func() {
		session := helpers.CF("set-space-quota", spaceName, quotaName)
		Eventually(session).Should(Say("Assigning space quota %s to space %s", quotaName, spaceName))
		Eventually(session).Should(Exit(0))

		session = helpers.CF("space", spaceName)
		Eventually(session).Should(Say("Space Quota:\\s+%s", quotaName))
		Eventually(session).Should(Exit(0))
	})
})
