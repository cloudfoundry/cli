package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unbind-service command", func() {
	var (
		serviceInstance string
		appName         string
	)

	BeforeEach(func() {
		serviceInstance = helpers.PrefixedRandomName("si")
		appName = helpers.PrefixedRandomName("app")
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "unbind-service", "app-name", "service-name")
		})
	})

	Context("when the environment is setup correctly", func() {
		var (
			org         string
			space       string
			service     string
			servicePlan string
			broker      helpers.ServiceBroker
			domain      string
		)

		BeforeEach(func() {
			org = helpers.NewOrgName()
			space = helpers.NewSpaceName()
			service = helpers.PrefixedRandomName("SERVICE")
			servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")

			setupCF(org, space)
			domain = defaultSharedDomain()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(org)
		})

		Context("when the service is provided by a user", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-user-provided-service", serviceInstance, "-p", "{}")).Should(Exit(0))
			})

			AfterEach(func() {
				Eventually(helpers.CF("delete-service", serviceInstance, "-f")).Should(Exit(0))
			})

			Context("when the service is bound to an app", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
					})
					Eventually(helpers.CF("bind-service", appName, serviceInstance)).Should(Exit(0))
				})

				It("unbinds the service", func() {
					Eventually(helpers.CF("services")).Should(SatisfyAll(
						Exit(0),
						Say("%s.*%s", serviceInstance, appName)),
					)
					Eventually(helpers.CF("unbind-service", appName, serviceInstance)).Should(Exit(0))
					Eventually(helpers.CF("services")).Should(SatisfyAll(
						Exit(0),
						Not(Say("%s.*%s", serviceInstance, appName)),
					))
				})
			})

			Context("when the service is not bound to an app", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "--no-route")).Should(Exit(0))
					})
				})

				It("returns a warning and continues", func() {
					session := helpers.CF("unbind-service", appName, serviceInstance)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Err).Should(Say("Binding between %s and %s did not exist", serviceInstance, appName))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the service does not exist", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
					})
				})

				It("fails to unbind the service", func() {
					session := helpers.CF("unbind-service", appName, "does-not-exist")
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Service instance %s not found", "does-not-exist"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the app does not exist", func() {
				It("fails to unbind the service", func() {
					session := helpers.CF("unbind-service", "does-not-exist", serviceInstance)
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("App %s not found", "does-not-exist"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Context("when the service is provided by a broker", func() {
			BeforeEach(func() {
				broker = helpers.NewServiceBroker(helpers.NewServiceBrokerName(), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
				broker.Push()
				broker.Configure()
				broker.Create()

				Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
			})

			AfterEach(func() {
				broker.Destroy()
			})

			Context("when the service is bound to an app", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
					})
					Eventually(helpers.CF("bind-service", appName, serviceInstance)).Should(Exit(0))
				})

				It("unbinds the service", func() {
					Eventually(helpers.CF("services")).Should(SatisfyAll(
						Exit(0),
						Say("%s.*%s", serviceInstance, appName)),
					)
					Eventually(helpers.CF("unbind-service", appName, serviceInstance)).Should(Exit(0))
					Eventually(helpers.CF("services")).Should(SatisfyAll(
						Exit(0),
						Not(Say("%s.*%s", serviceInstance, appName)),
					))
				})
			})

			Context("when the service is not bound to an app", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "--no-route")).Should(Exit(0))
					})
				})

				It("returns a warning and continues", func() {
					session := helpers.CF("unbind-service", appName, serviceInstance)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Err).Should(Say("Binding between %s and %s did not exist", serviceInstance, appName))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the service does not exist", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
					})
				})

				It("fails to unbind the service", func() {
					session := helpers.CF("unbind-service", appName, serviceInstance)
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Service instance %s not found", serviceInstance))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the app does not exist", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
				})

				It("fails to unbind the service", func() {
					session := helpers.CF("unbind-service", appName, serviceInstance)
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("App %s not found", appName))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
