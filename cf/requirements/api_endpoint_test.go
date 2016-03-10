package requirements_test

import (
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/requirements"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ApiEndpointRequirement", func() {
	var (
		config core_config.Repository
	)

	BeforeEach(func() {
		config = testconfig.NewRepository()
	})

	It("succeeds when given a config with an API endpoint", func() {
		config.SetApiEndpoint("api.example.com")
		req := NewApiEndpointRequirement(config)
		err := req.Execute()
		Expect(err).NotTo(HaveOccurred())
	})

	It("fails when given a config without an API endpoint", func() {
		req := NewApiEndpointRequirement(config)
		err := req.Execute()
		Expect(err).To(HaveOccurred())

		Expect(err.Error()).To(ContainSubstring("No API endpoint set"))
	})
})
