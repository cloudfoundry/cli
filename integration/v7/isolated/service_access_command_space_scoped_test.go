package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
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
		broker      *servicebrokerstub.ServiceBrokerStub
	)

	BeforeEach(func() {
		userName, _ = helpers.GetCredentials()

		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		helpers.SetupCF(orgName, spaceName)

		broker = servicebrokerstub.Create().RegisterSpaceScoped()
		service = broker.FirstServiceOfferingName()
		servicePlan = broker.FirstServicePlanName()
	})

	AfterEach(func() {
		broker.Forget()
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
