package isolated

import (
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
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

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "unbind-service", "app-name", "service-name")
		})
	})

	When("the environment is setup correctly", func() {
		var (
			org   string
			space string
		)

		BeforeEach(func() {
			org = helpers.NewOrgName()
			space = helpers.NewSpaceName()
			helpers.SetupCF(org, space)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(org)
		})

		When("the service is provided by a user", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-user-provided-service", serviceInstance, "-p", "{}")).Should(Exit(0))
			})

			AfterEach(func() {
				Eventually(helpers.CF("delete-service", serviceInstance, "-f")).Should(Exit(0))
			})

			When("the service is bound to an app", func() {
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

			When("the service is not bound to an app", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "--no-route")).Should(Exit(0))
					})
				})

				It("returns a warning and continues", func() {
					session := helpers.CF("unbind-service", appName, serviceInstance)
					Eventually(session).Should(Say("OK"))
					Eventually(session.Err).Should(Say("Binding between %s and %s did not exist", serviceInstance, appName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the service does not exist", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
					})
				})

				It("fails to unbind the service", func() {
					session := helpers.CF("unbind-service", appName, "does-not-exist")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Service instance %s not found", "does-not-exist"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the app does not exist", func() {
				It("fails to unbind the service", func() {
					session := helpers.CF("unbind-service", "does-not-exist", serviceInstance)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("App '%s' not found", "does-not-exist"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the service is provided by a broker", func() {
			var (
				service     string
				servicePlan string
				broker      *servicebrokerstub.ServiceBrokerStub
			)

			When("the unbinding is asynchronous", func() {
				BeforeEach(func() {
					broker = servicebrokerstub.New().WithAsyncDelay(time.Millisecond).EnableServiceAccess()
					service = broker.FirstServiceOfferingName()
					servicePlan = broker.FirstServicePlanName()

					Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
					})

					Eventually(func() *Session {
						session := helpers.CF("service", serviceInstance)
						return session.Wait()
					}, time.Minute*5, time.Second*5).Should(Say("create succeeded"))

					Eventually(helpers.CF("bind-service", appName, serviceInstance)).Should(Exit(0))
					client, err := helpers.CreateCCV2Client()
					Expect(err).ToNot(HaveOccurred())
					helpers.PollLastOperationUntilSuccess(client, appName, serviceInstance)
				})

				AfterEach(func() {
					broker.Forget()
				})

				It("returns that the unbind is in progress", func() {
					session := helpers.CF("unbind-service", appName, serviceInstance)
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Unbinding in progress. Use 'cf service %s' to check operation status.", serviceInstance))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the unbinding is blocking", func() {
				BeforeEach(func() {
					broker = servicebrokerstub.EnableServiceAccess()
					service = broker.FirstServiceOfferingName()
					servicePlan = broker.FirstServicePlanName()
				})

				AfterEach(func() {
					broker.Forget()
				})

				When("the service is bound to an app", func() {
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

				When("the service is not bound to an app", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "--no-route")).Should(Exit(0))
						})
					})

					It("returns a warning and continues", func() {
						session := helpers.CF("unbind-service", appName, serviceInstance)
						Eventually(session).Should(Say("OK"))
						Eventually(session.Err).Should(Say("Binding between %s and %s did not exist", serviceInstance, appName))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the service does not exist", func() {
					BeforeEach(func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
						})
					})

					It("fails to unbind the service", func() {
						session := helpers.CF("unbind-service", appName, serviceInstance)
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Service instance %s not found", serviceInstance))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the app does not exist", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
					})

					It("fails to unbind the service", func() {
						session := helpers.CF("unbind-service", appName, serviceInstance)
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("App '%s' not found", appName))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})
	})
})
