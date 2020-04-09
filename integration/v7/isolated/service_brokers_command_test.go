package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("service-brokers command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("service-brokers", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("service-brokers - List service brokers"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf service-brokers"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("delete-service-broker, disable-service-access, enable-service-access"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("environment is not set up", func() {
		It("displays an error and exits 1", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "service-brokers")
		})
	})

	When("the environment is set up correctly", func() {
		var (
			username string
			session  *Session
		)

		BeforeEach(func() {
			username = helpers.LoginCF()
		})

		JustBeforeEach(func() {
			session = helpers.CF("service-brokers")
		})

		When("there is a broker", func() {
			var (
				orgName          string
				spaceName        string
				broker1, broker2 *servicebrokerstub.ServiceBrokerStub
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)
				broker1 = servicebrokerstub.Register()
				broker2 = servicebrokerstub.Register()
			})

			AfterEach(func() {
				broker1.Forget()
				broker2.Forget()
				helpers.QuickDeleteOrg(orgName)
			})

			It("prints a table of service brokers", func() {
				Eventually(session).Should(Say("Getting service brokers as %s...", username))
				Eventually(session).Should(Say(`name\s+url\s+status`))
				Eventually(session).Should(Say(`%s\s+%s\s+%s`, broker1.Name, broker1.URL, "available"))
				Eventually(session).Should(Say(`%s\s+%s\s+%s`, broker2.Name, broker2.URL, "available"))
				Eventually(session).Should(Exit(0))
			})

			When("user is not authorized to see the brokers", func() {
				var unprivilegedUsername string

				BeforeEach(func() {
					var password string
					unprivilegedUsername, password = helpers.CreateUser()
					helpers.LogoutCF()
					helpers.LoginAs(unprivilegedUsername, password)
				})

				AfterEach(func() {
					helpers.LoginCF()
					helpers.TargetOrgAndSpace(orgName, spaceName)
					helpers.DeleteUser(unprivilegedUsername)
				})

				It("says that no service brokers were found", func() {
					Eventually(session).Should(Say("Getting service brokers as"))
					Eventually(session).Should(Say("No service brokers found"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the broker was created via the V2 API", func() {
			var (
				orgName   string
				spaceName string
				broker    *servicebrokerstub.ServiceBrokerStub
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)
				broker = servicebrokerstub.Create().RegisterViaV2()
			})

			AfterEach(func() {
				broker.Forget()
				helpers.QuickDeleteOrg(orgName)
			})

			It("shows as 'available'", func() {
				Eventually(session).Should(Say("Getting service brokers as %s...", username))
				Eventually(session).Should(Say(`name\s+url\s+status`))
				Eventually(session).Should(Say(`%s\s+%s\s+%s`, broker.Name, broker.URL, "available"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
