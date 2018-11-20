package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("disable service access command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("disable-service-access", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("\\s+disable-service-access - Disable access to a service or service plan for one or all orgs"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("\\s+cf disable-service-access SERVICE \\[-p PLAN\\] \\[-o ORG\\]"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("\\s+\\-o\\s+Disable access for a specified organization"))
				Eventually(session).Should(Say("\\s+\\-p\\s+Disable access to a specified service plan"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("\\s+marketplace, service-access, service-brokers"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("no service argument was provided", func() {
		It("displays a warning, the help text, and exits 1", func() {
			session := helpers.CF("disable-service-access")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SERVICE` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("\\s+disable-service-access - Disable access to a service or service plan for one or all orgs"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("\\s+cf disable-service-access SERVICE \\[-p PLAN\\] \\[-o ORG\\]"))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say("\\s+\\-o\\s+Disable access for a specified organization"))
			Eventually(session).Should(Say("\\s+\\-p\\s+Disable access to a specified service plan"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("\\s+marketplace, service-access, service-brokers"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("an extra argument is provided", func() {
		It("displays an error, and exits 1", func() {
			session := helpers.CF("disable-service-access", "a-service", "extra-arg")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say(`Incorrect Usage: unexpected argument "extra-arg"`))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("\\s+disable-service-access - Disable access to a service or service plan for one or all orgs"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("\\s+cf disable-service-access SERVICE \\[-p PLAN\\] \\[-o ORG\\]"))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say("\\s+\\-o\\s+Disable access for a specified organization"))
			Eventually(session).Should(Say("\\s+\\-p\\s+Disable access to a specified service plan"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("\\s+marketplace, service-access, service-brokers"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("not logged in", func() {
		BeforeEach(func() {
			helpers.LogoutCF()
		})

		It("displays FAILED, an informative error message, and exits 1", func() {
			session := helpers.CF("disable-service-access", "does-not-matter")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in."))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("logged in", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		Context("the service does not exist", func() {
			When("only the service is specified", func() {
				It("displays FAILED, an informative error message, and exits 1", func() {
					session := helpers.CF("disable-service-access", "some-service")
					Eventually(session).Should(Say("Disabling access to all plans of service some-service for all orgs as admin\\.\\.\\."))
					Eventually(session.Err).Should(Say("Service offering 'some-service' not found"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Context("a service broker is registered", func() {
			var (
				orgName     string
				spaceName   string
				domain      string
				service     string
				servicePlan string
				broker      helpers.ServiceBroker
			)

			BeforeEach(func() {
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
				helpers.TargetOrgAndSpace(orgName, spaceName)
				broker.Destroy()
				helpers.QuickDeleteOrg(orgName)
			})

			When("a service has access enabled for all orgs and plans", func() {
				BeforeEach(func() {
					session := helpers.CF("enable-service-access", service)
					Eventually(session).Should(Exit(0))
				})

				When("a service name is provided", func() {
					It("displays an informative message, exits 0, and disables the service for all orgs", func() {
						session := helpers.CF("disable-service-access", service)
						Eventually(session).Should(Say("Disabling access to all plans of service %s for all orgs as admin...", service))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("service-access", "-e", service)
						Eventually(session).Should(Exit(0))
						Eventually(session).Should(Say("broker:\\s+%s", broker.Name))
						Eventually(session).Should(Say("%s\\s+%s\\s+none",
							service,
							servicePlan,
						))
					})
				})

				When("a service name and plan name are provided", func() {
					It("displays an informative message, exits 0, and disables the plan for all orgs", func() {
						session := helpers.CF("disable-service-access", service, "-p", servicePlan)
						Eventually(session).Should(Say("Disabling access of plan %s for service %s as admin...", servicePlan, service))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("service-access", "-e", service)
						Eventually(session).Should(Exit(0))
						Eventually(session).Should(Say("broker:\\s+%s", broker.Name))
						Eventually(session).Should(Say("%s\\s+%s\\s+none",
							service,
							servicePlan,
						))
						Eventually(session).Should(Say("%s\\s+.*\\s+all",
							service,
						))
					})
				})
			})

			When("a service has access enabled for multiple orgs and a specific plan", func() {
				var orgName2 string

				BeforeEach(func() {
					orgName2 = helpers.NewOrgName()
					spaceName2 := helpers.NewSpaceName()
					helpers.CreateOrgAndSpace(orgName2, spaceName2)
					Eventually(helpers.CF("enable-service-access", service, "-o", orgName, "-p", servicePlan)).Should(Exit(0))
					Eventually(helpers.CF("enable-service-access", service, "-o", orgName2, "-p", servicePlan)).Should(Exit(0))
				})

				AfterEach(func() {
					helpers.QuickDeleteOrg(orgName2)
				})

				When("a service name and org is provided", func() {
					It("displays an informative message, and exits 0, and disables the service for the given org", func() {
						session := helpers.CF("disable-service-access", service, "-o", orgName)
						Eventually(session).Should(Say("Disabling access to all plans of service %s for the org %s as admin...", service, orgName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("service-access", "-e", service)
						Eventually(session).Should(Exit(0))
						Eventually(session).Should(Say("broker:\\s+%s", broker.Name))
						Eventually(session).Should(Say("%s\\s+%s\\s+limited\\s+%s",
							service,
							servicePlan,
							orgName2,
						))
					})
				})

				When("a service name, plan name and org is provided", func() {
					It("displays an informative message, and exits 0, disables the service for the given org and plan", func() {
						session := helpers.CF("disable-service-access", service, "-p", servicePlan, "-o", orgName)
						Eventually(session).Should(Say("Disabling access to plan %s of service %s for org %s as admin...", servicePlan, service, orgName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("service-access", "-e", service)
						Eventually(session).Should(Exit(0))
						Eventually(session).Should(Say("broker:\\s+%s", broker.Name))
						Eventually(session).Should(Say("%s\\s+%s\\s+limited\\s+%s",
							service,
							servicePlan,
							orgName2,
						))
					})
				})
			})

			When("the org does not exist", func() {
				It("displays FAILED, an informative error message, and exits 1", func() {
					session := helpers.CF("disable-service-access", service, "-o", "not-a-real-org")
					Eventually(session).Should(Say("Disabling access to all plans of service %s for the org not-a-real-org as admin...", service))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Organization 'not-a-real-org' not found"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the plan does not exist", func() {
				It("displays FAILED, an informative error message, and exits 1", func() {
					session := helpers.CF("disable-service-access", service, "-p", "plan-does-not-exist")
					Eventually(session).Should(Say("Disabling access of plan plan-does-not-exist for service %s as admin...", service))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("The plan plan-does-not-exist could not be found for service %s", service))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
