package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("purge-service-offering command", func() {
	BeforeEach(func() {
		helpers.TurnOnExperimental()
	})

	AfterEach(func() {
		helpers.TurnOffExperimental()
	})

	Describe("help", func() {
		When("the --help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("purge-service-offering", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("purge-service-offering - Recursively remove a service and child objects from Cloud Foundry database without making requests to a service broker"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf purge-service-offering SERVICE \[-p PROVIDER\] \[-f\]`))
				Eventually(session).Should(Say("WARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database\\. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings\\. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances\\. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup\\."))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("-f\\s+Force deletion without confirmation"))
				Eventually(session).Should(Say("-p\\s+Provider"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("marketplace, purge-service-instance, service-brokers"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("no args are passed", func() {
		It("displays an error message with help text", func() {
			session := helpers.CF("purge-service-offering")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SERVICE` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("purge-service-offering - Recursively remove a service and child objects from Cloud Foundry database without making requests to a service broker"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`cf purge-service-offering SERVICE \[-p PROVIDER\] \[-f\]`))
			Eventually(session).Should(Say("WARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database\\. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings\\. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances\\. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup\\."))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say("-f\\s+Force deletion without confirmation"))
			Eventually(session).Should(Say("-p\\s+Provider"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("marketplace, purge-service-instance, service-brokers"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("more than required number of args are passed", func() {
		It("displays an error message with help text and exits 1", func() {
			session := helpers.CF("purge-service-offering", "service-name", "extra")

			Eventually(session.Err).Should(Say(`Incorrect Usage: unexpected argument "extra"`))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("purge-service-offering - Recursively remove a service and child objects from Cloud Foundry database without making requests to a service broker"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`cf purge-service-offering SERVICE \[-p PROVIDER\] \[-f\]`))
			Eventually(session).Should(Say("WARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database\\. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings\\. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances\\. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup\\."))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say("-f\\s+Force deletion without confirmation"))
			Eventually(session).Should(Say("-p\\s+Provider"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("marketplace, purge-service-instance, service-brokers"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("service name is passed", func() {
		When("an API target is not set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("displays an error message that no API endpoint is set", func() {
				session := helpers.CF("purge-service-offering", "service-name")

				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("user is not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("displays an error message that user is not logged in", func() {
				session := helpers.CF("purge-service-offering", "service-name")

				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`Not logged in\. Use 'cf login' to log in\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("user is logged in", func() {
			BeforeEach(func() {
				helpers.LoginCF()
			})

			When("the service exists", func() {
				var (
					orgName     string
					spaceName   string
					domain      string
					service     string
					servicePlan string
					broker      helpers.ServiceBroker
					buffer      *Buffer
				)

				BeforeEach(func() {
					buffer = NewBuffer()

					orgName = helpers.NewOrgName()
					spaceName = helpers.NewSpaceName()
					helpers.SetupCF(orgName, spaceName)

					domain = helpers.DefaultSharedDomain()
					service = helpers.PrefixedRandomName("SERVICE")
					servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")

					broker = helpers.NewServiceBroker(helpers.NewServiceBrokerName(), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
					broker.Push()
					broker.Configure(true)
					broker.Create()
				})

				AfterEach(func() {
					broker.Destroy()
					helpers.QuickDeleteOrg(orgName)
				})

				When("the user enters 'y'", func() {
					BeforeEach(func() {
						buffer.Write([]byte("y\n"))
					})

					It("purges the service offering, asking for confirmation", func() {
						session := helpers.CFWithStdin(buffer, "purge-service-offering", service)

						Eventually(session).Should(Say("WARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database\\. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings\\. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances\\. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup\\."))
						Eventually(session).Should(Say("Really purge service offering %s from Cloud Foundry?", service))
						Eventually(session).Should(Say("Purging service %s...", service))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("marketplace")
						Eventually(session).Should(Say("OK"))
						Consistently(session).ShouldNot(Say(service))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the user enters something other than 'y' or 'yes'", func() {
					BeforeEach(func() {
						buffer.Write([]byte("wat\n\n"))
					})

					It("asks again", func() {
						session := helpers.CFWithStdin(buffer, "purge-service-offering", service)

						Eventually(session).Should(Say("WARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database\\. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings\\. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances\\. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup\\."))
						Eventually(session).Should(Say("Really purge service offering %s from Cloud Foundry?", service))
						Eventually(session).Should(Say(`invalid input \(not y, n, yes, or no\)`))
						Eventually(session).Should(Say("Really purge service offering %s from Cloud Foundry?", service))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the user enters 'n' or 'no'", func() {
					BeforeEach(func() {
						buffer.Write([]byte("n\n"))
					})

					It("does not purge the service offering", func() {
						session := helpers.CFWithStdin(buffer, "purge-service-offering", service)

						Eventually(session).Should(Say("WARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database\\. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings\\. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances\\. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup\\."))
						Eventually(session).Should(Say("Really purge service offering %s from Cloud Foundry?", service))
						Eventually(session).ShouldNot(Say("Purging service %s...", service))
						Eventually(session).ShouldNot(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the -f flag is provided", func() {
					It("purges the service offering without asking for confirmation", func() {
						session := helpers.CF("purge-service-offering", service, "-f")

						Eventually(session).Should(Say("WARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database\\. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings\\. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances\\. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup\\."))
						Eventually(session).Should(Say("Purging service %s...", service))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("the service does not exist", func() {
				It("prints a message the service offering does not exist, exiting 0", func() {
					session := helpers.CF("purge-service-offering", "missing-service")

					Eventually(session).Should(Say("Service offering 'missing-service' not found"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the -p flag is provided", func() {
				It("prints a warning that this flag is no longer supported", func() {
					session := helpers.CF("purge-service-offering", "some-service", "-p", "some-provider")

					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Flag '-p' is no longer supported"))
					Eventually(session).ShouldNot(Say("Purging service"))
					Eventually(session).ShouldNot(Say("OK"))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
