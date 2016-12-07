package isolated

import (
	"fmt"

	. "code.cloudfoundry.org/cli/integration/helpers"
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
		broker          ServiceBroker
		domain          string
	)

	BeforeEach(func() {
		org = NewOrgName()
		space = PrefixedRandomName("SPACE")
		service = PrefixedRandomName("SERVICE")
		servicePlan = PrefixedRandomName("SERVICE-PLAN")
		serviceInstance = PrefixedRandomName("si")

		setupCF(org, space)
		domain = defaultSharedDomain()
	})

	Context("when the service key is not found", func() {
		BeforeEach(func() {
			broker = NewServiceBroker(PrefixedRandomName("SERVICE-BROKER"), NewAssets().ServiceBroker, domain, service, servicePlan)
			broker.Push()
			broker.Configure()
			broker.Create()

			Eventually(CF("enable-service-access", service)).Should(Exit(0))
			Eventually(CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
		})

		AfterEach(func() {
			broker.Destroy()
		})

		It("outputs an error message and exits 1", func() {
			session := CF("service-key", serviceInstance, "some-service-key")
			Eventually(session).Should(Exit(1))
			Expect(session.Out).To(Say("FAILED"))
			Expect(session.Out).To(Say(fmt.Sprintf("No service key some-service-key found for service instance %s", serviceInstance)))
		})

		Context("when the --guid option is given", func() {
			It("outputs nothing and exits 0", func() {
				session := CF("service-key", serviceInstance, "some-service-key", "--guid")
				Eventually(session).Should(Exit(0))
				Expect(session.Out.Contents()).To(Equal([]byte("\n")))
				Expect(session.Err.Contents()).To(BeEmpty())
			})
		})
	})
})
