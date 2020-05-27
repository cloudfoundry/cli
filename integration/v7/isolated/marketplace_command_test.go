package isolated

import (
	"strings"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"

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
				Eventually(session).Should(Exit(0))
				expectMarketplaceHelpMessage(session)
			})
		})

		When("more than required number of args are passed", func() {
			It("displays an error message with help text and exits 1", func() {
				session := helpers.CF("marketplace", "lala")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "lala"`))
				expectMarketplaceHelpMessage(session)
			})
		})
	})

	When("an API target is not set", func() {
		BeforeEach(func() {
			helpers.UnsetAPI()
		})

		It("displays an error message that no API endpoint is set", func() {
			session := helpers.CF("marketplace")
			Eventually(session).Should(Exit(1))

			Expect(session).To(Say("FAILED"))
			Expect(session.Err).To(Say(`No API endpoint set\. Use 'cf login' or 'cf api' to target an endpoint\.`))
		})
	})

	When("services and plans are registered", func() {
		var (
			org1, org2, space1, space2 string
			broker1, broker2, broker3  *servicebrokerstub.ServiceBrokerStub
		)

		BeforeEach(func() {
			helpers.LoginCF()

			org1 = helpers.NewOrgName()
			org2 = helpers.NewOrgName()
			space1 = helpers.NewSpaceName()
			space2 = helpers.NewSpaceName()
			helpers.CreateOrgAndSpace(org1, space1)
			helpers.CreateOrgAndSpace(org2, space2)

			broker1 = servicebrokerstub.EnableServiceAccess()
			broker2 = servicebrokerstub.New().WithHigherNameThan(broker1).WithServiceOfferings(2).WithPlans(4).Register()

			Eventually(helpers.CF(
				"enable-service-access", broker2.FirstServiceOfferingName(),
				"-b", broker2.Name,
				"-p", broker2.Services[0].Plans[0].Name,
				"-o", org1,
			)).Should(Exit(0))

			Eventually(helpers.CF(
				"enable-service-access", broker2.FirstServiceOfferingName(),
				"-b", broker2.Name,
				"-p", broker2.Services[0].Plans[1].Name,
				"-o", org2,
			)).Should(Exit(0))

			Eventually(helpers.CF(
				"enable-service-access", broker2.FirstServiceOfferingName(),
				"-b", broker2.Name,
				"-p", broker2.Services[0].Plans[2].Name,
			)).Should(Exit(0))

			Eventually(helpers.CF(
				"enable-service-access", broker2.Services[1].Name,
				"-b", broker2.Name,
			)).Should(Exit(0))

			broker3 = servicebrokerstub.New().WithHigherNameThan(broker2).WithPlans(2)
			broker3.Services[0].Name = broker2.Services[0].Name
			broker3.Services[0].Plans[0].Free = false
			broker3.Services[0].Plans[0].Costs = []config.Cost{
				{
					Amount: map[string]float64{"gbp": 600.00, "usd": 649.00},
					Unit:   "MONTHLY",
				},
				{
					Amount: map[string]float64{"usd": 0.999},
					Unit:   "1GB of messages over 20GB",
				},
			}
			broker3.Services[0].Plans[1].Free = false
			broker3.EnableServiceAccess()
		})

		AfterEach(func() {
			helpers.LoginCF()
			helpers.QuickDeleteOrg(org1)
			helpers.QuickDeleteOrg(org2)
			broker1.Forget()
			broker2.Forget()
			broker3.Forget()
		})

		Context("no service name filter", func() {
			When("not logged in", func() {
				BeforeEach(func() {
					helpers.LogoutCF()
				})

				It("displays all public offerings and plans", func() {
					session := helpers.CF("marketplace")
					Eventually(session).Should(Exit(0))

					Expect(session).To(Say(`Getting all service offerings from marketplace\.\.\.`))
					Expect(session).To(Say(`\n\n`))
					Expect(session).To(Say(`offering\s+plans\s+description\s+broker`))
					Expect(session).To(Say(`%s\s+%s\s+%s\s+%s`, broker1.FirstServiceOfferingName(), planNamesOf(broker1), broker1.FirstServiceOfferingDescription(), broker1.Name))
					Expect(session).To(Say(`%s\s+%s\s+%s\s+%s`, broker2.FirstServiceOfferingName(), broker2.Services[0].Plans[2].Name, broker2.FirstServiceOfferingDescription(), broker2.Name))
					Expect(session).To(Say(`%s\s+%s\s+%s\s+%s`, broker2.Services[1].Name, broker2.Services[1].Plans[0].Name, broker2.Services[1].Description, broker2.Name))
					Expect(session).To(Say(`%s\s+%s\s+%s\s+%s`, broker3.FirstServiceOfferingName(), planNamesOf(broker3), broker3.FirstServiceOfferingDescription(), broker3.Name))
					Expect(session).To(Say(`\n\n`))
					Expect(session).To(Say(`TIP: Use 'cf marketplace -e SERVICE_OFFERING' to view descriptions of individual plans of a given service offering\.`))
				})

				It("can filter by service broker name", func() {
					session := helpers.CF("marketplace", "-b", broker2.Name)
					Eventually(session).Should(Exit(0))

					Expect(session).To(Say(`Getting all service offerings from marketplace for service broker %s\.\.\.`, broker2.Name))
					Expect(session).To(Say(`\n\n`))
					Expect(session).To(Say(`offering\s+plans\s+description\s+broker`))
					Expect(session).To(Say(`%s\s+%s\s+%s\s+%s`, broker2.FirstServiceOfferingName(), broker2.Services[0].Plans[2].Name, broker2.FirstServiceOfferingDescription(), broker2.Name))
					Expect(session).To(Say(`%s\s+%s\s+%s\s+%s`, broker2.Services[1].Name, broker2.Services[1].Plans[0].Name, broker2.Services[1].Description, broker2.Name))
					Expect(session).To(Say(`\n\n`))
					Expect(session).To(Say(`TIP: Use 'cf marketplace -e SERVICE_OFFERING' to view descriptions of individual plans of a given service offering\.`))

					Expect(string(session.Out.Contents())).NotTo(SatisfyAny(
						ContainSubstring(broker1.Name),
						ContainSubstring(broker3.Name),
					))
				})
			})

			When("logged in and targeting a space", func() {
				var username string

				BeforeEach(func() {
					helpers.TargetOrgAndSpace(org1, space1)
					username, _ = helpers.GetCredentials()
				})

				It("displays public offerings and plans, and those enabled for that space", func() {
					session := helpers.CF("marketplace")
					Eventually(session).Should(Exit(0))

					broker2Plans := strings.Join([]string{
						broker2.Services[0].Plans[0].Name,
						broker2.Services[0].Plans[2].Name},
						", ",
					)

					Expect(session).To(Say(`Getting all service offerings from marketplace in org %s / space %s as %s\.\.\.`, org1, space1, username))
					Expect(session).To(Say(`\n\n`))
					Expect(session).To(Say(`offering\s+plans\s+description\s+broker`))
					Expect(session).To(Say(`%s\s+%s\s+%s\s+%s`, broker1.FirstServiceOfferingName(), planNamesOf(broker1), broker1.FirstServiceOfferingDescription(), broker1.Name))
					Expect(session).To(Say(`%s\s+%s\s+%s\s+%s`, broker2.FirstServiceOfferingName(), broker2Plans, broker2.FirstServiceOfferingDescription(), broker2.Name))
					Expect(session).To(Say(`%s\s+%s\s+%s\s+%s`, broker2.Services[1].Name, broker2.Services[1].Plans[0].Name, broker2.Services[1].Description, broker2.Name))
					Expect(session).To(Say(`%s\s+%s\s+%s\s+%s`, broker3.FirstServiceOfferingName(), planNamesOf(broker3), broker3.FirstServiceOfferingDescription(), broker3.Name))
					Expect(session).To(Say(`\n\n`))
					Expect(session).To(Say(`TIP: Use 'cf marketplace -e SERVICE_OFFERING' to view descriptions of individual plans of a given service offering\.`))
				})

				It("can filter by service broker name", func() {
					session := helpers.CF("marketplace", "-b", broker2.Name)
					Eventually(session).Should(Exit(0))

					broker2Plans := strings.Join([]string{
						broker2.Services[0].Plans[0].Name,
						broker2.Services[0].Plans[2].Name},
						", ",
					)

					Expect(session).To(Say(`Getting all service offerings from marketplace for service broker %s in org %s / space %s as %s\.\.\.`, broker2.Name, org1, space1, username))
					Expect(session).To(Say(`\n\n`))
					Expect(session).To(Say(`offering\s+plans\s+description\s+broker`))
					Expect(session).To(Say(`%s\s+%s\s+%s\s+%s`, broker2.FirstServiceOfferingName(), broker2Plans, broker2.FirstServiceOfferingDescription(), broker2.Name))
					Expect(session).To(Say(`%s\s+%s\s+%s\s+%s`, broker2.Services[1].Name, broker2.Services[1].Plans[0].Name, broker2.Services[1].Description, broker2.Name))
					Expect(session).To(Say(`\n\n`))
					Expect(session).To(Say(`TIP: Use 'cf marketplace -e SERVICE_OFFERING' to view descriptions of individual plans of a given service offering\.`))

					Expect(string(session.Out.Contents())).NotTo(SatisfyAny(
						ContainSubstring(broker1.Name),
						ContainSubstring(broker3.Name),
					))
				})
			})
		})

		Context("filtering by service offering name", func() {
			When("not logged in", func() {
				BeforeEach(func() {
					helpers.LogoutCF()
				})

				It("filters by service offering name", func() {
					session := helpers.CF("marketplace", "-e", broker2.Services[0].Name)
					Eventually(session).Should(Exit(0))

					Expect(session).To(Say(`Getting service plan information for service offering %s\.\.\.`, broker2.Services[0].Name))
					expectMarketplaceServiceOfferingOutput(session, broker2, broker3)

					Expect(string(session.Out.Contents())).NotTo(SatisfyAny(
						ContainSubstring(broker1.Name),
						ContainSubstring(broker2.Services[1].Name),
					))
				})

				It("can also filter by service broker name", func() {
					session := helpers.CF("marketplace", "-e", broker2.Services[0].Name, "-b", broker2.Name)
					Eventually(session).Should(Exit(0))

					Expect(session).To(Say(`Getting service plan information for service offering %s from service broker %s\.\.\.`, broker2.Services[0].Name, broker2.Name))
					Expect(session).To(Say(`\n\n`))
					Expect(session).To(Say(`broker: %s`, broker2.Name))
					Expect(session).To(Say(`plan\s+description\s+free or paid\s+cost`))
					Expect(session).To(Say(`%s\s+%s\s+%s`, broker2.Services[0].Plans[2].Name, broker2.Services[0].Plans[2].Description, "free"))

					Expect(string(session.Out.Contents())).NotTo(SatisfyAny(
						ContainSubstring(broker1.Name),
						ContainSubstring(broker2.Services[1].Plans[0].Name),
						ContainSubstring(broker3.Name),
					))
				})
			})

			When("logged in and targeting a space", func() {
				var username string

				BeforeEach(func() {
					helpers.TargetOrgAndSpace(org1, space1)
					username, _ = helpers.GetCredentials()
				})

				It("filters by service offering name", func() {
					session := helpers.CF("marketplace", "-e", broker2.Services[0].Name)
					Eventually(session).Should(Exit(0))

					Expect(session).To(Say(`Getting service plan information for service offering %s in org %s / space %s as %s\.\.\.`, broker2.Services[0].Name, org1, space1, username))
					expectMarketplaceServiceOfferingOutput(session, broker2, broker3)
				})

				It("can also filter by service broker name", func() {
					session := helpers.CF("marketplace", "-e", broker2.Services[0].Name, "-b", broker2.Name)
					Eventually(session).Should(Exit(0))

					Expect(session).To(Say(`Getting service plan information for service offering %s from service broker %s in org %s / space %s as %s\.\.\.`, broker2.Services[0].Name, broker2.Name, org1, space1, username))
					Expect(session).To(Say(`\n\n`))
					Expect(session).To(Say(`broker: %s`, broker2.Name))
					Expect(session).To(Say(`plan\s+description\s+free or paid\s+cost`))
					Expect(session).To(Say(`%s\s+%s\s+%s`, broker2.Services[0].Plans[2].Name, broker2.Services[0].Plans[2].Description, "free"))

					Expect(string(session.Out.Contents())).NotTo(SatisfyAny(
						ContainSubstring(broker1.Name),
						ContainSubstring(broker3.Name),
					))
				})
			})
		})
	})
})

func expectMarketplaceServiceOfferingOutput(session *Session, broker2, broker3 *servicebrokerstub.ServiceBrokerStub) {
	Expect(session).To(Say(`\n\n`))
	Expect(session).To(Say(`broker: %s`, broker2.Name))
	Expect(session).To(Say(`plan\s+description\s+free or paid\s+cost`))
	Expect(session).To(Say(`%s\s+%s\s+%s`, broker2.Services[0].Plans[2].Name, broker2.Services[0].Plans[2].Description, "free"))

	Expect(session).To(Say(`\n\n`))
	Expect(session).To(Say(`broker: %s`, broker3.Name))
	Expect(session).To(Say(`plan\s+description\s+free or paid\s+cost`))
	Expect(session).To(Say(`%s\s+%s\s+%s\s+%s`, broker3.Services[0].Plans[0].Name, broker3.Services[0].Plans[0].Description, "paid", "GBP 600.00/MONTHLY, USD 649.00/MONTHLY, USD 1.00/1GB of messages over 20GB"))
	Expect(session).To(Say(`%s\s+%s\s+%s`, broker3.Services[0].Plans[1].Name, broker3.Services[0].Plans[1].Description, "paid"))
}

func planNamesOf(broker *servicebrokerstub.ServiceBrokerStub) string {
	var planNames []string
	for _, p := range broker.Services[0].Plans {
		planNames = append(planNames, p.Name)
	}
	return strings.Join(planNames, ", ")
}

func expectMarketplaceHelpMessage(session *Session) {
	Expect(session).To(Say(`NAME:`))
	Expect(session).To(Say(`marketplace - List available offerings in the marketplace`))
	Expect(session).To(Say(`USAGE:`))
	Expect(session).To(Say(`cf marketplace \[-e SERVICE_OFFERING\] \[-b SERVICE_BROKER\] \[--no-plans\]`))
	Expect(session).To(Say(`ALIAS:`))
	Expect(session).To(Say(`m`))
	Expect(session).To(Say(`OPTIONS:`))
	Expect(session).To(Say(`-e\s+Show plan details for a particular service offering`))
	Expect(session).To(Say(`--no-plans\s+Hide plan information for service offerings`))
	Expect(session).To(Say(`create-service, services`))
}
