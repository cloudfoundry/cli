package isolated

import (
	"regexp"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-private-domain command", func() {
	Context("Help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("create-private-domain", "DOMAINS", "Create a private domain for a specific org"))
		})

		It("displays the help information", func() {
			session := helpers.CF("create-private-domain", "--help")
			Eventually(session).Should(Say("NAME:\n"))
			Eventually(session).Should(Say(regexp.QuoteMeta("create-private-domain - Create a private domain for a specific org")))
			Eventually(session).Should(Say("USAGE:\n"))
			Eventually(session).Should(Say(regexp.QuoteMeta("cf create-private-domain ORG DOMAIN")))
			Eventually(session).Should(Say("ALIAS:\n"))
			Eventually(session).Should(Say("create-domain\n"))
			Eventually(session).Should(Say("SEE ALSO:\n"))
			Eventually(session).Should(Say("create-shared-domain, domains, share-private-domain"))
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

	When("the incorrect number of flags is provided", func() {
		It("errors and shows the help text", func() {
			session := helpers.CF("create-private-domain", domainName)
			Eventually(session.Err).Should(Say(`Incorrect Usage:`))
			Eventually(session).Should(Say("NAME:\n"))
			Eventually(session).Should(Say(regexp.QuoteMeta("create-private-domain - Create a private domain for a specific org")))
			Eventually(session).Should(Say("USAGE:\n"))
			Eventually(session).Should(Say(regexp.QuoteMeta("cf create-private-domain ORG DOMAIN")))
			Eventually(session).Should(Say("SEE ALSO:\n"))
			Eventually(session).Should(Say("create-shared-domain, domains, share-private-domain"))
			Eventually(session).Should(Exit(1))
		})

	})

	When("user is logged in", func() {
		var userName string

		BeforeEach(func() {
			userName, _ = helpers.GetCredentials()
		})

		When("org exists", func() {
			When("domain name is valid", func() {
				It("should create the private domain", func() {
					session := helpers.CF("create-private-domain", orgName, domainName)

					Eventually(session).Should(Say("Creating private domain %s for org %s as %s...", domainName, orgName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("TIP: Domain '%s' is a private domain. Run 'cf share-private-domain' to share this domain with a different org.", domainName))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("domains")
					Eventually(session).Should(Say(`%s\s+private`, domainName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("domain name is invalid", func() {
				When("invalid format", func() {
					BeforeEach(func() {
						domainName = "invalid-domain-name%*$$#)*" + helpers.RandomName()
					})

					It("should fail and return an error", func() {
						session := helpers.CF("create-private-domain", orgName, domainName)

						Eventually(session).Should(Say("Creating private domain %s for org %s as %s...", regexp.QuoteMeta(domainName), orgName, userName))
						Eventually(session.Err).Should(Say("RFC 1035"))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})

				When("preexisting domain", func() {
					var privateDomain1 helpers.Domain
					BeforeEach(func() {
						privateDomain1 = helpers.NewDomain(orgName, helpers.NewDomainName())
						privateDomain1.Create()
					})
					It("should fail and return an error", func() {
						session := helpers.CF("create-private-domain", orgName, privateDomain1.Name)

						Eventually(session.Err).Should(Say("The domain name \"%s\" is already in use", privateDomain1.Name))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})
	})

	When("user is not logged in", func() {
		BeforeEach(func() {
			helpers.LogoutCF()
		})

		It("displays an error message and fails", func() {
			session := helpers.CF("domains")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' or 'cf login --sso' to log in."))
			Eventually(session).Should(Exit(1))
		})
	})
})
