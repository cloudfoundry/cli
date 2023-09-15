package isolated

import (
	"regexp"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("share-private-domain command", func() {
	Context("Help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("share-private-domain", "ORG ADMIN", "Share a private domain with a specific org"))
			})

			It("displays the help information", func() {
				session := helpers.CF("share-private-domain", "--help")
				Eventually(session).Should(Say("NAME:\n"))
				Eventually(session).Should(Say(regexp.QuoteMeta("share-private-domain - Share a private domain with a specific org")))
				Eventually(session).Should(Say("USAGE:\n"))
				Eventually(session).Should(Say(regexp.QuoteMeta("cf share-private-domain ORG DOMAIN")))
				Eventually(session).Should(Say("SEE ALSO:\n"))
				Eventually(session).Should(Say("create-private-domain, domains, unshare-private-domain"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	var (
		sharedWithOrgName string
		orgName           string
		spaceName         string
		domainName        string
		userName          string
	)

	BeforeEach(func() {
		sharedWithOrgName = helpers.NewOrgName()
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		domainName = helpers.NewDomainName()
	})

	When("user is logged in", func() {
		BeforeEach(func() {
			userName, _ = helpers.GetCredentials()
			helpers.SetupCF(orgName, spaceName)
			helpers.CreateOrg(sharedWithOrgName)
			domain := helpers.NewDomain(orgName, domainName)
			domain.CreatePrivate()
		})

		It("should create the shared domain", func() {
			session := helpers.CF("share-private-domain", sharedWithOrgName, domainName)

			Eventually(session).Should(Say("Sharing domain %s with org %s as %s...", domainName, sharedWithOrgName, userName))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))

			session = helpers.CF("domains")
			Eventually(session).Should(Say(`%s\s+private`, domainName))
			Eventually(session).Should(Exit(0))
		})
	})

	When("there is an error", func() {
		It("should report failure and exit with non-zero exit code", func() {
			session := helpers.CF("share-private-domain", orgName, domainName)

			Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' or 'cf login --sso' to log in."))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Exit(1))
		})
	})
})
