package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("service-key command", func() {
	var (
		org             string
		space           string
		service         string
		servicePlan     string
		serviceInstance string
		broker          *servicebrokerstub.ServiceBrokerStub
	)

	BeforeEach(func() {
		org = helpers.NewOrgName()
		space = helpers.NewSpaceName()
		serviceInstance = helpers.PrefixedRandomName("si")

		helpers.SetupCF(org, space)
	})

	AfterEach(func() {
		helpers.QuickDeleteOrg(org)
	})

	When("the service key is not found", func() {
		BeforeEach(func() {
			broker = servicebrokerstub.EnableServiceAccess()
			service = broker.FirstServiceOfferingName()
			servicePlan = broker.FirstServicePlanName()

			Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
			Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
		})

		AfterEach(func() {
			broker.Forget()
		})

		It("outputs an error message and exits 1", func() {
			session := helpers.CF("service-key", serviceInstance, "some-service-key")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Say(fmt.Sprintf("No service key some-service-key found for service instance %s", serviceInstance)))
			Eventually(session).Should(Exit(1))
		})

		When("the --guid option is given", func() {
			It("outputs nothing and exits 0", func() {
				session := helpers.CF("service-key", serviceInstance, "some-service-key", "--guid")
				Eventually(session).Should(Exit(0))
				Expect(session.Out.Contents()).To(Equal([]byte("\n")))
				Expect(session.Err.Contents()).To(BeEmpty())
			})
		})
	})
})
