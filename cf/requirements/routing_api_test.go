package requirements_test

import (
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/requirements"

	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RoutingApi", func() {
	var (
		config      coreconfig.Repository
		requirement requirements.RoutingAPIRequirement
	)

	BeforeEach(func() {
		config = testconfig.NewRepositoryWithAccessToken(coreconfig.TokenInfo{Username: "my-user"})
		requirement = requirements.NewRoutingAPIRequirement(config)
	})

	Context("when the config has a zero-length RoutingApiEndpoint", func() {
		BeforeEach(func() {
			config.SetRoutingApiEndpoint("")
		})

		It("errors", func() {
			err := requirement.Execute()
			Expect(err.Error()).To(ContainSubstring("This command requires the Routing API. Your targeted endpoint reports it is not enabled."))
		})
	})

	Context("when the config has a RoutingApiEndpoint", func() {
		BeforeEach(func() {
			config.SetRoutingApiEndpoint("api.example.com")
		})

		It("does not error", func() {
			err := requirement.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
