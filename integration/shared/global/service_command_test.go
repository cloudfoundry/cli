// +build !partialPush

package global

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
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
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("service", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`\s+service - Show service instance info`))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`\s+cf service SERVICE_INSTANCE`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`\s+\-\-guid\s+Retrieve and display the given service's guid\. All other output for the service is suppressed\.`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say(`\s+bind-service, rename-service, update-service`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "service", "some-service")
		})
	})

	When("an api is targeted, the user is logged in, and an org and space are targeted", func() {
		var (
			orgName   string
			spaceName string
			userName  string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)

			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the service instance does not exist", func() {
			It("returns an error and exits 1", func() {
				session := helpers.CF("service", serviceInstanceName)
				Eventually(session).Should(Say("Showing info of service %s in org %s / space %s as %s", serviceInstanceName, orgName, spaceName, userName))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Service instance %s not found", serviceInstanceName))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the service instance belongs to this space", func() {
			When("the service instance is a user provided service instance", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-user-provided-service", serviceInstanceName)).Should(Exit(0))
				})

				AfterEach(func() {
					Eventually(helpers.CF("delete-service", serviceInstanceName, "-f")).Should(Exit(0))
				})

				When("the --guid flag is provided", func() {
					It("displays the service instance GUID", func() {
						session := helpers.CF("service", serviceInstanceName, "--guid")
						Consistently(session).ShouldNot(Say("Showing info of service %s in org %s / space %s as %s", serviceInstanceName, orgName, spaceName, userName))
						Eventually(session).Should(Say(helpers.UserProvidedServiceInstanceGUID(serviceInstanceName)))
						Eventually(session).Should(Exit(0))
					})
				})

				When("no apps are bound to the service instance", func() {
					It("displays service instance info", func() {
						session := helpers.CF("service", serviceInstanceName)
						Eventually(session).Should(Say("Showing info of service %s in org %s / space %s as %s", serviceInstanceName, orgName, spaceName, userName))
						Eventually(session).Should(Say(""))
						Eventually(session).Should(Say(`name:\s+%s`, serviceInstanceName))
						Eventually(session).Should(Say(`service:\s+user-provided`))
						Eventually(session).Should(Say(""))
						Eventually(session).Should(Say("There are no bound apps for this service."))
						Eventually(session).Should(Say(""))
						Eventually(session).Should(Exit(0))
					})
				})

				When("apps are bound to the service instance", func() {
					var (
						appName1 string
						appName2 string
					)

					BeforeEach(func() {
						appName1 = helpers.PrefixedRandomName("1-INTEGRATION-APP")
						appName2 = helpers.PrefixedRandomName("2-INTEGRATION-APP")

						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CF("push", appName1, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
							Eventually(helpers.CF("push", appName2, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
						})
					})

					AfterEach(func() {
						Eventually(helpers.CF("delete", appName1, "-f")).Should(Exit(0))
						Eventually(helpers.CF("delete", appName2, "-f")).Should(Exit(0))
					})

					When("the service bindings do not have binding names", func() {
						BeforeEach(func() {
							Eventually(helpers.CF("bind-service", appName1, serviceInstanceName)).Should(Exit(0))
							Eventually(helpers.CF("bind-service", appName2, serviceInstanceName)).Should(Exit(0))
						})

						AfterEach(func() {
							Eventually(helpers.CF("unbind-service", appName1, serviceInstanceName)).Should(Exit(0))
							Eventually(helpers.CF("unbind-service", appName2, serviceInstanceName)).Should(Exit(0))
						})

						It("displays service instance info", func() {
							session := helpers.CF("service", serviceInstanceName)
							Eventually(session).Should(Say("Showing info of service %s in org %s / space %s as %s", serviceInstanceName, orgName, spaceName, userName))
							Eventually(session).Should(Say(""))
							Eventually(session).Should(Say(`name:\s+%s`, serviceInstanceName))
							Eventually(session).Should(Say(`service:\s+user-provided`))
							Eventually(session).Should(Say(""))
							Eventually(session).Should(Say("bound apps:"))
							Eventually(session).Should(Say(`name\s+binding name\s+status\s+message`))
							Eventually(session).Should(Say(appName1))
							Eventually(session).Should(Say(appName2))

							Eventually(session).Should(Exit(0))
						})
					})

					When("the service bindings have binding names", func() {
						var (
							bindingName1 string
							bindingName2 string
						)

						BeforeEach(func() {
							helpers.SkipIfVersionLessThan(ccversion.MinVersionBindingNameV2)
							bindingName1 = helpers.PrefixedRandomName("BINDING-NAME")
							bindingName2 = helpers.PrefixedRandomName("BINDING-NAME")
							Eventually(helpers.CF("bind-service", appName1, serviceInstanceName, "--binding-name", bindingName1)).Should(Exit(0))
							Eventually(helpers.CF("bind-service", appName2, serviceInstanceName, "--binding-name", bindingName2)).Should(Exit(0))
						})

						AfterEach(func() {
							Eventually(helpers.CF("unbind-service", appName1, serviceInstanceName)).Should(Exit(0))
							Eventually(helpers.CF("unbind-service", appName2, serviceInstanceName)).Should(Exit(0))
						})

						It("displays service instance info", func() {
							session := helpers.CF("service", serviceInstanceName)
							Eventually(session).Should(Say("Showing info of service %s in org %s / space %s as %s", serviceInstanceName, orgName, spaceName, userName))
							Eventually(session).Should(Say(""))
							Eventually(session).Should(Say(`name:\s+%s`, serviceInstanceName))
							Eventually(session).Should(Say(`service:\s+user-provided`))
							Eventually(session).Should(Say(""))
							Eventually(session).Should(Say("bound apps:"))
							Eventually(session).Should(Say(`name\s+binding name\s+status\s+message`))
							Eventually(session).Should(Say(`%s\s+%s`, appName1, bindingName1))
							Eventually(session).Should(Say(`%s\s+%s`, appName2, bindingName2))
							Eventually(session).Should(Say(""))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("we update the user provided service instance with tags", func() {
					BeforeEach(func() {
						helpers.SkipIfVersionLessThan(ccversion.MinVersionUserProvidedServiceTagsV2)
						Eventually(helpers.CF("update-user-provided-service", serviceInstanceName,
							"-t", "foo, bar")).Should(Exit(0))
					})

					It("displays service instance info", func() {
						session := helpers.CF("service", serviceInstanceName)
						Eventually(session).Should(Say("Showing info of service %s in org %s / space %s as %s", serviceInstanceName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`tags:\s+foo, bar`))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("a user-provided service instance is created with tags", func() {
				BeforeEach(func() {
					helpers.SkipIfVersionLessThan(ccversion.MinVersionUserProvidedServiceTagsV2)
					Eventually(helpers.CF("create-user-provided-service", serviceInstanceName, "-t", "database, email")).Should(Exit(0))
				})

				It("displays tag info", func() {
					session := helpers.CF("service", serviceInstanceName)
					Eventually(session).Should(Say("Showing info of service %s in org %s / space %s as %s", serviceInstanceName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`tags:\s+database, email`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the service instance is a managed service instance", func() {
				var (
					domain      string
					service     string
					servicePlan string
					broker      helpers.ServiceBroker
				)

				BeforeEach(func() {
					domain = helpers.DefaultSharedDomain()
					service = helpers.PrefixedRandomName("SERVICE")
					servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")

					broker = helpers.NewServiceBroker(helpers.NewServiceBrokerName(), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
					broker.Push()
					broker.Configure(true)
					broker.Create()

					Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
					Eventually(helpers.CF("create-service", service, servicePlan, serviceInstanceName, "-t", "database, email")).Should(Exit(0))
				})

				AfterEach(func() {
					Eventually(helpers.CF("delete-service", serviceInstanceName, "-f")).Should(Exit(0))
					broker.Destroy()
				})

				When("the --guid flag is provided", func() {
					It("displays the service instance GUID", func() {
						session := helpers.CF("service", serviceInstanceName, "--guid")
						Consistently(session).ShouldNot(Say("Showing info of service %s in org %s / space %s as %s", serviceInstanceName, orgName, spaceName, userName))
						Eventually(session).Should(Say(helpers.ManagedServiceInstanceGUID(serviceInstanceName)))
						Eventually(session).Should(Exit(0))
					})
				})

				When("apps are bound to the service instance", func() {
					var (
						appName1 string
						appName2 string
					)

					BeforeEach(func() {
						appName1 = helpers.PrefixedRandomName("1-INTEGRATION-APP")
						appName2 = helpers.PrefixedRandomName("2-INTEGRATION-APP")

						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CF("push", appName1, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
							Eventually(helpers.CF("push", appName2, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
						})
					})

					AfterEach(func() {
						Eventually(helpers.CF("delete", appName1, "-f")).Should(Exit(0))
						Eventually(helpers.CF("delete", appName2, "-f")).Should(Exit(0))
					})

					When("the service bindings do not have binding names", func() {
						BeforeEach(func() {
							Eventually(helpers.CF("bind-service", appName1, serviceInstanceName)).Should(Exit(0))
							Eventually(helpers.CF("bind-service", appName2, serviceInstanceName)).Should(Exit(0))
						})

						AfterEach(func() {
							Eventually(helpers.CF("unbind-service", appName1, serviceInstanceName)).Should(Exit(0))
							Eventually(helpers.CF("unbind-service", appName2, serviceInstanceName)).Should(Exit(0))
						})

						It("displays service instance info", func() {
							session := helpers.CF("service", serviceInstanceName)
							Eventually(session).Should(Say(`Showing info of service %s in org %s / space %s as %s\.\.\.`, serviceInstanceName, orgName, spaceName, userName))
							Eventually(session).Should(Say("\n\n"))
							Eventually(session).Should(Say(`name:\s+%s`, serviceInstanceName))
							Consistently(session).ShouldNot(Say("shared from:"))
							Eventually(session).Should(Say(`service:\s+%s`, service))
							Eventually(session).Should(Say(`tags:\s+database, email`))
							Eventually(session).Should(Say(`plan:\s+%s`, servicePlan))
							Eventually(session).Should(Say(`description:\s+fake service`))
							Eventually(session).Should(Say(`documentation:\s+http://documentation\.url`))
							Eventually(session).Should(Say(`dashboard:\s+http://example\.com`))
							Eventually(session).Should(Say("\n\n"))
							Consistently(session).ShouldNot(Say("shared with spaces:"))
							Eventually(session).Should(Say(`Showing status of last operation from service %s\.\.\.`, serviceInstanceName))
							Eventually(session).Should(Say("\n\n"))
							Eventually(session).Should(Say(`status:\s+create succeeded`))
							Eventually(session).Should(Say("message:"))
							Eventually(session).Should(Say(`started:\s+\d{4}-[01]\d-[0-3]\dT[0-2][0-9]:[0-5]\d:[0-5]\dZ`))
							Eventually(session).Should(Say(`updated:\s+\d{4}-[01]\d-[0-3]\dT[0-2][0-9]:[0-5]\d:[0-5]\dZ`))
							Eventually(session).Should(Say("\n\n"))
							Eventually(session).Should(Say("bound apps:"))
							Eventually(session).Should(Say(`name\s+binding name\s+status\s+message`))
							Eventually(session).Should(Say(appName1))
							Eventually(session).Should(Say(appName2))

							Eventually(session).Should(Exit(0))
						})
					})

					When("the service bindings have binding names", func() {
						var (
							bindingName1 string
							bindingName2 string
						)

						BeforeEach(func() {
							helpers.SkipIfVersionLessThan(ccversion.MinVersionBindingNameV2)
							bindingName1 = helpers.PrefixedRandomName("BINDING-NAME")
							bindingName2 = helpers.PrefixedRandomName("BINDING-NAME")
							Eventually(helpers.CF("bind-service", appName1, serviceInstanceName, "--binding-name", bindingName1)).Should(Exit(0))
							Eventually(helpers.CF("bind-service", appName2, serviceInstanceName, "--binding-name", bindingName2)).Should(Exit(0))
						})

						AfterEach(func() {
							Eventually(helpers.CF("unbind-service", appName1, serviceInstanceName)).Should(Exit(0))
							Eventually(helpers.CF("unbind-service", appName2, serviceInstanceName)).Should(Exit(0))
						})

						It("displays service instance info", func() {
							session := helpers.CF("service", serviceInstanceName)
							Eventually(session).Should(Say(`Showing info of service %s in org %s / space %s as %s\.\.\.`, serviceInstanceName, orgName, spaceName, userName))
							Eventually(session).Should(Say("\n\n"))
							Eventually(session).Should(Say(`name:\s+%s`, serviceInstanceName))
							Consistently(session).ShouldNot(Say("shared from:"))
							Eventually(session).Should(Say(`service:\s+%s`, service))
							Eventually(session).Should(Say(`tags:\s+database, email`))
							Eventually(session).Should(Say(`plan:\s+%s`, servicePlan))
							Eventually(session).Should(Say(`description:\s+fake service`))
							Eventually(session).Should(Say(`documentation:\s+http://documentation\.url`))
							Eventually(session).Should(Say(`dashboard:\s+http://example\.com`))
							Eventually(session).Should(Say("\n\n"))
							Consistently(session).ShouldNot(Say("shared with spaces:"))
							Eventually(session).Should(Say(`Showing status of last operation from service %s\.\.\.`, serviceInstanceName))
							Eventually(session).Should(Say("\n\n"))
							Eventually(session).Should(Say(`status:\s+create succeeded`))
							Eventually(session).Should(Say("message:"))
							Eventually(session).Should(Say(`started:\s+\d{4}-[01]\d-[0-3]\dT[0-2][0-9]:[0-5]\d:[0-5]\dZ`))
							Eventually(session).Should(Say(`updated:\s+\d{4}-[01]\d-[0-3]\dT[0-2][0-9]:[0-5]\d:[0-5]\dZ`))
							Eventually(session).Should(Say("\n\n"))
							Eventually(session).Should(Say("bound apps:"))
							Eventually(session).Should(Say(`name\s+binding name\s+status\s+message`))
							Eventually(session).Should(Say(`%s\s+%s`, appName1, bindingName1))
							Eventually(session).Should(Say(`%s\s+%s`, appName2, bindingName2))

							Eventually(session).Should(Exit(0))
						})
					})

					When("the binding has a state", func() {
						var (
							bindingName1 string
							bindingName2 string
						)

						BeforeEach(func() {
							helpers.SkipIfVersionLessThan(ccversion.MinVersionAsyncBindingsV2)
							bindingName1 = helpers.PrefixedRandomName("BINDING-NAME")
							bindingName2 = helpers.PrefixedRandomName("BINDING-NAME")
							Eventually(helpers.CF("bind-service", appName1, serviceInstanceName, "--binding-name", bindingName1)).Should(Exit(0))
							Eventually(helpers.CF("bind-service", appName2, serviceInstanceName, "--binding-name", bindingName2)).Should(Exit(0))
						})

						AfterEach(func() {
							Eventually(helpers.CF("unbind-service", appName1, serviceInstanceName)).Should(Exit(0))
							Eventually(helpers.CF("unbind-service", appName2, serviceInstanceName)).Should(Exit(0))
						})

						It("displays it in the status field", func() {
							session := helpers.CF("service", serviceInstanceName)
							Eventually(session).Should(Say(`name:\s+%s`, serviceInstanceName))
							Eventually(session).Should(Say("bound apps:"))
							Eventually(session).Should(Say(`name\s+binding name\s+status\s+message`))
							Eventually(session).Should(Say(`%s\s+%s\s+create succeeded`, appName1, bindingName1))
							Eventually(session).Should(Say(`%s\s+%s\s+create succeeded`, appName2, bindingName2))

							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})
	})

	Context("service instance sharing when there are multiple spaces", func() {
		var (
			orgName         string
			sourceSpaceName string

			service     string
			servicePlan string
			broker      helpers.ServiceBroker
		)

		BeforeEach(func() {
			helpers.SkipIfVersionLessThan(ccversion.MinVersionShareServiceV3)
			orgName = helpers.NewOrgName()
			sourceSpaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, sourceSpaceName)

			domain := helpers.DefaultSharedDomain()
			service = helpers.PrefixedRandomName("SERVICE")
			servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")
			broker = helpers.NewServiceBroker(helpers.NewServiceBrokerName(), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
			broker.Push()
			broker.Configure(true)
			broker.Create()

			Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
			Eventually(helpers.CF("create-service", service, servicePlan, serviceInstanceName)).Should(Exit(0))
		})

		AfterEach(func() {
			// need to login as admin
			helpers.LoginCF()
			helpers.TargetOrgAndSpace(orgName, sourceSpaceName)
			broker.Destroy()
			helpers.QuickDeleteOrg(orgName)
		})

		Context("service has no type of shares", func() {
			When("the service is shareable", func() {
				It("should not display shared from or shared with information, but DOES display not currently shared info", func() {
					session := helpers.CF("service", serviceInstanceName)
					Eventually(session).Should(Say("This service is not currently shared."))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("service is shared between two spaces", func() {
			var (
				targetSpaceName string
			)

			BeforeEach(func() {
				targetSpaceName = helpers.NewSpaceName()
				helpers.CreateOrgAndSpace(orgName, targetSpaceName)
				helpers.TargetOrgAndSpace(orgName, sourceSpaceName)
				Eventually(helpers.CF("share-service", serviceInstanceName, "-s", targetSpaceName)).Should(Exit(0))
			})

			When("the user is targeted to the source space", func() {
				When("there are externally bound apps to the service", func() {
					BeforeEach(func() {
						helpers.TargetOrgAndSpace(orgName, targetSpaceName)
						helpers.WithHelloWorldApp(func(appDir string) {
							appName1 := helpers.NewAppName()
							Eventually(helpers.CF("push", appName1, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
							Eventually(helpers.CF("bind-service", appName1, serviceInstanceName)).Should(Exit(0))

							appName2 := helpers.NewAppName()
							Eventually(helpers.CF("push", appName2, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
							Eventually(helpers.CF("bind-service", appName2, serviceInstanceName)).Should(Exit(0))
						})
						helpers.TargetOrgAndSpace(orgName, sourceSpaceName)
					})

					It("should display the number of bound apps next to the target space name", func() {
						session := helpers.CF("service", serviceInstanceName)
						Eventually(session).Should(Say("shared with spaces:"))
						Eventually(session).Should(Say(`org\s+space\s+bindings`))
						Eventually(session).Should(Say(`%s\s+%s\s+2`, orgName, targetSpaceName))
						Eventually(session).Should(Exit(0))
					})
				})

				When("there are no externally bound apps to the service", func() {
					It("should NOT display the number of bound apps next to the target space name", func() {
						session := helpers.CF("service", serviceInstanceName)
						Eventually(session).Should(Say("shared with spaces:"))
						Eventually(session).Should(Say(`org\s+space\s+bindings`))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the service is no longer shareable", func() {
					Context("due to global settings", func() {
						BeforeEach(func() {
							helpers.DisableFeatureFlag("service_instance_sharing")
						})

						AfterEach(func() {
							helpers.EnableFeatureFlag("service_instance_sharing")
						})

						It("should display that the service instance feature flag is disabled", func() {
							session := helpers.CF("service", serviceInstanceName)
							Eventually(session).Should(Say(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform.`))
							Eventually(session).Should(Exit(0))
						})
					})

					Context("due to service broker settings", func() {
						BeforeEach(func() {
							broker.Configure(false)
							broker.Update()
						})

						It("should display that service instance sharing is disabled for this service", func() {
							session := helpers.CF("service", serviceInstanceName)
							Eventually(session).Should(Say("Service instance sharing is disabled for this service."))
							Eventually(session).Should(Exit(0))
						})
					})

					Context("due to global settings AND service broker settings", func() {
						BeforeEach(func() {
							helpers.DisableFeatureFlag("service_instance_sharing")
							broker.Configure(false)
							broker.Update()
						})

						AfterEach(func() {
							helpers.EnableFeatureFlag("service_instance_sharing")
						})

						It("should display that service instance sharing is disabled for this service", func() {
							session := helpers.CF("service", serviceInstanceName)
							Eventually(session).Should(Say(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform. Also, service instance sharing is disabled for this service.`))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})

			When("the user is targeted to the target space", func() {
				var appName1, appName2 string

				BeforeEach(func() {
					// We test that the app names are listed in alphanumeric sort order
					appName1 = helpers.PrefixedRandomName("2-INTEGRATION-APP")
					appName2 = helpers.PrefixedRandomName("1-INTEGRATION-APP")
					helpers.TargetOrgAndSpace(orgName, targetSpaceName)
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName1, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
						Eventually(helpers.CF("bind-service", appName1, serviceInstanceName)).Should(Exit(0))

						Eventually(helpers.CF("push", appName2, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
						Eventually(helpers.CF("bind-service", appName2, serviceInstanceName)).Should(Exit(0))
					})
				})

				When("there are bound apps to the service with no binding names", func() {
					It("should display the bound apps in alphanumeric sort order", func() {
						session := helpers.CF("service", serviceInstanceName)
						Eventually(session).Should(Say(`shared from org/space:\s+%s / %s`, orgName, sourceSpaceName))
						Eventually(session).Should(Say("\n\n"))
						Eventually(session).Should(Say("bound apps:"))
						Eventually(session).Should(Say(`name\s+binding name\s+status\s+message`))
						Eventually(session).Should(Say(appName2))
						Eventually(session).Should(Say(appName1))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
