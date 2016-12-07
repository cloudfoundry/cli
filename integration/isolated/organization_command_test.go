package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("organization command", func() {
	var (
		orgName   string
		spaceName string
		quotaName string
	)
	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.PrefixedRandomName("SPACE")
		quotaName = helpers.PrefixedRandomName("INTEGRATION-QUOTA")

		setupCF(orgName, spaceName)
		session := helpers.CF("create-quota", quotaName, "-a", "1")
		Eventually(session).Should(Exit(0))
		session = helpers.CF("set-quota", orgName, quotaName)
		Eventually(session).Should(Exit(0))
	})

	It("successfully displays the organization information", func() {
		session := helpers.CF("org", orgName)
		Eventually(session).Should(Say(orgName))
		Eventually(session).Should(Say("quota:\\s+%s.+1 app instance limit", quotaName))
		Eventually(session).Should(Exit(0))
	})
})
