package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/fakeservicebroker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("service-access command", func() {
	var (
		userName string
	)

	BeforeEach(func() {
		userName, _ = helpers.GetCredentials()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("service-access", "--help")
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Say(`\s+service-access - List service access settings`))
				Eventually(session).Should(Say(`USAGE:`))
				Eventually(session).Should(Say(`\s+cf service-access \[-b BROKER\] \[-e SERVICE\] \[-o ORG\]`))
				Eventually(session).Should(Say(`OPTIONS:`))
				Eventually(session).Should(Say(`\s+-b\s+Access for plans of a particular broker`))
				Eventually(session).Should(Say(`\s+-e\s+Access for service name of a particular service offering`))
				Eventually(session).Should(Say(`\s+-o\s+Plans accessible by a particular organization`))
				Eventually(session).Should(Say(`SEE ALSO:`))
				Eventually(session).Should(Say(`\s+disable-service-access, enable-service-access, marketplace, service-brokers`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "service-access")
		})
	})

	When("the environment is setup correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
			helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)
		})

		When("-b is provided with a broker name that does not exist", func() {
			It("shows an error message", func() {
				session := helpers.CF("service-access", "-b", "non-existent-broker")
				Eventually(session).Should(Say(`Getting service access for broker non-existent-broker as %s\.\.\.`, userName))
				Eventually(session.Err).Should(Say(`Service broker 'non-existent-broker' not found\.`))
				Eventually(session.Err).Should(Say(`TIP: Use 'cf service-brokers' to see a list of available brokers\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("-e is provided with a service name that does not exist", func() {
			It("shows an error message", func() {
				session := helpers.CF("service-access", "-e", "non-existent-service")
				Eventually(session).Should(Say(`Getting service access for service non-existent-service as %s\.\.\.`, userName))
				Eventually(session.Err).Should(Say(`Service offering 'non-existent-service' not found\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("-o is provided with a org name that does not exist", func() {
			It("shows an error message", func() {
				session := helpers.CF("service-access", "-o", "non-existent-org")
				Eventually(session).Should(Say(`Getting service access for organization non-existent-org as %s\.\.\.`, userName))
				Eventually(session.Err).Should(Say(`Organization 'non-existent-org' not found`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("there are service offerings", func() {
			var (
				orgName   string
				spaceName string

				service     string
				servicePlan string
				broker      *fakeservicebroker.FakeServiceBroker
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)

				broker = fakeservicebroker.New()
				broker.Services[0].Plans[1].Name = helpers.GenerateHigherName(helpers.NewPlanName, broker.Services[0].Plans[0].Name)
				broker.Register()
				service = broker.ServiceName()
				servicePlan = broker.ServicePlanName()
			})

			AfterEach(func() {
				broker.Destroy()
				helpers.QuickDeleteOrg(orgName)
			})

			It("displays all service access information", func() {
				By("showing 'none' when service access is disabled")
				session := helpers.CF("service-access")
				Eventually(session).Should(Say("Getting service access as %s...", userName))
				Eventually(session).Should(Say(`service\s+plan\s+access\s+org`))
				Eventually(session).Should(Say(`%s\s+%s\s+%s`, service, servicePlan, "none"))
				Eventually(session).Should(Exit(0))

				By("showing 'all' when service access is enabled globally")
				Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
				session = helpers.CF("service-access")
				Eventually(session).Should(Say("Getting service access as %s...", userName))
				Eventually(session).Should(Say(`service\s+plan\s+access\s+org`))
				Eventually(session).Should(Say(`%s\s+%s\s+%s`, service, servicePlan, "all"))
				Eventually(session).Should(Exit(0))
			})

			When("some services are only accessible to certain organizations", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("enable-service-access", service, "-o", orgName)).Should(Exit(0))
				})

				It("shows 'limited' access to the service", func() {
					session := helpers.CF("service-access")
					Eventually(session).Should(Say("Getting service access as %s...", userName))
					Eventually(session).Should(Say(`service\s+plan\s+access\s+org`))
					Eventually(session).Should(Say(`%s\s+%s\s+%s\s+%s`, service, servicePlan, "limited", orgName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("multiple brokers are registered and with varying service accessibility", func() {
				var (
					otherBroker *fakeservicebroker.FakeServiceBroker

					otherOrgName string
				)

				BeforeEach(func() {
					helpers.SetupCF(orgName, spaceName)

					otherBroker = fakeservicebroker.New().WithNameSuffix("other")
					otherBroker.Services[0].Plans[1].Name = helpers.GenerateLowerName(helpers.NewPlanName, otherBroker.Services[0].Plans[0].Name)
					otherBroker.Register()

					otherOrgName = helpers.GenerateLowerName(helpers.NewOrgName, orgName)
					helpers.CreateOrg(otherOrgName)

					Eventually(
						helpers.CF("enable-service-access",
							service,
							"-o", otherOrgName,
							"-p", servicePlan)).Should(Exit(0))
					Eventually(helpers.CF("enable-service-access", otherBroker.Services[0].Name)).Should(Exit(0))
				})

				AfterEach(func() {
					helpers.QuickDeleteOrg(otherOrgName)
					otherBroker.Destroy()
				})

				When("the -b flag is passed", func() {
					It("shows only services from the specified broker", func() {
						session := helpers.CF("service-access", "-b", otherBroker.Name())
						Eventually(session).Should(Say("Getting service access for broker %s as %s...", otherBroker.Name(), userName))
						Eventually(session).Should(Say(`broker:\s+%s`, otherBroker.Name()))
						Eventually(session).Should(Say(`service\s+plan\s+access\s+org`))
						Eventually(session).Should(Say(`%s\s+%s\s+%s`, otherBroker.Services[0].Name, otherBroker.Services[0].Plans[0].Name, "all"))
						Eventually(string(session.Out.Contents())).ShouldNot(ContainSubstring(service))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the -e flag is passed", func() {
					It("shows only services from the specified service", func() {
						session := helpers.CF("service-access", "-e", otherBroker.Services[0].Name)
						Eventually(session).Should(Say("Getting service access for service %s as %s...", otherBroker.Services[0].Name, userName))
						Eventually(session).Should(Say(`broker:\s+%s`, otherBroker.Name()))
						Eventually(session).Should(Say(`service\s+plan\s+access\s+org`))
						Eventually(session).Should(Say(`%s\s+%s\s+%s`, otherBroker.Services[0].Name, otherBroker.Services[0].Plans[0].Name, "all"))
						Eventually(string(session.Out.Contents())).ShouldNot(ContainSubstring(service))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the -o flag is passed", func() {
					It("displays only plans accessible by the specified organization", func() {
						By("not displaying brokers that were only enabled in a different org than the provided one")
						session := helpers.CF("service-access", "-o", orgName)
						Eventually(session).Should(Say(`broker:\s+%s`, otherBroker.Name()))
						Eventually(session).Should(Say(`%s\s+%s\s+all`,
							otherBroker.Services[0].Name,
							otherBroker.Services[0].Plans[1].Name,
						))
						Eventually(session).Should(Say(`%s\s+%s\s+all`,
							otherBroker.Services[0].Name,
							otherBroker.Services[0].Plans[0].Name,
						))
						Consistently(session).ShouldNot(Say(`broker:\s+%s`, broker.Name()))
						Eventually(session).Should(Exit(0))

						By("displaying brokers that were enabled in the provided org")
						session = helpers.CF("service-access", "-o", otherOrgName)
						Eventually(session).Should(Say(`broker:\s+%s`, broker.Name()))
						Eventually(session).Should(Say(`%s\s+%s\s+limited\s+%s`,
							broker.Services[0].Name,
							broker.Services[0].Plans[0].Name,
							otherOrgName,
						))
						Eventually(session).Should(Say(`broker:\s+%s`, otherBroker.Name()))
						Eventually(session).Should(Say(`%s\s+%s\s+all`,
							otherBroker.Services[0].Name,
							otherBroker.Services[0].Plans[1].Name,
						))
						Eventually(session).Should(Say(`%s\s+%s\s+all`,
							otherBroker.Services[0].Name,
							otherBroker.Services[0].Plans[0].Name,
						))

						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
