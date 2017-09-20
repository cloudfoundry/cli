package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
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
		broker          helpers.ServiceBroker
		domain          string
	)

	BeforeEach(func() {
		org = helpers.NewOrgName()
		space = helpers.NewSpaceName()
		service = helpers.PrefixedRandomName("SERVICE")
		servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")
		serviceInstance = helpers.PrefixedRandomName("si")

		setupCF(org, space)
		domain = defaultSharedDomain()
	})

	AfterEach(func() {
		helpers.QuickDeleteOrg(org)
	})

	Context("when the service key is not found", func() {
		BeforeEach(func() {
			broker = helpers.NewServiceBroker(helpers.NewServiceBrokerName(), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
			broker.Push()
			broker.Configure()
			broker.Create()

			Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
			Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
		})

		AfterEach(func() {
			broker.Destroy()
		})

		It("outputs an error message and exits 1", func() {
			session := helpers.CF("service-key", serviceInstance, "some-service-key")
			Eventually(session.Out).Should(Say("FAILED"))
			Eventually(session.Out).Should(Say(fmt.Sprintf("No service key some-service-key found for service instance %s", serviceInstance)))
			Eventually(session).Should(Exit(1))
		})

		Context("when the --guid option is given", func() {
			It("outputs nothing and exits 0", func() {
				session := helpers.CF("service-key", serviceInstance, "some-service-key", "--guid")
				Eventually(session).Should(Exit(0))
				Expect(session.Out.Contents()).To(Equal([]byte("\n")))
				Expect(session.Err.Contents()).To(BeEmpty())
			})
		})
	})
})
