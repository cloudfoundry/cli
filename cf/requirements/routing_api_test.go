package requirements_test

import (
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"

	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RoutingApi", func() {
	var (
		ui          *testterm.FakeUI
		config      core_config.Repository
		requirement requirements.RoutingAPIRequirement
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		config = testconfig.NewRepositoryWithAccessToken(core_config.TokenInfo{Username: "my-user"})
		requirement = requirements.NewRoutingAPIRequirement(ui, config)
	})

	Context("when the config has a zero-length RoutingApiEndpoint", func() {
		BeforeEach(func() {
			config.SetRoutingApiEndpoint("")
		})

		It("panics and prints a failure message", func() {
			Expect(func() { requirement.Execute() }).To(Panic())
			Expect(ui.Outputs).To(ContainElement("Routing API URI missing. Please log in again to set the URI automatically."))
		})
	})

	Context("when the config has a RoutingApiEndpoint", func() {
		BeforeEach(func() {
			config.SetRoutingApiEndpoint("api.example.com")
		})

		It("does not print anything", func() {
			requirement.Execute()
			Expect(ui.Outputs).To(BeEmpty())
		})

		It("returns true", func() {
			Expect(requirement.Execute()).To(BeTrue())
		})
	})
})
