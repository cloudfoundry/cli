package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("share-service command", func() {
	var (
		shareServiceCommand = "share-service"
		serviceInstanceName = "fake-service-instance-name"
		shareToSpaceName    = "fake-space-name"
		shareToOrgName      = "fake-org-name"
	)

	Describe("help", func() {

		matchHelpMessage := SatisfyAll(
			Say("NAME:"),
			Say("share-service - Share a service instance with another space"),
			Say("USAGE:"),
			Say(`cf share-service SERVICE_INSTANCE -s OTHER_SPACE \[-o OTHER_ORG\]`),
			Say("OPTIONS:"),
			Say(`-s\s+The space to share the service instance into`),
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

		When("the service instance name is missing", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF(shareServiceCommand, "-s", shareToSpaceName)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required argument `SERVICE_INSTANCE` was not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("the space name flag is missing", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF(shareServiceCommand, serviceInstanceName)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required flag `-s' was not specified"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("an extra parameter is specified", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF(shareServiceCommand, serviceInstanceName, "-s", shareToSpaceName, "anotherRandomParameter")
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
				session := helpers.CF(shareServiceCommand, serviceInstanceName, "-s", shareToSpaceName, "--anotherRandomFlag")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `anotherRandomFlag'"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(
				true,
				true,
				ReadOnlyOrg,
				shareServiceCommand, serviceInstanceName, "-s", shareToSpaceName)
		})
	})

	Context("share-service command is valid", func() {
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

		Describe("command parameters are invalid", func() {

			Context("service instance cannot be retrieved", func() {
				It("fails with an error", func() {
					session := helpers.CF(shareServiceCommand, serviceInstanceName, "-s", shareToSpaceName)
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(SatisfyAll(
						Say("Sharing service instance %s into org %s / space %s as %s...", serviceInstanceName, orgName, shareToSpaceName, username),
						Say("FAILED"),
					))
					Expect(session.Err).To(Say("Service instance '%s' not found", serviceInstanceName))
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
						session := helpers.CF(shareServiceCommand, serviceInstanceName, "-s", shareToSpaceName)
						Eventually(session).Should(Exit(1))
						Expect(session.Out).To(SatisfyAll(
							Say("Sharing service instance %s into org %s / space %s as %s...", serviceInstanceName, orgName, shareToSpaceName, username),
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
						session := helpers.CF(shareServiceCommand, serviceInstanceName, "-s", shareToSpaceName, "-o", shareToOrgName)
						Eventually(session).Should(Exit(1))
						Expect(session.Out).To(SatisfyAll(
							Say("Sharing service instance %s into org %s / space %s as %s...", serviceInstanceName, shareToOrgName, shareToSpaceName, username),
							Say("FAILED"),
						))
						Eventually(session.Err).Should(Say("Space '%s' not found.", shareToSpaceName))
					})
				})

				Context("specified organization cannot be retrieved", func() {
					It("fails with an error", func() {
						session := helpers.CF(shareServiceCommand, serviceInstanceName, "-s", shareToSpaceName, "-o", shareToOrgName)
						Eventually(session).Should(Exit(1))
						Expect(session.Out).To(SatisfyAll(
							Say("Sharing service instance %s into org %s / space %s as %s...", serviceInstanceName, shareToOrgName, shareToSpaceName, username),
							Say("FAILED"),
						))
						Eventually(session.Err).Should(Say("Organization '%s' not found.", shareToOrgName))
					})
				})
			})
		})

		Describe("when sharing the service instance succeeds", func() {
			var broker *servicebrokerstub.ServiceBrokerStub

			BeforeEach(func() {
				broker = servicebrokerstub.New().Create().EnableServiceAccess()

				serviceInstanceName = helpers.NewServiceInstanceName()
				helpers.CreateManagedServiceInstance(
					broker.FirstServiceOfferingName(),
					broker.FirstServicePlanName(),
					serviceInstanceName,
				)
			})

			AfterEach(func() {
				broker.Forget()
			})

			When("space is in targeted org", func() {
				BeforeEach(func() {
					shareToSpaceName = helpers.NewSpaceName()
					helpers.CreateSpace(shareToSpaceName)
				})

				AfterEach(func() {
					helpers.QuickDeleteSpace(shareToSpaceName)
				})

				It("shares the service to space in targeted org", func() {
					session := helpers.CF(shareServiceCommand, serviceInstanceName, "-s", shareToSpaceName)

					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(SatisfyAll(
						Say("Sharing service instance %s into org %s / space %s as %s...", serviceInstanceName, orgName, shareToSpaceName, username),
						Say("OK"),
					))

					By("validating the service is shared", func() {
						session = helpers.CF("service", serviceInstanceName)
						Eventually(session).Should(Exit(0))
						Expect(session.Out).To(SatisfyAll(
							Say(`Shared with spaces`),
							Say(`org\s+space\s+bindings`),
							Say(`%s\s+%s\s+0`, orgName, shareToSpaceName),
						))
					})
				})
			})

			When("the space to share is in specified org", func() {
				BeforeEach(func() {
					shareToOrgName = helpers.NewOrgName()
					helpers.CreateOrg(shareToOrgName)

					shareToSpaceName = helpers.NewSpaceName()
					helpers.CreateSpaceInOrg(shareToSpaceName, shareToOrgName)
				})

				AfterEach(func() {
					helpers.QuickDeleteOrg(shareToOrgName)
				})

				It("shares the service to space in specified org", func() {
					session := helpers.CF(shareServiceCommand, serviceInstanceName, "-s", shareToSpaceName, "-o", shareToOrgName)

					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(SatisfyAll(
						Say("Sharing service instance %s into org %s / space %s as %s...", serviceInstanceName, shareToOrgName, shareToSpaceName, username),
						Say("OK"),
					))

					By("validating the service is shared", func() {
						session = helpers.CF("service", serviceInstanceName)
						Eventually(session).Should(Exit(0))
						Expect(session.Out).To(SatisfyAll(
							Say(`Shared with spaces`),
							Say(`org\s+space\s+bindings`),
							Say(`%s\s+%s\s+0`, shareToOrgName, shareToSpaceName),
						))
					})
				})
			})
		})
	})
})
