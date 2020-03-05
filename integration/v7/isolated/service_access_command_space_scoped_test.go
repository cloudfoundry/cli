package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/fakeservicebroker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("service-access command for space-scoped service offerings", func() {
	var (
		userName string

		orgName   string
		spaceName string

		service     string
		servicePlan string
		broker      *fakeservicebroker.FakeServiceBroker
	)

	BeforeEach(func() {
		userName, _ = helpers.GetCredentials()

		helpers.LoginCF()
		helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)

		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		helpers.SetupCF(orgName, spaceName)

		broker = fakeservicebroker.New().WithSpaceScoped()
		broker.Services[0].Plans[1].Name = helpers.GenerateHigherName(helpers.NewPlanName, broker.Services[0].Plans[0].Name)
		broker.EnsureBrokerIsAvailable()
		service = broker.ServiceName()
		servicePlan = broker.ServicePlanName()
	})

	AfterEach(func() {
		broker.Destroy()
		helpers.QuickDeleteOrg(orgName)
	})

	It("displays service access information with space and org", func() {
		session := helpers.CF("service-access")
		Eventually(session).Should(Exit(0))
		Expect(session).To(Say("Getting service access as %s...", userName))
		Expect(session).To(Say(`service\s+plan\s+access\s+space`))
		Expect(session).To(Say(`%s\s+%s\s+%s\s+%s \(org: %s\)`, service, servicePlan, "limited", spaceName, orgName))
	})
})
