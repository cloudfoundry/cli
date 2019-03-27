package isolated

import (
	"fmt"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-shared-domain command", func() {
	Context("Help", func() {
		It("displays the help information", func() {
			session := helpers.CF("create-shared-domain", "--help")
			Eventually(session).Should(Say("NAME:\n"))
			Eventually(session).Should(Say(regexp.QuoteMeta("create-shared-domain - Create a domain that can be used by all orgs (admin-only)")))
			Eventually(session).Should(Say("USAGE:\n"))
			Eventually(session).Should(Say(regexp.QuoteMeta("cf create-shared-domain DOMAIN [--internal]")))
			Eventually(session).Should(Say("OPTIONS:\n"))
			Eventually(session).Should(Say(`--internal\s+Applications that use internal routes communicate directly on the container network`))
			Eventually(session).Should(Say("SEE ALSO:\n"))
			Eventually(session).Should(Say("create-private-domain, domains"))
			Eventually(session).Should(Exit(0))
		})
	})

	var (
		orgName    string
		spaceName  string
		domainName string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		helpers.SetupCF(orgName, spaceName)
		domainName = helpers.NewDomainName()
	})

	When("user is logged in as admin", func() {
		When("No optional flags are specified", func() {
			When("domain name is valid", func() {
				It("should create the shared domain", func() {
					session := helpers.CF("create-shared-domain", domainName)

					Eventually(session).Should(Say("Creating shared domain %s as admin...", domainName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("TIP: Domain '%s' is shared with all orgs. Run 'cf domains' to view available domains.", domainName))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("domains")
					Eventually(session).Should(Say(`%s\s+shared`, domainName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("domain name is invalid", func() {
				BeforeEach(func() {
					domainName = "invalid-domain-name%*$$#)*" + helpers.RandomName()
				})

				It("should fail and return an error", func() {
					session := helpers.CF("create-shared-domain", domainName)

					Eventually(session).Should(Say("Creating shared domain %s as admin...", regexp.QuoteMeta(domainName)))
					Eventually(session.Err).Should(Say("RFC 1035"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the --internal flag is specified", func() {
			It("creates a domain with internal flag", func() {
				session := helpers.CF("create-shared-domain", domainName, "--internal")

				Eventually(session).Should(Say("Creating shared domain %s as admin...", domainName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("domains")
				Eventually(session).Should(Say(`%s\s+shared\s+internal`, domainName))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("user is not logged in as admin", func() {
		var (
			username string
			password string
		)

		BeforeEach(func() {
			helpers.LoginCF()
			username, password = helpers.CreateUser()
			helpers.LoginAs(username, password)
		})

		It("should not be able to create shared domain", func() {
			session := helpers.CF("create-shared-domain", domainName)
			Eventually(session).Should(Say(fmt.Sprintf("Creating shared domain %s as %s...", domainName, username)))
			Eventually(session.Err).Should(Say("You are not authorized to perform the requested action"))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Exit(1))
		})
	})
})
