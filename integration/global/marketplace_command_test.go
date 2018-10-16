package global

import (
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("marketplace command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
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

	Context("when no flags are passed", func() {
		Context("and the API target is not set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("prints an error message that the API is not set", func() {
				session := helpers.CF("marketplace")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the API is set", func() {
			Context("when not logged in", func() {
				Context("and no services exist", func() {
					BeforeEach(func() {
						helpers.LogoutCF()
					})

					It("prints a message saying no services exist", func() {
						session := helpers.CF("marketplace")
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say("No service offerings found"))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("and services are available", func() {
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

						broker1 = createBroker(domain, "SERVICE-1", "SERVICE-PLAN-1")
						enableServiceAccess(broker1)
						broker2 = createBroker(domain, "SERVICE-2", "SERVICE-PLAN-2")
						enableServiceAccess(broker2)

						helpers.LogoutCF()
					})

					AfterEach(func() {
						helpers.SetupCF(org, space)
						broker1.Destroy()
						broker2.Destroy()
						helpers.QuickDeleteOrg(org)
					})

					It("prints a table of all publically available services", func() {
						session := helpers.CF("marketplace")
						Eventually(session).Should(Say("Getting all services from marketplace"))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say("service\\s+plans\\s+description"))
						Eventually(session).Should(Say("%s\\s+%s\\s+fake service", getServiceName(broker1), getBrokerPlanNames(broker1)))
						Eventually(session).Should(Say("%s\\s+%s\\s+fake service", getServiceName(broker2), getBrokerPlanNames(broker2)))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when logged in", func() {
				var user string

				BeforeEach(func() {
					helpers.LoginCF()
					user, _ = helpers.GetCredentials()
				})

				Context("and no space is targeted", func() {
					BeforeEach(func() {
						helpers.TargetOrg(ReadOnlyOrg)
					})

					It("prints an error that a space must be targeted", func() {
						session := helpers.CF("marketplace")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Cannot list marketplace services without a targeted space"))
						Eventually(session).Should(Exit(1))
					})
				})

				Context("when a space is targeted", func() {
					BeforeEach(func() {
						helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)
					})

					Context("and there are no services", func() {
						It("prints a message saying no services exist", func() {
							session := helpers.CF("marketplace")
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("No service offerings found"))
							Eventually(session).Should(Exit(0))
						})
					})

					Context("when a service is scoped to another space", func() {
						var (
							broker1 helpers.ServiceBroker
							org1    string
							space1  string

							broker2 helpers.ServiceBroker
							org2    string
							space2  string
						)

						BeforeEach(func() {
							org1 = helpers.NewOrgName()
							space1 = helpers.NewSpaceName()
							helpers.SetupCF(org1, space1)
							helpers.TargetOrgAndSpace(org1, space1)

							domain := helpers.DefaultSharedDomain()

							broker1 = createBroker(domain, "SERVICE-1", "SERVICE-PLAN-1")
							enableServiceAccessForOrg(broker1, org1)

							org2 = helpers.NewOrgName()
							space2 = helpers.NewSpaceName()
							helpers.CreateOrgAndSpace(org2, space2)
							helpers.TargetOrgAndSpace(org2, space2)

							broker2 = createBroker(domain, "SERVICE-2", "SERVICE-PLAN-2")
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

						It("does not display services in other spaces", func() {
							session := helpers.CF("marketplace")
							Eventually(session).Should(Say("Getting services from marketplace in org %s / space %s as %s\\.\\.\\.", org2, space2, user))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("service\\s+plans\\s+description"))
							Consistently(session).ShouldNot(Say("%s\\s+%s\\s+fake service", getServiceName(broker1), getBrokerPlanNames(broker1)))
							Eventually(session).Should(Say("%s\\s+%s\\s+fake service", getServiceName(broker2), getBrokerPlanNames(broker2)))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})
	})

	Context("when the -s flag is passed", func() {
		Context("the api is not set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("prints an error message that the API is not set", func() {
				session := helpers.CF("marketplace")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the api is set", func() {
			Context("when not logged in", func() {
				BeforeEach(func() {
					helpers.LogoutCF()
				})

				Context("the specified service does not exist", func() {
					It("prints an error that the service doesn't exist", func() {
						session := helpers.CF("marketplace", "-s", "not-found-service")
						Eventually(session).Should(Say("Getting service plan information for service not-found-service\\.\\.\\."))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Service offering 'not-found-service' not found"))
						Eventually(session).Should(Exit(1))
					})
				})

				Context("when the specified service exists", func() {
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

						broker = createBroker(domain, "SERVICE", "SERVICE-PLAN")
						enableServiceAccess(broker)

						helpers.LogoutCF()
					})

					AfterEach(func() {
						helpers.LoginCF()
						helpers.TargetOrgAndSpace(org, space)
						broker.Destroy()
						helpers.QuickDeleteOrg(org)
					})

					It("prints the service", func() {
						description := "Shared fake Server, 5tb persistent disk, 40 max concurrent connections"
						session := helpers.CF("marketplace", "-s", getServiceName(broker))
						Eventually(session).Should(Say("Getting service plan information for service %s\\.\\.\\.", getServiceName(broker)))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say("service plan\\s+description\\s+free or paid"))
						Eventually(session).Should(Say("%s\\s+%s\\s+%s", getPlanName(broker), description, "free"))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when logged in", func() {
				var user string

				BeforeEach(func() {
					helpers.LoginCF()
					user, _ = helpers.GetCredentials()
				})

				Context("when no space is targeted", func() {
					BeforeEach(func() {
						helpers.TargetOrg(ReadOnlyOrg)
					})

					It("prints an error that a space must be targeted", func() {
						session := helpers.CF("marketplace", "-s", "service")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Cannot list marketplace services without a targeted space"))
						Eventually(session).Should(Exit(1))
					})
				})

				Context("when a space is targeted", func() {
					BeforeEach(func() {
						helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)
					})

					Context("the specified service does not exist", func() {
						It("prints an error that the service doesn't exist", func() {
							session := helpers.CF("marketplace", "-s", "not-found-service")
							Eventually(session).Should(Say("Getting service plan information for service not-found-service as %s\\.\\.\\.", user))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Service offering 'not-found-service' not found"))
							Eventually(session).Should(Exit(1))
						})
					})

					Context("when the specified service exists", func() {
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

							broker = createBroker(domain, "SERVICE", "SERVICE-PLAN")
							enableServiceAccess(broker)
						})

						AfterEach(func() {
							broker.Destroy()
							helpers.QuickDeleteOrg(org)
						})

						It("prints the service", func() {
							description := "Shared fake Server, 5tb persistent disk, 40 max concurrent connections"
							session := helpers.CF("marketplace", "-s", getServiceName(broker))
							Eventually(session).Should(Say("Getting service plan information for service %s as %s\\.\\.\\.", getServiceName(broker), user))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("service plan\\s+description\\s+free or paid"))
							Eventually(session).Should(Say("%s\\s+%s\\s+%s", getPlanName(broker), description, "free"))
							Eventually(session).Should(Exit(0))
						})
					})

					Context("when trying to list a service not available in our space", func() {
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

							broker = createBroker(domain, "SERVICE", "SERVICE-PLAN")
							enableServiceAccessForOrg(broker, org)

							helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)
						})

						AfterEach(func() {
							helpers.TargetOrgAndSpace(org, space)
							broker.Destroy()
							helpers.QuickDeleteOrg(org)
						})

						It("prints an error that the service doesn't exist", func() {
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

func createBroker(domain, serviceName, planName string) helpers.ServiceBroker {
	service := helpers.PrefixedRandomName(serviceName)
	servicePlan := helpers.PrefixedRandomName(planName)
	broker := helpers.NewServiceBroker(helpers.NewServiceBrokerName(), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
	broker.Push()
	broker.Configure(true)
	broker.Create()

	return broker
}

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
