package isolated

import (
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
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
				Eventually(session).Should(Say("cf marketplace \\[-s SERVICE\\]"))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("m"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("-s\\s+Show plan details for a particular service offering"))
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
						broker1 helpers.ServiceBroker
						broker2 helpers.ServiceBroker
						org     string
						space   string
					)

					BeforeEach(func() {
						org = helpers.NewOrgName()
						space = helpers.NewSpaceName()
						helpers.SetupCF(org, space)
						helpers.TargetOrgAndSpace(org, space)

						domain := helpers.DefaultSharedDomain()

						broker1 = helpers.CreateBroker(domain, helpers.PrefixedRandomName("SERVICE-1"), "SERVICE-PLAN-1")
						enableServiceAccess(broker1)
						broker2 = helpers.CreateBroker(domain, helpers.PrefixedRandomName("SERVICE-2"), "SERVICE-PLAN-2")
						enableServiceAccess(broker2)

						helpers.LogoutCF()
					})

					AfterEach(func() {
						helpers.SetupCF(org, space)
						broker1.Destroy()
						broker2.Destroy()
						helpers.QuickDeleteOrg(org)
					})

					It("displays a table of all available services and a tip", func() {
						session := helpers.CF("marketplace")
						Eventually(session).Should(Say("Getting all services from marketplace"))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say("\n\n"))
						Eventually(session).Should(Say("service\\s+plans\\s+description"))
						Eventually(session).Should(Say("%s\\s+%s\\s+fake service", getServiceName(broker1), getBrokerPlanNames(broker1)))
						Eventually(session).Should(Say("%s\\s+%s\\s+fake service", getServiceName(broker2), getBrokerPlanNames(broker2)))
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
						broker1      helpers.ServiceBroker
						org1, space1 string

						broker2      helpers.ServiceBroker
						org2, space2 string
					)

					BeforeEach(func() {
						org1 = helpers.NewOrgName()
						space1 = helpers.NewSpaceName()
						helpers.SetupCF(org1, space1)
						helpers.TargetOrgAndSpace(org1, space1)

						domain := helpers.DefaultSharedDomain()

						broker1 = helpers.CreateBroker(domain, helpers.PrefixedRandomName("SERVICE-1"), "SERVICE-PLAN-1")
						enableServiceAccessForOrg(broker1, org1)

						org2 = helpers.NewOrgName()
						space2 = helpers.NewSpaceName()
						helpers.CreateOrgAndSpace(org2, space2)
						helpers.TargetOrgAndSpace(org2, space2)

						broker2 = helpers.CreateBroker(domain, helpers.PrefixedRandomName("SERVICE-2"), "SERVICE-PLAN-2")
						enableServiceAccess(broker2)
					})

					AfterEach(func() {
						helpers.TargetOrgAndSpace(org2, space2)
						broker2.Destroy()
						helpers.QuickDeleteOrg(org2)

						helpers.TargetOrgAndSpace(org1, space1)
						broker1.Destroy()
						helpers.QuickDeleteOrg(org1)
					})

					When("cc api version < 2.125.0", func() {
						BeforeEach(func() {
							helpers.SkipIfVersionAtLeast(ccversion.MinVersionServiceBrokerNameV2)
						})

						It("displays a table and tip that does not include that service", func() {
							session := helpers.CF("marketplace")
							Eventually(session).Should(Say("Getting services from marketplace in org %s / space %s as %s\\.\\.\\.", org2, space2, user))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("\n\n"))
							Eventually(session).Should(Say("service\\s+plans\\s+description\\s+broker"))
							Consistently(session).ShouldNot(Say(getServiceName(broker1)))
							Consistently(session).ShouldNot(Say(getBrokerPlanNames(broker1)))
							Eventually(session).Should(Say("%s\\s+%s\\s+fake service\\s*", getServiceName(broker2), getBrokerPlanNames(broker2)))
							Eventually(session).Should(Say("TIP: Use 'cf marketplace -s SERVICE' to view descriptions of individual plans of a given service."))
							Eventually(session).Should(Exit(0))
						})
					})

					When("cc api version >= 2.125.0", func() {
						BeforeEach(func() {
							helpers.SkipIfVersionLessThan(ccversion.MinVersionServiceBrokerNameV2)
						})

						It("displays a table with broker name", func() {
							session := helpers.CF("marketplace")
							Eventually(session).Should(Say("Getting services from marketplace in org %s / space %s as %s\\.\\.\\.", org2, space2, user))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("\n\n"))
							Eventually(session).Should(Say("service\\s+plans\\s+description\\s+broker"))
							Consistently(session).ShouldNot(Say(getServiceName(broker1)))
							Consistently(session).ShouldNot(Say(getBrokerPlanNames(broker1)))
							Consistently(session).ShouldNot(Say(broker1.Name))
							Eventually(session).Should(Say("%s\\s+%s\\s+fake service\\s+%s", getServiceName(broker2), getBrokerPlanNames(broker2), broker2.Name))
							Eventually(session).Should(Say("TIP: Use 'cf marketplace -s SERVICE' to view descriptions of individual plans of a given service."))
							Eventually(session).Should(Exit(0))
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
							broker helpers.ServiceBroker
							org    string
							space  string
						)

						BeforeEach(func() {
							org = helpers.NewOrgName()
							space = helpers.NewSpaceName()
							helpers.SetupCF(org, space)
							helpers.TargetOrgAndSpace(org, space)

							domain := helpers.DefaultSharedDomain()

							broker = helpers.CreateBroker(domain, helpers.PrefixedRandomName("SERVICE"), "SERVICE-PLAN")
							enableServiceAccess(broker)

							helpers.LogoutCF()
						})

						AfterEach(func() {
							helpers.LoginCF()
							helpers.TargetOrgAndSpace(org, space)
							broker.Destroy()
							helpers.QuickDeleteOrg(org)
						})

						It("displays extended information about the service", func() {
							description := "Shared fake Server, 5tb persistent disk, 40 max concurrent connections"
							session := helpers.CF("marketplace", "-s", getServiceName(broker))
							Eventually(session).Should(Say("Getting service plan information for service %s\\.\\.\\.", getServiceName(broker)))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("\n\n"))
							Eventually(session).Should(Say("service plan\\s+description\\s+free or paid"))
							Eventually(session).Should(Say("%s\\s+%s\\s+%s", getPlanName(broker), description, "free"))
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
								broker helpers.ServiceBroker
								org    string
								space  string
							)

							BeforeEach(func() {
								org = helpers.NewOrgName()
								space = helpers.NewSpaceName()
								helpers.SetupCF(org, space)
								helpers.TargetOrgAndSpace(org, space)

								domain := helpers.DefaultSharedDomain()

								broker = helpers.CreateBroker(domain, helpers.PrefixedRandomName("SERVICE"), "SERVICE-PLAN")
								enableServiceAccess(broker)
							})

							AfterEach(func() {
								broker.Destroy()
								helpers.QuickDeleteOrg(org)
							})

							It("displays extended information about the service", func() {
								description := "Shared fake Server, 5tb persistent disk, 40 max concurrent connections"
								session := helpers.CF("marketplace", "-s", getServiceName(broker))
								Eventually(session).Should(Say("Getting service plan information for service %s as %s\\.\\.\\.", getServiceName(broker), user))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Say("\n\n"))
								Eventually(session).Should(Say("service plan\\s+description\\s+free or paid"))
								Eventually(session).Should(Say("%s\\s+%s\\s+%s", getPlanName(broker), description, "free"))
								Eventually(session).Should(Exit(0))
							})
						})

						When("the specified service is accessible but not in the targeted space", func() {
							var (
								broker helpers.ServiceBroker
								org    string
								space  string
							)

							BeforeEach(func() {
								org = helpers.NewOrgName()
								space = helpers.NewSpaceName()
								helpers.SetupCF(org, space)
								helpers.TargetOrgAndSpace(org, space)

								domain := helpers.DefaultSharedDomain()

								broker = helpers.CreateBroker(domain, helpers.PrefixedRandomName("SERVICE"), "SERVICE-PLAN")
								enableServiceAccessForOrg(broker, org)

								helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)
							})

							AfterEach(func() {
								helpers.TargetOrgAndSpace(org, space)
								broker.Destroy()
								helpers.QuickDeleteOrg(org)
							})

							It("displays an error that the service doesn't exist", func() {
								session := helpers.CF("marketplace", "-s", getServiceName(broker))
								Eventually(session).Should(Say("Getting service plan information for service %s as %s\\.\\.\\.", getServiceName(broker), user))
								Eventually(session).Should(Say("FAILED"))
								Eventually(session.Err).Should(Say("Service offering '%s' not found", getServiceName(broker)))
								Eventually(session).Should(Exit(1))
							})
						})
					})
				})
			})
		})
	})
})

func enableServiceAccess(broker helpers.ServiceBroker) {
	Eventually(helpers.CF("enable-service-access", getServiceName(broker))).Should(Exit(0))
}

func enableServiceAccessForOrg(broker helpers.ServiceBroker, orgName string) {
	Eventually(helpers.CF("enable-service-access", getServiceName(broker), "-o", orgName)).Should(Exit(0))
}

func getServiceName(broker helpers.ServiceBroker) string {
	return broker.Service.Name
}

func getPlanName(broker helpers.ServiceBroker) string {
	return broker.SyncPlans[0].Name
}

func getBrokerPlanNames(broker helpers.ServiceBroker) string {
	return strings.Join(plansToNames(append(broker.SyncPlans, broker.AsyncPlans...)), ", ")
}

func plansToNames(plans []helpers.Plan) []string {
	planNames := []string{}
	for _, plan := range plans {
		planNames = append(planNames, plan.Name)
	}
	return planNames
}
