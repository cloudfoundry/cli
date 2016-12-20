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
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("unbind-service", appName, serviceInstance)
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("unbind-service", appName, serviceInstance)
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("Not logged in. Use 'cf login' to log in."))
			})
		})

		Context("when there no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no targeted org error message", func() {
				session := helpers.CF("unbind-service", appName, serviceInstance)
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No org targeted, use 'cf target -o ORG' to target an org."))
			})
		})

		Context("when there no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no targeted space error message", func() {
				session := helpers.CF("unbind-service", appName, serviceInstance)
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No space targeted, use 'cf target -s SPACE' to target a space."))
			})
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
			space = helpers.PrefixedRandomName("SPACE")
			service = helpers.PrefixedRandomName("SERVICE")
			servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")

			setupCF(org, space)
			domain = defaultSharedDomain()
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
					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(Say("OK"))
					Expect(session.Err).To(Say("Binding between %s and %s did not exist", serviceInstance, appName))
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
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(Say("FAILED"))
					Expect(session.Err).To(Say("Service instance %s not found", "does-not-exist"))
				})
			})

			Context("when the app does not exist", func() {
				It("fails to unbind the service", func() {
					session := helpers.CF("unbind-service", "does-not-exist", serviceInstance)
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(Say("FAILED"))
					Expect(session.Err).To(Say("App %s not found", "does-not-exist"))
				})
			})
		})

		Context("when the service is provided by a broker", func() {
			BeforeEach(func() {
				broker = helpers.NewServiceBroker(helpers.PrefixedRandomName("SERVICE-BROKER"), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
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
					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(Say("OK"))
					Expect(session.Err).To(Say("Binding between %s and %s did not exist", serviceInstance, appName))
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
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(Say("FAILED"))
					Expect(session.Err).To(Say("Service instance %s not found", serviceInstance))
				})
			})

			Context("when the app does not exist", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
				})

				It("fails to unbind the service", func() {
					session := helpers.CF("unbind-service", appName, serviceInstance)
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(Say("FAILED"))
					Expect(session.Err).To(Say("App %s not found", appName))
				})
			})
		})
	})
})
