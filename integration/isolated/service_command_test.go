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
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "service", "some-service")
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

		Context("when the service instance belongs to this space", func() {
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
						Consistently(session.Out).ShouldNot(Say("Showing info of service %s in org %s / space %s as %s", serviceInstanceName, orgName, spaceName, userName))
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
						Consistently(session.Out).ShouldNot(Say("shared from:"))
						Eventually(session.Out).Should(Say("service:\\s+user-provided"))
						Eventually(session.Out).Should(Say("bound apps:\\s+%s, %s", appName1, appName2))
						Consistently(session.Out).ShouldNot(Say("tags:"))
						Consistently(session.Out).ShouldNot(Say("plan:"))
						Consistently(session.Out).ShouldNot(Say("description:"))
						Consistently(session.Out).ShouldNot(Say("documentation:"))
						Consistently(session.Out).ShouldNot(Say("dashboard:"))
						Consistently(session.Out).ShouldNot(Say("shared with spaces:"))
						Consistently(session.Out).ShouldNot(Say("last operation"))
						Consistently(session.Out).ShouldNot(Say("status:"))
						Consistently(session.Out).ShouldNot(Say("message:"))
						Consistently(session.Out).ShouldNot(Say("started:"))
						Consistently(session.Out).ShouldNot(Say("updated:"))
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
						Consistently(session.Out).ShouldNot(Say("Showing info of service %s in org %s / space %s as %s", serviceInstanceName, orgName, spaceName, userName))
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
						Consistently(session.Out).ShouldNot(Say("shared from:"))
						Eventually(session.Out).Should(Say("service:\\s+%s", service))
						Eventually(session.Out).Should(Say("bound apps:\\s+%s, %s", appName1, appName2))
						Eventually(session.Out).Should(Say("tags:\\s+database, email"))
						Eventually(session.Out).Should(Say("plan:\\s+%s", servicePlan))
						Eventually(session.Out).Should(Say("description:\\s+fake service"))
						Eventually(session.Out).Should(Say("documentation:"))
						Eventually(session.Out).Should(Say("dashboard:\\s+http://example\\.com"))
						Eventually(session.Out).Should(Say("\n\n"))
						Consistently(session.Out).ShouldNot(Say("shared with spaces:"))
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

	PDescribe("a user with SpaceDeveloperRole access can see a service instance's sharing information", func() {
		Context("when there are two spaces in the current org and a service broker", func() {
			var (
				orgName         string
				targetSpaceName string
				sourceSpaceName string

				domain      string
				service     string
				servicePlan string
				broker      helpers.ServiceBroker
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				targetSpaceName = helpers.NewSpaceName()
				sourceSpaceName = helpers.NewSpaceName()
				setupCF(orgName, targetSpaceName)
				setupCF(orgName, sourceSpaceName)

				domain = defaultSharedDomain()
				service = helpers.PrefixedRandomName("SERVICE")
				servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")
				broker = helpers.NewServiceBroker(helpers.NewServiceBrokerName(), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
				broker.Push()
				broker.Configure()
				broker.Create()

				Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
			})

			AfterEach(func() {
				// need to login as admin
				helpers.LoginCF()
				helpers.TargetOrgAndSpace(orgName, sourceSpaceName)
				broker.Destroy()
				helpers.QuickDeleteOrg(orgName)
			})

			Context("when there is a user with SpaceDeveloperRole access to both spaces", func() {
				var username string

				BeforeEach(func() {
					username = helpers.PrefixedRandomName("user")
					password := helpers.RandomName()

					Eventually(helpers.CF("create-user", username, password)).Should(Exit(0))
					Eventually(helpers.CF("set-space-role", username, orgName, sourceSpaceName, "SpaceDeveloper")).Should(Exit(0))
					Eventually(helpers.CF("set-space-role", username, orgName, targetSpaceName, "SpaceDeveloper")).Should(Exit(0))
					Eventually(helpers.CF("auth", username, password)).Should(Exit(0))
				})

				Context("when the user creates a service instance in the source space", func() {
					BeforeEach(func() {
						helpers.TargetOrgAndSpace(orgName, sourceSpaceName)
						Eventually(helpers.CF("create-service", service, servicePlan, serviceInstanceName, "-t", "database, email")).Should(Exit(0))
					})

					AfterEach(func() {
						helpers.TargetOrgAndSpace(orgName, sourceSpaceName)
						Eventually(helpers.CF("delete-service", serviceInstanceName, "-f")).Should(Exit(0))
					})

					It("should not display shared from or shared with information", func() {
						session := helpers.CF("service", serviceInstanceName)
						Consistently(session.Out).ShouldNot(Say("shared from:"))
						Consistently(session.Out).ShouldNot(Say("shared with spaces:"))
						Eventually(session).Should(Exit(0))
					})

					Context("when this user shares the service instance with the target space", func() {
						BeforeEach(func() {
							Eventually(helpers.CF("v3-share-service", serviceInstanceName, "-s", targetSpaceName)).Should(Exit(0))
						})

						AfterEach(func() {
							helpers.TargetOrgAndSpace(orgName, sourceSpaceName)
							Eventually(helpers.CF("v3-unshare-service", serviceInstanceName, "-s", targetSpaceName, "-f")).Should(Exit(0))
						})

						Context("when the user is targeted to the target space", func() {
							BeforeEach(func() {
								helpers.TargetOrgAndSpace(orgName, targetSpaceName)
							})

							It("the user should see where the service instance is shared from and not see where it is shared with", func() {
								session := helpers.CF("service", serviceInstanceName)
								Eventually(session.Out).Should(Say("Showing info of service %s in org %s / space %s as %s\\.\\.\\.", serviceInstanceName, orgName, targetSpaceName, username))
								Eventually(session.Out).Should(Say("\n\n"))
								Eventually(session.Out).Should(Say("name:\\s+%s", serviceInstanceName))
								Eventually(session.Out).Should(Say("shared from org/space:\\s+%s / %s", orgName, sourceSpaceName))
								Eventually(session.Out).Should(Say("service:\\s+%s", service))
								Eventually(session.Out).Should(Say("bound apps:"))
								Eventually(session.Out).Should(Say("tags:\\s+database, email"))
								Eventually(session.Out).Should(Say("plan:\\s+%s", servicePlan))
								Eventually(session.Out).Should(Say("description:\\s+fake service"))
								Eventually(session.Out).Should(Say("documentation:"))
								Eventually(session.Out).Should(Say("dashboard:\\s+http://example\\.com"))
								Eventually(session.Out).Should(Say("\n\n"))
								Consistently(session.Out).ShouldNot(Say("shared with spaces:"))
								Eventually(session.Out).Should(Say("Showing status of last operation from service %s\\.\\.\\.", serviceInstanceName))
								Eventually(session.Out).Should(Say("\n\n"))
								Eventually(session.Out).Should(Say("status:\\s+create succeeded"))
								Eventually(session.Out).Should(Say("message:"))
								Eventually(session.Out).Should(Say("started:\\s+\\d{4}-[01]\\d-[0-3]\\dT[0-2][0-9]:[0-5]\\d:[0-5]\\dZ"))
								Eventually(session.Out).Should(Say("updated:\\s+\\d{4}-[01]\\d-[0-3]\\dT[0-2][0-9]:[0-5]\\d:[0-5]\\dZ"))
								Eventually(session).Should(Exit(0))
							})

							Context("when the user binds the shared service instance to apps in the target space and then targets the source space", func() {
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
									helpers.TargetOrgAndSpace(orgName, sourceSpaceName)
								})

								AfterEach(func() {
									helpers.TargetOrgAndSpace(orgName, targetSpaceName)
									Eventually(helpers.CF("unbind-service", appName1, serviceInstanceName)).Should(Exit(0))
									Eventually(helpers.CF("unbind-service", appName2, serviceInstanceName)).Should(Exit(0))
									Eventually(helpers.CF("delete", appName1, "-f")).Should(Exit(0))
									Eventually(helpers.CF("delete", appName2, "-f")).Should(Exit(0))
								})

								It("should display shared with information with the correct number of bound apps", func() {
									session := helpers.CF("service", serviceInstanceName)
									Eventually(session.Out).Should(Say("Showing info of service %s in org %s / space %s as %s\\.\\.\\.", serviceInstanceName, orgName, sourceSpaceName, username))
									Eventually(session.Out).Should(Say("name:\\s+%s", serviceInstanceName))
									Consistently(session.Out).ShouldNot(Say("shared from org/space:"))
									Eventually(session.Out).Should(Say("service:"))
									Eventually(session.Out).Should(Say("dashboard:"))
									Eventually(session.Out).Should(Say("shared with spaces:"))
									Eventually(session.Out).Should(Say("org\\s+space\\s+bindings"))
									Eventually(session.Out).Should(Say("%s\\s+%s\\s+2", orgName, targetSpaceName))
									Eventually(session.Out).Should(Say("Showing status of last operation from service %s\\.\\.\\.", serviceInstanceName))
									Eventually(session).Should(Exit(0))
								})
							})
						})

						Context("when the user is targeted to the source space", func() {
							BeforeEach(func() {
								helpers.TargetOrgAndSpace(orgName, sourceSpaceName)
							})

							It("should display which spaces the service instance is shared with and not display where it is shared from", func() {
								session := helpers.CF("service", serviceInstanceName)
								Eventually(session.Out).Should(Say("Showing info of service %s in org %s / space %s as %s\\.\\.\\.", serviceInstanceName, orgName, sourceSpaceName, username))
								Eventually(session.Out).Should(Say("\n\n"))
								Eventually(session.Out).Should(Say("name:\\s+%s", serviceInstanceName))
								Consistently(session.Out).ShouldNot(Say("shared from org/space:"))
								Eventually(session.Out).Should(Say("service:\\s+%s", service))
								Eventually(session.Out).Should(Say("bound apps:"))
								Eventually(session.Out).Should(Say("tags:\\s+database, email"))
								Eventually(session.Out).Should(Say("plan:\\s+%s", servicePlan))
								Eventually(session.Out).Should(Say("description:\\s+fake service"))
								Eventually(session.Out).Should(Say("documentation:"))
								Eventually(session.Out).Should(Say("dashboard:\\s+http://example\\.com"))
								Eventually(session.Out).Should(Say("\n\n"))
								Eventually(session.Out).Should(Say("shared with spaces:"))
								Eventually(session.Out).Should(Say("org\\s+space\\s+bindings"))
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
	})
})
