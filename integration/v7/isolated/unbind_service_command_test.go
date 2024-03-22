package isolated

import (
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unbind-service command", func() {
	const command = "unbind-service"

	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+unbind-service - Unbind a service instance from an app\n`),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf unbind-service APP_NAME SERVICE_INSTANCE\n`),
			Say(`\n`),
			Say(`ALIAS:\n`),
			Say(`\s+us\n`),
			Say(`\n`),
			Say(`OPTIONS:\n`),
			Say(`\s+--wait, -w\s+Wait for the operation to complete\n`),
			Say(`\n`),
			Say(`SEE ALSO:\n`),
			Say(`\s+apps, delete-service, services\n`),
		)

		When("the -h flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF(command, "-h")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("the --help flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF(command, "--help")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("no arguments are provided", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF(command)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required arguments `APP_NAME` and `SERVICE_INSTANCE` were not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("unknown flag is passed", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF(command, "-u")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `u"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "bind-service", "app-name", "service-name")
		})
	})

	When("targeting a space", func() {
		var (
			orgName   string
			spaceName string
			username  string
		)

		getBindingCount := func(serviceInstanceName string) int {
			var receiver struct {
				Resources []resources.ServiceCredentialBinding `json:"resources"`
			}
			helpers.Curl(&receiver, "/v3/service_credential_bindings?service_instance_names=%s", serviceInstanceName)
			return len(receiver.Resources)
		}

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)

			username, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("user-provided route service", func() {
			var (
				serviceInstanceName string
				appName             string
			)

			BeforeEach(func() {
				serviceInstanceName = helpers.NewServiceInstanceName()
				Eventually(helpers.CF("cups", serviceInstanceName)).Should(Exit(0))

				appName = helpers.NewAppName()
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})

				Eventually(helpers.CF("bind-service", appName, serviceInstanceName, "--wait")).Should(Exit(0))
			})

			It("deletes the binding", func() {
				session := helpers.CF(command, appName, serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Unbinding app %s from service %s in org %s / space %s as %s\.\.\.\n`, appName, serviceInstanceName, orgName, spaceName, username),
					Say(`OK\n`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(getBindingCount(serviceInstanceName)).To(BeZero())
			})
		})

		Context("managed service instance with synchronous broker response", func() {
			var (
				broker              *servicebrokerstub.ServiceBrokerStub
				serviceInstanceName string
				appName             string
			)

			BeforeEach(func() {
				broker = servicebrokerstub.EnableServiceAccess()
				serviceInstanceName = helpers.NewServiceInstanceName()
				helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)

				appName = helpers.NewAppName()
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})

				Eventually(helpers.CF("bind-service", appName, serviceInstanceName, "--wait")).Should(Exit(0))
			})

			AfterEach(func() {
				broker.Forget()
			})

			It("deletes the binding", func() {
				session := helpers.CF(command, appName, serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Unbinding app %s from service %s in org %s / space %s as %s\.\.\.\n`, appName, serviceInstanceName, orgName, spaceName, username),
					Say(`OK\n`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(getBindingCount(serviceInstanceName)).To(BeZero())
			})
		})

		Context("managed service instance with asynchronous broker response", func() {
			var (
				broker              *servicebrokerstub.ServiceBrokerStub
				serviceInstanceName string
				appName             string
			)

			BeforeEach(func() {
				broker = servicebrokerstub.EnableServiceAccess()
				serviceInstanceName = helpers.NewServiceInstanceName()
				helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)

				appName = helpers.NewAppName()
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})

				Eventually(helpers.CF("bind-service", appName, serviceInstanceName, "--wait")).Should(Exit(0))

				broker.WithAsyncDelay(time.Second).Configure()
			})

			AfterEach(func() {
				broker.Forget()
			})

			It("starts to delete the binding", func() {
				session := helpers.CF(command, appName, serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Unbinding app %s from service %s in org %s / space %s as %s\.\.\.\n`, appName, serviceInstanceName, orgName, spaceName, username),
					Say(`OK\n`),
					Say(`Unbinding in progress. Use 'cf service %s' to check operation status\.\n`, serviceInstanceName),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(getBindingCount(serviceInstanceName)).NotTo(BeZero())
			})

			When("--wait flag specified", func() {
				It("waits for completion", func() {
					session := helpers.CF(command, appName, serviceInstanceName, "--wait")
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say(`Unbinding app %s from service %s in org %s / space %s as %s\.\.\.\n`, appName, serviceInstanceName, orgName, spaceName, username),
						Say(`Waiting for the operation to complete\.+\n`),
						Say(`\n`),
						Say(`OK\n`),
					))

					Expect(string(session.Err.Contents())).To(BeEmpty())
					Expect(getBindingCount(serviceInstanceName)).To(BeZero())
				})
			})
		})

		Context("binding does not exist", func() {
			var (
				serviceInstanceName string
				appName             string
			)

			BeforeEach(func() {
				serviceInstanceName = helpers.NewServiceInstanceName()
				Eventually(helpers.CF("cups", serviceInstanceName)).Should(Exit(0))

				appName = helpers.NewAppName()
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})
			})

			It("succeeds saying the binding did not exist", func() {
				session := helpers.CF(command, appName, serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Unbinding app %s from service %s in org %s / space %s as %s\.\.\.\n`, appName, serviceInstanceName, orgName, spaceName, username),
					Say(`Binding between %s and %s did not exist\n`, serviceInstanceName, appName),
					Say(`OK\n`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())
			})
		})

		Context("app does not exist", func() {
			var serviceInstanceName string

			BeforeEach(func() {
				serviceInstanceName = helpers.NewServiceInstanceName()
				Eventually(helpers.CF("cups", serviceInstanceName)).Should(Exit(0))
			})

			It("displays FAILED and app not found", func() {
				session := helpers.CF(command, "does-not-exist", serviceInstanceName)
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("App 'does-not-exist' not found"))
			})
		})

		Context("service instance does not exist", func() {
			var appName string

			BeforeEach(func() {
				appName = helpers.NewAppName()
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})
			})

			It("displays FAILED and service not found", func() {
				session := helpers.CF(command, appName, "does-not-exist")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("Service instance 'does-not-exist' not found"))
			})
		})
	})
})
