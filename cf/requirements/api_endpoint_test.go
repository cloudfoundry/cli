package requirements_test

import (
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	. "code.cloudfoundry.org/cli/cf/requirements"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
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
