package isolated

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/fakeservicebroker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("rename-service command", func() {
	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
	})

	When("there is a service instance created", func() {
		var (
			instanceName string
			serviceName  string
			orgName      string
			spaceName    string
			broker       *fakeservicebroker.FakeServiceBroker
		)

		BeforeEach(func() {
			instanceName = helpers.PrefixedRandomName("INSTANCE")
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)

			broker = fakeservicebroker.New().Register()
			serviceName = broker.ServiceName()

			Eventually(helpers.CF("enable-service-access", serviceName)).Should(Exit(0))
			Eventually(helpers.CF("create-service", serviceName, broker.ServicePlanName(), instanceName)).Should(Exit(0))
		})

		AfterEach(func() {
			Eventually(helpers.CF("delete-service", "my-new-instance-name", "-f")).Should(Exit(0))
			broker.Destroy()
			helpers.QuickDeleteOrg(orgName)
		})

		When("and that service access is revoked for a non-admin user", func() {
			var unprivilegedUsername string

			BeforeEach(func() {
				Eventually(helpers.CF("disable-service-access", serviceName)).Should(Exit(0))

				var password string
				unprivilegedUsername, password = helpers.CreateUserInSpaceRole(orgName, spaceName, "SpaceDeveloper")
				helpers.LoginAs(unprivilegedUsername, password)
				helpers.TargetOrgAndSpace(orgName, spaceName)
			})

			AfterEach(func() {
				helpers.LoginCF()
				helpers.TargetOrgAndSpace(orgName, spaceName)
				helpers.DeleteUser(unprivilegedUsername)
			})

			When("CC API allows updating a service when plan is not visible", func() {
				BeforeEach(func() {
					helpers.SkipIfVersionLessThan(ccversion.MinVersionUpdateServiceNameWhenPlanNotVisibleV2)
				})

				It("can still rename the service", func() {
					session := helpers.CF("rename-service", instanceName, "my-new-instance-name")
					Eventually(session).Should(Exit(0))

					session = helpers.CF("services")
					Eventually(session).Should(Exit(0))
					Eventually(session).Should(Say("my-new-instance-name"))
				})
			})
		})
	})
})
