package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("share-service command", func() {
	var (
		shareServiceCommand = "v3-share-service"
		serviceInstanceName = "fake-service-instance-name"
		shareToSpaceName    = "fake-space-name"
		shareToOrgName      = "fake-org-name"
	)

	Describe("help", func() {

		matchHelpMessage := SatisfyAll(
			Say("NAME:"),
			Say("share-service - Share a service instance with another space"),
			Say("USAGE:"),
			Say(`cf share-service SERVICE_INSTANCE OTHER_SPACE \[-o OTHER_ORG\]`),
			Say("OPTIONS:"),
			Say(`-o\s+Org of the other space \(Default: targeted org\)`),
			Say("SEE ALSO:"),
			Say("bind-service, service, services, unshare-service"),
		)

		When("the -h flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF(shareServiceCommand, "-h")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("the service instance name and space name is missing", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF(shareServiceCommand)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required arguments `SERVICE_INSTANCE` and `OTHER_SPACE` were not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("the space name is missing", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF(shareServiceCommand, serviceInstanceName)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required argument `OTHER_SPACE` was not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("an extra parameter is specified", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF(shareServiceCommand, serviceInstanceName, shareToSpaceName, "anotherRandomParameter")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "anotherRandomParameter"`))
				Expect(session.Out).To(SatisfyAll(
					Say(`FAILED\n\n`),
					matchHelpMessage,
				))
			})
		})

		When("an extra flag is specified", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF(shareServiceCommand, serviceInstanceName, "--anotherRandomFlag")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `anotherRandomFlag'"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, shareServiceCommand, serviceInstanceName, shareToSpaceName)
		})
	})

	Describe("command parameters are invalid", func() {
		var (
			orgName  string
			username string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName := helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)

			username, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("service instance cannot be retrieved", func() {
			It("fails with an error", func() {
				session := helpers.CF(shareServiceCommand, serviceInstanceName, shareToSpaceName)
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(SatisfyAll(
					Say("Sharing service instance %s to org %s / space %s as %s...", serviceInstanceName, orgName, shareToSpaceName, username),
					Say("FAILED"),
				))
				Expect(session.Err).To(Say("Service instance %s not found", serviceInstanceName))
			})
		})

		Context("service instance exists", func() {
			var broker *servicebrokerstub.ServiceBrokerStub

			BeforeEach(func() {
				broker = servicebrokerstub.New().Create().EnableServiceAccess()

				serviceInstanceName = helpers.NewServiceInstanceName()
				helpers.CreateManagedServiceInstance(
					broker.FirstServiceOfferingName(),
					broker.FirstServicePlanName(),
					serviceInstanceName,
				)

				shareToSpaceName = helpers.NewSpaceName()
				shareToOrgName = helpers.NewOrgName()
			})

			AfterEach(func() {
				broker.Forget()
			})

			Context("space cannot be retrieved in targeted org", func() {
				It("fails with an error", func() {
					session := helpers.CF(shareServiceCommand, serviceInstanceName, shareToSpaceName)
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(SatisfyAll(
						Say("Sharing service instance %s to org %s / space %s as %s...", serviceInstanceName, orgName, shareToSpaceName, username),
						Say("FAILED"),
					))
					Eventually(session.Err).Should(Say("Space '%s' not found.", shareToSpaceName))
				})
			})

			Context("space cannot be retrieved in specified org", func() {
				BeforeEach(func() {
					helpers.CreateOrg(shareToOrgName)
				})

				AfterEach(func() {
					helpers.QuickDeleteOrg(shareToOrgName)
				})

				It("fails with an error", func() {
					session := helpers.CF(shareServiceCommand, serviceInstanceName, shareToSpaceName, "-o", shareToOrgName)
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(SatisfyAll(
						Say("Sharing service instance %s to org %s / space %s as %s...", serviceInstanceName, shareToOrgName, shareToSpaceName, username),
						Say("FAILED"),
					))
					Eventually(session.Err).Should(Say("Space '%s' not found.", shareToSpaceName))
				})
			})

			Context("specified organization cannot be retrieved", func() {
				It("fails with an error", func() {
					session := helpers.CF(shareServiceCommand, serviceInstanceName, shareToSpaceName, "-o", shareToOrgName)
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(SatisfyAll(
						Say("Sharing service instance %s to org %s / space %s as %s...", serviceInstanceName, shareToOrgName, shareToSpaceName, username),
						Say("FAILED"),
					))
					Eventually(session.Err).Should(Say("Organization '%s' not found.", shareToOrgName))
				})
			})
		})
	})
})
