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

var _ = Describe("delete-private-domain command", func() {
	Context("Help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("delete-private-domain", "DOMAINS", "Delete a private domain"))
		})

		It("Displays command usage to output", func() {
			session := helpers.CF("delete-private-domain", "--help")

			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say(`\s+delete-private-domain - Delete a private domain`))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`\s+cf delete-private-domain DOMAIN \[-f\]`))
			Eventually(session).Should(Say("ALIAS:"))
			Eventually(session).Should(Say(`\s+delete-domain`))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say(`\s+-f\s+Force deletion without confirmation`))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say(`\s+delete-shared-domain, domains, unshare-private-domain`))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is set up correctly", func() {
		var (
			buffer     *Buffer
			orgName    string
			spaceName  string
			domainName string
			username   string
		)

		BeforeEach(func() {
			buffer = NewBuffer()
			domainName = helpers.NewDomainName()
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			username, _ = helpers.GetCredentials()
			helpers.SetupCF(orgName, spaceName)

			session := helpers.CF("create-private-domain", orgName, domainName)
			Eventually(session).Should(Exit(0))
		})

		When("the -f flag is not given", func() {
			When("the user enters 'y'", func() {
				var sharedDomainName string
				BeforeEach(func() {
					_, err := buffer.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())
					sharedDomainName = helpers.NewDomainName()
					session := helpers.CF("create-shared-domain", sharedDomainName)
					Eventually(session).Should(Exit(0))
				})
				When("the user attempts to delete-private-domain a private domain", func() {
					It("it asks for confirmation and deletes the domain", func() {
						session := helpers.CFWithStdin(buffer, "delete-private-domain", domainName)
						Eventually(session).Should(Say("Deleting the private domain will remove associated routes which will make apps with this domain unreachable."))
						Eventually(session).Should(Say(`Really delete the private domain %s\?`, domainName))
						Eventually(session).Should(Say(regexp.QuoteMeta(`Deleting private domain %s as %s...`), domainName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say("\n\nTIP: Run 'cf domains' to view available domains."))
						Eventually(session).Should(Exit(0))
					})
				})
				When("the user attempts to delete-private-domain a shared domain", func() {
					It("it fails and provides the appropriate error message", func() {
						session := helpers.CFWithStdin(buffer, "delete-private-domain", sharedDomainName)
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say(`Domain '%s' is a shared domain, not a private domain.`, sharedDomainName))

						Eventually(session).Should(Exit(1))
					})
				})
			})

			When("the user enters 'n'", func() {
				BeforeEach(func() {
					_, err := buffer.Write([]byte("n\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("it asks for confirmation and does not delete the domain", func() {
					session := helpers.CFWithStdin(buffer, "delete-private-domain", domainName)
					Eventually(session).Should(Say("Deleting the private domain will remove associated routes which will make apps with this domain unreachable."))
					Eventually(session).Should(Say(`Really delete the private domain %s\?`, domainName))
					Eventually(session).Should(Say(`'%s' has not been deleted`, domainName))
					Consistently(session).ShouldNot(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the -f flag is given", func() {
			It("it deletes the domain without asking for confirmation", func() {
				session := helpers.CFWithStdin(buffer, "delete-private-domain", domainName, "-f")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Deleting private domain %s as %s...`), domainName, username))
				Consistently(session).ShouldNot(Say("Are you sure"))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say("\n\nTIP: Run 'cf domains' to view available domains."))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("domains")
				Consistently(session).ShouldNot(Say(`%s\s+private`, domainName))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the domain doesn't exist", func() {
			It("displays OK and returns successfully", func() {
				session := helpers.CFWithStdin(buffer, "delete-private-domain", "nonexistent.com", "-f")
				Eventually(session.Err).Should(Say(`Domain 'nonexistent\.com' does not exist\.`))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})

		})
	})
})
