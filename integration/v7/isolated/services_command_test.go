package isolated

import (
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("services command", func() {
	const command = "services"

	var userName string

	BeforeEach(func() {
		userName, _ = helpers.GetCredentials()
	})

	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`services - List all service instances in the target space\n`),
			Say(`USAGE:\n`),
			Say(`cf services\n`),
			Say(`ALIAS:\n`),
			Say(`s\n`),
			Say(`OPTIONS:\n`),
			Say(`--no-apps\s+Do not retrieve bound apps information\.\n`),
			Say(`SEE ALSO:\n`),
			Say(`create-service, marketplace\n`),
		)

		When("--help flag is set", func() {
			It("displays command usage", func() {
				session := helpers.CF(command, "--help")
				Eventually(session).Should(Exit(0))
				Expect(session).To(matchHelpMessage)
			})
		})

		When("an argument is passed", func() {
			It("displays an error message with help text and exits 1", func() {
				session := helpers.CF(command, "lala")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "lala"`))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("an unknown flag is passed", func() {
			It("displays an error message with help text and exits 1", func() {
				session := helpers.CF(command, "-m")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `m'"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
	})

	Context("has no services", func() {
		BeforeEach(func() {
			helpers.LoginCF()
			helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)
		})

		It("tells the user 'no services found'", func() {
			session := helpers.CF(command)
			Eventually(session).Should(Exit(0))

			Expect(session).To(SatisfyAll(
				Say(`Getting service instances in org %s / space %s as %s...\n`, ReadOnlyOrg, ReadOnlySpace, userName),
				Say(`No service instances found\.\n`),
			))
		})
	})

	Context("has services and applications", func() {
		var (
			orgName   string
			spaceName string

			broker *servicebrokerstub.ServiceBrokerStub

			managedService1      string
			managedService2      string
			userProvidedService1 string
			userProvidedService2 string
			appName1             string
			appName2             string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)

			userProvidedService1 = helpers.PrefixedRandomName("UPS1")
			userProvidedService2 = helpers.PrefixedRandomName("UPS2")
			Eventually(helpers.CF("cups", userProvidedService1, "-p", `{"username": "admin", "password": "admin"}`)).Should(Exit(0))
			Eventually(helpers.CF("cups", userProvidedService2, "-p", `{"username": "admin", "password": "admin"}`)).Should(Exit(0))

			broker = servicebrokerstub.New().WithPlans(2).EnableServiceAccess()
			managedService1 = helpers.PrefixedRandomName("MANAGED1")
			managedService2 = helpers.PrefixedRandomName("MANAGED2")
			helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), managedService1)

			broker.Services[0].Plans[0].MaintenanceInfo = &config.MaintenanceInfo{Version: "2.0.0"}
			broker.Configure().Register()
			helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), managedService2)

			appName1 = helpers.PrefixedRandomName("APP1")
			appName2 = helpers.PrefixedRandomName("APP2")
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName1, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				Eventually(helpers.CF("push", appName2, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})
			Eventually(helpers.CF("bind-service", appName1, managedService1)).Should(Exit(0))
			Eventually(helpers.CF("bind-service", appName1, managedService2)).Should(Exit(0))
			Eventually(helpers.CF("bind-service", appName1, userProvidedService1)).Should(Exit(0))
			Eventually(helpers.CF("bind-service", appName1, userProvidedService2)).Should(Exit(0))
			Eventually(helpers.CF("bind-service", appName2, managedService2)).Should(Exit(0))
			Eventually(helpers.CF("bind-service", appName2, userProvidedService2)).Should(Exit(0))
		})

		AfterEach(func() {
			broker.Forget()
			helpers.QuickDeleteOrg(orgName)
		})

		It("displays all service information", func() {
			By("including bound apps by default", func() {
				session := helpers.CF(command)
				Eventually(session).Should(Exit(0))

				Expect(session).To(SatisfyAll(
					Say("Getting service instances in org %s / space %s as %s...", orgName, spaceName, userName),
					Say(`name\s+offering\s+plan\s+bound apps\s+last operation\s+broker\s+upgrade available\n`),
					Say(`%s\s+%s\s+%s\s+%s\s+%s\s+%s\s+%s\n`, managedService1, broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), appName1, "create succeeded", broker.Name, "yes"),
					Say(`%s\s+%s\s+%s\s+%s, %s\s+%s\s+%s\s+%s\n`, managedService2, broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), appName1, appName2, "create succeeded", broker.Name, "no"),
					Say(`%s\s+%s\s+%s\s+%s\s*\n`, userProvidedService1, "user-provided", appName1, "create succeeded"),
					Say(`%s\s+%s\s+%s, %s\s+%s\s*\n`, userProvidedService2, "user-provided", appName1, appName2, "create succeeded"),
				))
			})

			By("not showing apps when --no-apps is provided", func() {
				session := helpers.CF(command, "--no-apps")
				Eventually(session).Should(Exit(0))

				Expect(session).To(SatisfyAll(
					Say("Getting service instances in org %s / space %s as %s...", orgName, spaceName, userName),
					Say(`name\s+offering\s+plan\s+last operation\s+broker\s+upgrade available\n`),
					Say(`%s\s+%s\s+%s\s+%s\s+%s\s+%s\n`, managedService1, broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), "create succeeded", broker.Name, "yes"),
					Say(`%s\s+%s\s+%s\s+%s\s+%s\s+%s\n`, managedService2, broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), "create succeeded", broker.Name, "no"),
					Say(`%s\s+%s\s+%s\s*\n`, userProvidedService1, "user-provided", "create succeeded"),
					Say(`%s\s+%s\s+%s\s*\n`, userProvidedService2, "user-provided", "create succeeded"),
				))
			})
		})
	})

	Context("has shared service instances", func() {
		var (
			managedService, appNameOnSpaceA, appNameOnSpaceB string
		)

		BeforeEach(func() {
			orgName := helpers.NewOrgName()
			spaceA := helpers.NewSpaceName()
			spaceB := helpers.NewSpaceName()
			managedService = helpers.PrefixedRandomName("MANAGED1")
			appNameOnSpaceA = helpers.PrefixedRandomName("APP1")
			appNameOnSpaceB = helpers.PrefixedRandomName("APP1")

			helpers.SetupCF(orgName, spaceA)
			helpers.CreateOrgAndSpace(orgName, spaceB)
			broker := servicebrokerstub.New().WithPlans(2).EnableServiceAccess()
			helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), managedService)

			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appNameOnSpaceA, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})
			Eventually(helpers.CF("bind-service", appNameOnSpaceA, managedService)).Should(Exit(0))
			Eventually(helpers.CF("share-service", managedService, "-s", spaceB)).Should(Exit(0))

			helpers.TargetOrgAndSpace(orgName, spaceB)
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appNameOnSpaceB, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})
			Eventually(helpers.CF("bind-service", appNameOnSpaceB, managedService)).Should(Exit(0))
			helpers.TargetOrgAndSpace(orgName, spaceA)
		})

		It("should not output bound apps in the shared spaces", func() {
			session := helpers.CF(command)
			Eventually(session).Should(Exit(0))
			Expect(session).To(SatisfyAll(
				Say(managedService),
				Say(appNameOnSpaceA),
				Not(Say(appNameOnSpaceB)),
			))
		})
	})
})
