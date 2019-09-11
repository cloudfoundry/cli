package isolated_test

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/fakeservicebroker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("FakeServiceBroker", func() {
	var broker *fakeservicebroker.FakeServiceBroker

	BeforeEach(func() {
		helpers.SetupCFWithGeneratedOrgAndSpaceNames()
	})

	AfterEach(func() {
		if broker != nil {
			broker.Destroy()
		}
	})

	It("can create and reuse a service broker and use it, and dispose of it", func() {
		broker = fakeservicebroker.New().Async().Register()
		service := broker.ServiceName()
		servicePlan := broker.ServicePlanName()
		serviceInstance := helpers.PrefixedRandomName("si")

		Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
		Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))

		broker.Destroy()
		broker = nil
	})

	It("can reuse and reconfigure the broker that has service instances", func() {
		broker = fakeservicebroker.New().Async().Register()

		service := broker.ServiceName()
		servicePlan := broker.ServicePlanName()
		serviceInstance := helpers.PrefixedRandomName("si")

		Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
		Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))

		broker = fakeservicebroker.New().Async().Register()
	})
})
