package isolated

import (
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("marketplace command", func() {
	Describe("help", func() {
		When("the --help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("marketplace", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("marketplace - List available offerings in the marketplace"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf marketplace \\[-s SERVICE\\] \\[--no-plans\\]"))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("m"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("-s\\s+Show plan details for a particular service offering"))
				Eventually(session).Should(Say("--no-plans\\s+Hide plan information for service offerings"))
				Eventually(session).Should(Say("create-service, services"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("no flags are passed", func() {
		When("an API target is not set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("displays an error message that no API endpoint is set", func() {
				session := helpers.CF("marketplace")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("an API endpoint is set", func() {
			When("not logged in", func() {
				When("there are accessible services", func() {
					var (
						broker1 *servicebrokerstub.ServiceBrokerStub
						broker2 *servicebrokerstub.ServiceBrokerStub
						org     string
						space   string
					)

					BeforeEach(func() {
						org = helpers.NewOrgName()
						space = helpers.NewSpaceName()
						helpers.SetupCF(org, space)
						helpers.TargetOrgAndSpace(org, space)

						broker1 = servicebrokerstub.New().WithPlans(2).EnableServiceAccess()
						broker2 = servicebrokerstub.New().WithPlans(2).EnableServiceAccess()

						helpers.LogoutCF()
					})

					AfterEach(func() {
						helpers.SetupCF(org, space)
						broker1.Forget()
						broker2.Forget()
						helpers.QuickDeleteOrg(org)
					})

					It("displays a table of all available services and a tip", func() {
						session := helpers.CF("marketplace")
						Eventually(session).Should(Say("Getting all services from marketplace"))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say("\n\n"))
						Eventually(session).Should(Say("service\\s+plans\\s+description"))
						Eventually(session).Should(Say("%s\\s+%s\\s+%s", broker1.FirstServiceOfferingName(), planNamesOf(broker1), broker1.FirstServiceOfferingDescription()))
						Eventually(session).Should(Say("%s\\s+%s\\s+%s", broker2.FirstServiceOfferingName(), planNamesOf(broker2), broker2.FirstServiceOfferingDescription()))
						Eventually(session).Should(Say("TIP: Use 'cf marketplace -s SERVICE' to view descriptions of individual plans of a given service."))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("logged in", func() {
				var user string

				BeforeEach(func() {
					helpers.LoginCF()
					user, _ = helpers.GetCredentials()
				})

				When("no space is targeted", func() {
					BeforeEach(func() {
						helpers.TargetOrg(ReadOnlyOrg)
					})

					It("displays an error that a space must be targeted", func() {
						session := helpers.CF("marketplace")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Cannot list marketplace services without a targeted space"))
						Eventually(session).Should(Exit(1))
					})
				})

				When("a service is accessible but not in the currently targeted org", func() {
					var (
						broker1      *servicebrokerstub.ServiceBrokerStub
						org1, space1 string

						broker2      *servicebrokerstub.ServiceBrokerStub
						org2, space2 string
					)

					BeforeEach(func() {
						org1 = helpers.NewOrgName()
						space1 = helpers.NewSpaceName()
						helpers.SetupCF(org1, space1)
						helpers.TargetOrgAndSpace(org1, space1)

						broker1 = servicebrokerstub.New().WithPlans(2).Register()
						enableServiceAccessForOrg(broker1, org1)

						org2 = helpers.NewOrgName()
						space2 = helpers.NewSpaceName()
						helpers.CreateOrgAndSpace(org2, space2)
						helpers.TargetOrgAndSpace(org2, space2)

						broker2 = servicebrokerstub.New().WithPlans(2).EnableServiceAccess()
					})

					AfterEach(func() {
						broker2.Forget()
						helpers.QuickDeleteOrg(org2)

						broker1.Forget()
						helpers.QuickDeleteOrg(org1)
					})

					When("CC API does not return broker names in response", func() {
						It("displays a table and tip that does not include that service", func() {
							session := helpers.CF("marketplace")
							Eventually(session).Should(Say("Getting services from marketplace in org %s / space %s as %s\\.\\.\\.", org2, space2, user))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("\n\n"))
							Eventually(session).Should(Say("service\\s+plans\\s+description\\s+broker"))
							Consistently(session).ShouldNot(Say(broker1.FirstServiceOfferingName()))
							Consistently(session).ShouldNot(Say(planNamesOf(broker1)))
							Eventually(session).Should(Say("%s\\s+%s\\s+%s\\s*", broker2.FirstServiceOfferingName(), planNamesOf(broker2), broker2.FirstServiceOfferingDescription()))
							Eventually(session).Should(Say("TIP: Use 'cf marketplace -s SERVICE' to view descriptions of individual plans of a given service."))
							Eventually(session).Should(Exit(0))
						})
					})

					When("CC API returns broker names in response", func() {
						It("displays a table with broker name", func() {
							session := helpers.CF("marketplace")
							Eventually(session).Should(Say("Getting services from marketplace in org %s / space %s as %s\\.\\.\\.", org2, space2, user))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("\n\n"))
							Eventually(session).Should(Say("service\\s+plans\\s+description\\s+broker"))
							Consistently(session).ShouldNot(Say(broker1.FirstServiceOfferingName()))
							Consistently(session).ShouldNot(Say(planNamesOf(broker1)))
							Consistently(session).ShouldNot(Say(broker1.Name))
							Eventually(session).Should(Say("%s\\s+%s\\s+%s\\s+%s", broker2.FirstServiceOfferingName(), planNamesOf(broker2), broker2.FirstServiceOfferingDescription(), broker2.Name))
							Eventually(session).Should(Say("TIP: Use 'cf marketplace -s SERVICE' to view descriptions of individual plans of a given service."))
							Eventually(session).Should(Exit(0))
						})

						When("--no-plans is passed", func() {
							It("displays a table with broker name", func() {
								session := helpers.CF("marketplace", "--no-plans")
								Eventually(session).Should(Say("Getting services from marketplace in org %s / space %s as %s\\.\\.\\.", org2, space2, user))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Say("\n\n"))
								Eventually(session).Should(Say("service\\s+description\\s+broker"))
								Eventually(session).Should(Say("%s\\s+%s\\s+%s", broker2.FirstServiceOfferingName(), broker2.FirstServiceOfferingDescription(), broker2.Name))
								Eventually(session).Should(Say("TIP: Use 'cf marketplace -s SERVICE' to view descriptions of individual plans of a given service."))
								Eventually(session).Should(Exit(0))
							})
						})
					})
				})
			})
		})

		When("the -s flag is passed", func() {
			When("an api endpoint is not set", func() {
				BeforeEach(func() {
					helpers.UnsetAPI()
				})

				It("displays an error message that no API endpoint is set", func() {
					session := helpers.CF("marketplace")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the api is set", func() {
				When("not logged in", func() {
					BeforeEach(func() {
						helpers.LogoutCF()
					})

					When("the specified service does not exist", func() {
						It("displays an error that the service doesn't exist", func() {
							session := helpers.CF("marketplace", "-s", "not-found-service")
							Eventually(session).Should(Say("Getting service plan information for service not-found-service\\.\\.\\."))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Service offering 'not-found-service' not found"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("the specified service exists", func() {
						var (
							broker *servicebrokerstub.ServiceBrokerStub
							org    string
							space  string
						)

						BeforeEach(func() {
							org = helpers.NewOrgName()
							space = helpers.NewSpaceName()
							helpers.SetupCF(org, space)

							broker = servicebrokerstub.New().WithPlans(2).EnableServiceAccess()

							helpers.LogoutCF()
						})

						AfterEach(func() {
							helpers.LoginCF()
							helpers.TargetOrgAndSpace(org, space)
							broker.Forget()
							helpers.QuickDeleteOrg(org)
						})

						It("displays extended information about the service", func() {
							session := helpers.CF("marketplace", "-s", broker.FirstServiceOfferingName())
							Eventually(session).Should(Say("Getting service plan information for service %s\\.\\.\\.", broker.FirstServiceOfferingName()))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("\n\n"))
							Eventually(session).Should(Say("service plan\\s+description\\s+free or paid"))
							Eventually(session).Should(Say("%s\\s+%s\\s+%s", broker.FirstServicePlanName(), broker.FirstServicePlanDescription(), "free"))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("logged in", func() {
					var user string

					BeforeEach(func() {
						helpers.LoginCF()
						user, _ = helpers.GetCredentials()
					})

					When("no space is targeted", func() {
						BeforeEach(func() {
							helpers.TargetOrg(ReadOnlyOrg)
						})

						It("displays an error that a space must be targeted", func() {
							session := helpers.CF("marketplace", "-s", "service")
							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Cannot list plan information for service without a targeted space"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("a space is targeted", func() {
						BeforeEach(func() {
							helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)
						})

						When("the specified service does not exist", func() {
							It("displays an error that the service doesn't exist", func() {
								session := helpers.CF("marketplace", "-s", "not-found-service")
								Eventually(session).Should(Say("Getting service plan information for service not-found-service as %s\\.\\.\\.", user))
								Eventually(session).Should(Say("FAILED"))
								Eventually(session.Err).Should(Say("Service offering 'not-found-service' not found"))
								Eventually(session).Should(Exit(1))
							})
						})

						When("the specified service exists", func() {
							var (
								broker *servicebrokerstub.ServiceBrokerStub
								org    string
								space  string
							)

							BeforeEach(func() {
								org = helpers.NewOrgName()
								space = helpers.NewSpaceName()
								helpers.SetupCF(org, space)

								broker = servicebrokerstub.New().WithPlans(2).EnableServiceAccess()
							})

							AfterEach(func() {
								broker.Forget()
								helpers.QuickDeleteOrg(org)
							})

							It("displays extended information about the service", func() {
								session := helpers.CF("marketplace", "-s", broker.FirstServiceOfferingName())
								Eventually(session).Should(Say("Getting service plan information for service %s as %s\\.\\.\\.", broker.FirstServiceOfferingName(), user))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Say("\n\n"))
								Eventually(session).Should(Say("service plan\\s+description\\s+free or paid"))
								Eventually(session).Should(Say("%s\\s+%s\\s+%s", broker.FirstServicePlanName(), broker.FirstServicePlanDescription(), "free"))
								Eventually(session).Should(Exit(0))
							})
						})

						When("the specified service is accessible but not in the targeted space", func() {
							var (
								broker *servicebrokerstub.ServiceBrokerStub
								org    string
								space  string
							)

							BeforeEach(func() {
								org = helpers.NewOrgName()
								space = helpers.NewSpaceName()
								helpers.SetupCF(org, space)

								broker = servicebrokerstub.New().WithPlans(2).Register()
								enableServiceAccessForOrg(broker, org)

								helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)
							})

							AfterEach(func() {
								broker.Forget()
								helpers.QuickDeleteOrg(org)
							})

							It("displays an error that the service doesn't exist", func() {
								session := helpers.CF("marketplace", "-s", broker.FirstServiceOfferingName())
								Eventually(session).Should(Say("Getting service plan information for service %s as %s\\.\\.\\.", broker.FirstServiceOfferingName(), user))
								Eventually(session).Should(Say("FAILED"))
								Eventually(session.Err).Should(Say("Service offering '%s' not found", broker.FirstServiceOfferingName()))
								Eventually(session).Should(Exit(1))
							})
						})
					})
				})
			})
		})
	})
})

func enableServiceAccessForOrg(broker *servicebrokerstub.ServiceBrokerStub, orgName string) {
	Eventually(helpers.CF("enable-service-access", broker.FirstServiceOfferingName(), "-o", orgName)).Should(Exit(0))
}

func planNamesOf(broker *servicebrokerstub.ServiceBrokerStub) string {
	var planNames []string
	for _, p := range broker.Services[0].Plans {
		planNames = append(planNames, p.Name)
	}
	return strings.Join(planNames, ", ")
}
