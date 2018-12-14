package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("enable service access command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("enable-service-access", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("\\s+enable-service-access - Enable access to a service or service plan for one or all orgs"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("\\s+cf enable-service-access SERVICE \\[-p PLAN\\] \\[-o ORG\\]"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("\\s+\\-o\\s+Enable access for a specified organization"))
				Eventually(session).Should(Say("\\s+\\-p\\s+Enable access to a specified service plan"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("\\s+marketplace, service-access, service-brokers"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("no service argument was provided", func() {
		It("displays a warning, the help text, and exits 1", func() {
			session := helpers.CF("enable-service-access")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SERVICE` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("\\s+enable-service-access - Enable access to a service or service plan for one or all orgs"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("\\s+cf enable-service-access SERVICE \\[-p PLAN\\] \\[-o ORG\\]"))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say("\\s+\\-o\\s+Enable access for a specified organization"))
			Eventually(session).Should(Say("\\s+\\-p\\s+Enable access to a specified service plan"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("\\s+marketplace, service-access, service-brokers"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("two services arguments are provided", func() {
		It("displays an error, and exits 1", func() {
			session := helpers.CF("enable-service-access", "a-service", "another-service")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say(`Incorrect Usage: unexpected argument "another-service"`))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("\\s+enable-service-access - Enable access to a service or service plan for one or all orgs"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("\\s+cf enable-service-access SERVICE \\[-p PLAN\\] \\[-o ORG\\]"))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say("\\s+\\-o\\s+Enable access for a specified organization"))
			Eventually(session).Should(Say("\\s+\\-p\\s+Enable access to a specified service plan"))
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
			session := helpers.CF("enable-service-access", "does-not-matter")
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
					session := helpers.CF("enable-service-access", "some-service")
					Eventually(session).Should(Say("Enabling access to all plans of service some-service for all orgs as admin\\.\\.\\."))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Service offering 'some-service' not found"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("a service and an org are specified", func() {
				It("displays FAILED, an informative error message, and exits 1", func() {
					session := helpers.CF("enable-service-access", "some-service", "-o", "some-org")
					Eventually(session).Should(Say("Enabling access to all plans of service some-service for the org some-org as admin\\.\\.\\."))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Service offering 'some-service' not found"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("a service and a plan are specified", func() {
				It("displays FAILED, an informative error message, and exits 1", func() {
					session := helpers.CF("enable-service-access", "some-service", "-p", "some-plan")
					Eventually(session).Should(Say("Enabling access of plan some-plan for service some-service as admin\\.\\.\\."))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Service offering 'some-service' not found"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("a service, a plan and an org are specified", func() {
				It("displays FAILED, an informative error message, and exits 1", func() {
					session := helpers.CF("enable-service-access", "some-service", "-p", "some-plan", "-o", "some-org")
					Eventually(session).Should(Say("Enabling access to plan some-plan of service some-service for org some-org as admin\\.\\.\\."))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Service offering 'some-service' not found"))
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

				broker = helpers.CreateBroker(domain, service, servicePlan)
			})

			AfterEach(func() {
				broker.Destroy()
				helpers.QuickDeleteOrg(orgName)
			})

			When("a service name is provided", func() {
				It("displays an informative message, exits 0, and enables the service for all orgs", func() {
					session := helpers.CF("enable-service-access", service)
					Eventually(session).Should(Say("Enabling access to all plans of service %s for all orgs as admin...", service))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("service-access", "-e", service)
					Eventually(session).Should(Exit(0))
					Eventually(session).Should(Say("broker:\\s+%s", broker.Name))
					Eventually(session).Should(Say("%s\\s+%s\\s+all",
						service,
						servicePlan,
					))
				})

				When("service is already enabled for an org", func() {
					BeforeEach(func() {
						session := helpers.CF("enable-service-access", service, "-p", servicePlan, "-o", orgName)
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})

					It("disables the limited access for that org", func() {
						session := helpers.CF("enable-service-access", service, "-p", servicePlan, "-v")
						Eventually(session).Should(Say("Enabling access of plan %s for service %s as admin...", servicePlan, service))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("service-access", "-e", service)
						Eventually(session).Should(Exit(0))
						Eventually(session).Should(Say("broker:\\s+%s", broker.Name))
						Eventually(session).Should(Say("%s\\s+%s\\s+all\\s",
							service,
							servicePlan,
						))

					})
				})
			})

			When("a service name and plan name are provided", func() {
				It("displays an informative message, exits 0, and enables the plan for all orgs", func() {
					session := helpers.CF("enable-service-access", service, "-p", servicePlan)
					Eventually(session).Should(Say("Enabling access of plan %s for service %s as admin...", servicePlan, service))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("service-access", "-e", service)
					Eventually(session).Should(Exit(0))
					Eventually(session).Should(Say("broker:\\s+%s", broker.Name))
					Eventually(session).Should(Say("%s\\s+%s\\s+all",
						service,
						servicePlan,
					))
				})
			})

			When("a service name and org is provided", func() {
				It("displays an informative message, and exits 0", func() {
					session := helpers.CF("enable-service-access", service, "-o", orgName)
					Eventually(session).Should(Say("Enabling access to all plans of service %s for the org %s as admin...", service, orgName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("a service name, plan name and org is provided", func() {
				It("displays an informative message, and exits 0", func() {
					session := helpers.CF("enable-service-access", service, "-p", servicePlan, "-o", orgName)
					Eventually(session).Should(Say("Enabling access to plan %s of service %s for org %s as admin...", servicePlan, service, orgName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("service-access", "-e", service)
					Eventually(session).Should(Exit(0))
					Eventually(session).Should(Say("broker:\\s+%s", broker.Name))
					Eventually(session).Should(Say("%s\\s+%s\\s+%s\\s+%s",
						service,
						servicePlan,
						"limited",
						orgName,
					))
				})
			})

			When("the org does not exist", func() {
				It("displays FAILED, an informative error message, and exits 1", func() {
					session := helpers.CF("enable-service-access", service, "-o", "not-a-real-org")
					Eventually(session).Should(Say("Enabling access to all plans of service %s for the org not-a-real-org as admin...", service))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Organization 'not-a-real-org' not found"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the plan does not exist", func() {
				It("displays FAILED, an informative error message, and exits 1", func() {
					session := helpers.CF("enable-service-access", service, "-p", "plan-does-not-exist")
					Eventually(session).Should(Say("Enabling access of plan plan-does-not-exist for service %s as admin...", service))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("The plan plan-does-not-exist could not be found for service %s", service))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the plan does exist and the org does not exist", func() {
				It("displays FAILED, an informative error message, and exits 1", func() {
					session := helpers.CF("enable-service-access", service, "-p", servicePlan, "-o", "not-a-real-org")
					Eventually(session).Should(Say("Enabling access to plan %s of service %s for org not-a-real-org as admin...", servicePlan, service))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Organization 'not-a-real-org' not found"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the plan does not exist and the org does exist", func() {
				It("displays FAILED, an informative error message, and exits 1", func() {
					session := helpers.CF("enable-service-access", service, "-p", "not-a-real-plan", "-o", orgName)
					Eventually(session).Should(Say("Enabling access to plan not-a-real-plan of service %s for org %s as admin...", service, orgName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Service plan 'not-a-real-plan' not found"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when access is already enabled in the org", func() {
				BeforeEach(func() {
					session := helpers.CF("enable-service-access", service, "-o", orgName)
					Eventually(session).Should(Say("OK"))
				})

				It("displays FAILED, an informative error message and exits 1", func() {
					session := helpers.CF("enable-service-access", service, "-o", orgName)
					Eventually(session).Should(Say("Enabling access to all plans of service %s for the org %s as admin...", service, orgName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("^This combination of ServicePlan and Organization is already taken: organization_id and service_plan_id unique"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when access is already globally enabled", func() {
				BeforeEach(func() {
					helpers.CF("enable-service-access", service)
				})

				When("the service already has access globally enabled", func() {
					It("displays an informative message, and exits 0", func() {
						session := helpers.CF("enable-service-access", service)
						Eventually(session).Should(Say("Enabling access to all plans of service %s for all orgs as admin...", service))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the plan already has access globally enabled", func() {
					It("displays an informative message, and exits 0", func() {
						session := helpers.CF("enable-service-access", service, "-p", servicePlan)
						Eventually(session).Should(Say("Enabling access of plan %s for service %s as admin...", servicePlan, service))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
