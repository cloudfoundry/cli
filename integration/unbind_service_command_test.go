package integration

import (
	. "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unbind-service command", func() {
	var (
		org             string
		space           string
		service         string
		servicePlan     string
		serviceInstance string
		appName         string
		broker          ServiceBroker
	)

	BeforeEach(func() {
		Skip("until #129631341")
		org = PrefixedRandomName("ORG")
		space = PrefixedRandomName("SPACE")
		service = PrefixedRandomName("SERVICE")
		servicePlan = PrefixedRandomName("SERVICE-PLAN")
		serviceInstance = PrefixedRandomName("si")
		appName = PrefixedRandomName("app")

		setupCF(org, space)
	})

	AfterEach(func() {
		setAPI()
		loginCF()
		Eventually(CF("delete-org", "-f", org), CFLongTimeout).Should(Exit(0))
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				unsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := CF("unbind-service", appName, serviceInstance)
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				logoutCF()
			})

			It("fails with not logged in message", func() {
				session := CF("unbind-service", appName, serviceInstance)
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("Not logged in. Use 'cf login' to log in."))
			})
		})

		Context("when there no space set", func() {
			BeforeEach(func() {
				logoutCF()
				loginCF()
			})

			It("fails with no targeted space error message", func() {
				session := CF("unbind-service", appName, serviceInstance)
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No org targeted, use 'cf target -o ORG' to target an org."))
			})
		})
	})

	Context("when the environment is setup correctly", func() {
		Context("when the service is provided by a user", func() {
			BeforeEach(func() {
				Eventually(CF("create-user-provided-service", serviceInstance, "-p", "{}")).Should(Exit(0))
			})

			AfterEach(func() {
				Eventually(CF("delete-service", serviceInstance, "-f")).Should(Exit(0))
			})

			Context("when the service is bound to an app", func() {
				BeforeEach(func() {
					WithSimpleApp(func(appDir string) {
						Eventually(CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route"), CFLongTimeout).Should(Exit(0))
					})
					Eventually(CF("bind-service", appName, serviceInstance)).Should(Exit(0))
				})

				It("unbinds the service", func() {
					Eventually(CF("services")).Should(SatisfyAll(
						Exit(0),
						Say("%s.*%s", serviceInstance, appName)),
					)
					Eventually(CF("unbind-service", appName, serviceInstance), CFLongTimeout).Should(Exit(0))
					Eventually(CF("services")).Should(SatisfyAll(
						Exit(0),
						Not(Say("%s.*%s", serviceInstance, appName)),
					))
				})
			})

			Context("when the service is not bound to an app", func() {
				BeforeEach(func() {
					WithSimpleApp(func(appDir string) {
						Eventually(CF("push", appName, "--no-start", "-p", appDir, "--no-route"), CFLongTimeout).Should(Exit(0))
					})
				})

				It("returns a warning and continues", func() {
					session := CF("unbind-service", appName, serviceInstance)
					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(Say("OK"))
					Expect(session.Err).To(Say("Binding between %s and %s did not exist", serviceInstance, appName))
				})
			})

			Context("when the service does not exist", func() {
				BeforeEach(func() {
					WithSimpleApp(func(appDir string) {
						Eventually(CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route"), CFLongTimeout).Should(Exit(0))
					})
				})

				It("fails to unbind the service", func() {
					session := CF("unbind-service", appName, "does-not-exist")
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(Say("FAILED"))
					Expect(session.Err).To(Say("Service instance %s not found", "does-not-exist"))
				})
			})

			Context("when the app does not exist", func() {
				It("fails to unbind the service", func() {
					session := CF("unbind-service", "does-not-exist", serviceInstance)
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(Say("FAILED"))
					Expect(session.Err).To(Say("App %s not found", "does-not-exist"))
				})
			})
		})

		Context("when the service is provided by a broker", func() {
			BeforeEach(func() {
				broker = NewServiceBroker(PrefixedRandomName("SERVICE-BROKER"), NewAssets().ServiceBroker, "bosh-lite.com", service, servicePlan)
				broker.Push()
				broker.Configure()
				broker.Create()

				Eventually(CF("enable-service-access", service)).Should(Exit(0))
			})

			AfterEach(func() {
				broker.Destroy()
			})

			Context("when the service is bound to an app", func() {
				BeforeEach(func() {
					Eventually(CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
					WithSimpleApp(func(appDir string) {
						Eventually(CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route"), CFLongTimeout).Should(Exit(0))
					})
					Eventually(CF("bind-service", appName, serviceInstance)).Should(Exit(0))
				})

				It("unbinds the service", func() {
					Eventually(CF("services")).Should(SatisfyAll(
						Exit(0),
						Say("%s.*%s", serviceInstance, appName)),
					)
					Eventually(CF("unbind-service", appName, serviceInstance), CFLongTimeout).Should(Exit(0))
					Eventually(CF("services")).Should(SatisfyAll(
						Exit(0),
						Not(Say("%s.*%s", serviceInstance, appName)),
					))
				})
			})

			Context("when the service is not bound to an app", func() {
				BeforeEach(func() {
					Eventually(CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
					WithSimpleApp(func(appDir string) {
						Eventually(CF("push", appName, "--no-start", "-p", appDir, "--no-route"), CFLongTimeout).Should(Exit(0))
					})
				})

				It("returns a warning and continues", func() {
					session := CF("unbind-service", appName, serviceInstance)
					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(Say("OK"))
					Expect(session.Err).To(Say("Binding between %s and %s did not exist", serviceInstance, appName))
				})
			})

			Context("when the service does not exist", func() {
				BeforeEach(func() {
					WithSimpleApp(func(appDir string) {
						Eventually(CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route"), CFLongTimeout).Should(Exit(0))
					})
				})

				It("fails to unbind the service", func() {
					session := CF("unbind-service", appName, serviceInstance)
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(Say("FAILED"))
					Expect(session.Err).To(Say("Service instance %s not found", serviceInstance))
				})
			})

			Context("when the app does not exist", func() {
				BeforeEach(func() {
					Eventually(CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
				})

				It("fails to unbind the service", func() {
					session := CF("unbind-service", appName, serviceInstance)
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(Say("FAILED"))
					Expect(session.Err).To(Say("App %s not found", appName))
				})
			})
		})
	})
})
