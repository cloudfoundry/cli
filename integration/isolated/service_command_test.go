package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("service command", func() {
	var serviceInstanceName string

	BeforeEach(func() {
		serviceInstanceName = helpers.PrefixedRandomName("SI")
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("service", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("\\s+service - Show service instance info"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("\\s+cf service SERVICE_INSTANCE"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("\\s+\\-\\-guid\\s+Retrieve and display the given service's guid\\. All other output for the service is suppressed\\."))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("\\s+bind-service, rename-service, update-service"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("service", serviceInstanceName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("service", serviceInstanceName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in\\. Use 'cf login' to log in\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when an org is not targeted", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no targeted org error", func() {
				session := helpers.CF("service", serviceInstanceName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when a space is not targeted", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no targeted space error", func() {
				session := helpers.CF("service", serviceInstanceName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space\\."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when an api is targeted, the user is logged in, and an org and space are targeted", func() {
		var (
			orgName   string
			spaceName string
			userName  string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			setupCF(orgName, spaceName)

			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when the service instance does not exist", func() {
			It("returns an error and exits 1", func() {
				session := helpers.CF("service", serviceInstanceName)
				Eventually(session.Out).Should(Say("Showing info of service %s in org %s / space %s as %s", serviceInstanceName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Service instance %s not found", serviceInstanceName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the service instance exists", func() {
			Context("when the service instance is a user provided service instance", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-user-provided-service", serviceInstanceName)).Should(Exit(0))
				})

				AfterEach(func() {
					Eventually(helpers.CF("delete-service", serviceInstanceName, "-f")).Should(Exit(0))
				})

				Context("when the --guid flag is provided", func() {
					It("displays the service instance GUID", func() {
						session := helpers.CF("service", serviceInstanceName, "--guid")
						Eventually(session.Out).Should(Say("Showing info of service %s in org %s / space %s as %s", serviceInstanceName, orgName, spaceName, userName))
						Eventually(session.Out).Should(Say(""))
						Eventually(session.Out).Should(Say(helpers.UserProvidedServiceInstanceGUID(serviceInstanceName)))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when apps are bound to the service instance", func() {
					var (
						appName1 string
						appName2 string
					)

					BeforeEach(func() {
						appName1 = helpers.NewAppName()
						appName2 = helpers.NewAppName()
						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CF("push", appName1, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
							Eventually(helpers.CF("push", appName2, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
						})
						Eventually(helpers.CF("bind-service", appName1, serviceInstanceName)).Should(Exit(0))
						Eventually(helpers.CF("bind-service", appName2, serviceInstanceName)).Should(Exit(0))
					})

					AfterEach(func() {
						Eventually(helpers.CF("unbind-service", appName1, serviceInstanceName)).Should(Exit(0))
						Eventually(helpers.CF("unbind-service", appName2, serviceInstanceName)).Should(Exit(0))
						Eventually(helpers.CF("delete", appName1, "-f")).Should(Exit(0))
						Eventually(helpers.CF("delete", appName2, "-f")).Should(Exit(0))
					})

					It("displays service instance info", func() {
						session := helpers.CF("service", serviceInstanceName)
						Eventually(session.Out).Should(Say("Showing info of service %s in org %s / space %s as %s", serviceInstanceName, orgName, spaceName, userName))
						Eventually(session.Out).Should(Say(""))
						Eventually(session.Out).Should(Say("name:\\s+%s", serviceInstanceName))
						Eventually(session.Out).Should(Say("service:\\s+user-provided"))
						Eventually(session.Out).Should(Say("bound apps:\\s+%s, %s", appName1, appName2))
						Eventually(session.Out).ShouldNot(Say("tags:"))
						Eventually(session.Out).ShouldNot(Say("plan:"))
						Eventually(session.Out).ShouldNot(Say("description:"))
						Eventually(session.Out).ShouldNot(Say("documentation:"))
						Eventually(session.Out).ShouldNot(Say("dashboard:"))
						Eventually(session.Out).ShouldNot(Say("last operation"))
						Eventually(session.Out).ShouldNot(Say("status:"))
						Eventually(session.Out).ShouldNot(Say("message:"))
						Eventually(session.Out).ShouldNot(Say("started:"))
						Eventually(session.Out).ShouldNot(Say("updated:"))

						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when the service instance is a managed service instance", func() {
				var (
					domain      string
					service     string
					servicePlan string
					broker      helpers.ServiceBroker
				)

				BeforeEach(func() {
					domain = defaultSharedDomain()
					service = helpers.PrefixedRandomName("SERVICE")
					servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")

					broker = helpers.NewServiceBroker(helpers.NewServiceBrokerName(), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
					broker.Push()
					broker.Configure()
					broker.Create()

					Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
					Eventually(helpers.CF("create-service", service, servicePlan, serviceInstanceName, "-t", "database, email")).Should(Exit(0))
				})

				AfterEach(func() {
					Eventually(helpers.CF("delete-service", serviceInstanceName, "-f")).Should(Exit(0))
					broker.Destroy()
				})

				Context("when the --guid flag is provided", func() {
					It("displays the service instance GUID", func() {
						session := helpers.CF("service", serviceInstanceName, "--guid")
						Eventually(session.Out).Should(Say("Showing info of service %s in org %s / space %s as %s", serviceInstanceName, orgName, spaceName, userName))
						Eventually(session.Out).Should(Say(""))
						Eventually(session.Out).Should(Say(helpers.ManagedServiceInstanceGUID(serviceInstanceName)))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when apps are bound to the service instance", func() {
					var (
						appName1 string
						appName2 string
					)

					BeforeEach(func() {
						appName1 = helpers.NewAppName()
						appName2 = helpers.NewAppName()
						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CF("push", appName1, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
							Eventually(helpers.CF("push", appName2, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
						})
						Eventually(helpers.CF("bind-service", appName1, serviceInstanceName)).Should(Exit(0))
						Eventually(helpers.CF("bind-service", appName2, serviceInstanceName)).Should(Exit(0))
					})

					AfterEach(func() {
						Eventually(helpers.CF("unbind-service", appName1, serviceInstanceName)).Should(Exit(0))
						Eventually(helpers.CF("unbind-service", appName2, serviceInstanceName)).Should(Exit(0))
						Eventually(helpers.CF("delete", appName1, "-f")).Should(Exit(0))
						Eventually(helpers.CF("delete", appName2, "-f")).Should(Exit(0))
					})

					It("displays service instance info", func() {
						session := helpers.CF("service", serviceInstanceName)
						Eventually(session.Out).Should(Say("Showing info of service %s in org %s / space %s as %s\\.\\.\\.", serviceInstanceName, orgName, spaceName, userName))
						Eventually(session.Out).Should(Say("\n\n"))
						Eventually(session.Out).Should(Say("name:\\s+%s", serviceInstanceName))
						Eventually(session.Out).Should(Say("service:\\s+%s", service))
						Eventually(session.Out).Should(Say("bound apps:\\s+%s, %s", appName1, appName2))
						Eventually(session.Out).Should(Say("tags:\\s+database, email"))
						Eventually(session.Out).Should(Say("plan:\\s+%s", servicePlan))
						Eventually(session.Out).Should(Say("description:\\s+fake service"))
						Eventually(session.Out).Should(Say("documentation:"))
						Eventually(session.Out).Should(Say("dashboard:\\s+http://example\\.com"))
						Eventually(session.Out).Should(Say("\n\n"))
						Eventually(session.Out).Should(Say("Showing status of last operation from service %s\\.\\.\\.", serviceInstanceName))
						Eventually(session.Out).Should(Say("\n\n"))
						Eventually(session.Out).Should(Say("status:\\s+create succeeded"))
						Eventually(session.Out).Should(Say("message:"))
						Eventually(session.Out).Should(Say("started:\\s+\\d{4}-[01]\\d-[0-3]\\dT[0-2][0-9]:[0-5]\\d:[0-5]\\dZ"))
						Eventually(session.Out).Should(Say("updated:\\s+\\d{4}-[01]\\d-[0-3]\\dT[0-2][0-9]:[0-5]\\d:[0-5]\\dZ"))

						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
