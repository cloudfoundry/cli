package requirements_test

import (
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/requirements"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("APIEndpointRequirement", func() {
	var (
		config coreconfig.Repository
	)

	BeforeEach(func() {
		config = testconfig.NewRepository()
	})

	It("succeeds when given a config with an API endpoint", func() {
		config.SetAPIEndpoint("api.example.com")
		req := NewAPIEndpointRequirement(config)
		err := req.Execute()
		Expect(err).NotTo(HaveOccurred())
	})

	It("fails when given a config without an API endpoint", func() {
		req := NewAPIEndpointRequirement(config)
		err := req.Execute()
		Expect(err).To(HaveOccurred())

		Expect(err.Error()).To(ContainSubstring("No API endpoint set"))
	})
})
