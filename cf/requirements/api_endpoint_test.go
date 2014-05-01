package requirements_test

import (
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/requirements"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("ApiEndpointRequirement", func() {
	var (
		ui     *testterm.FakeUI
		config configuration.Repository
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		config = testconfig.NewRepository()
	})

	It("succeeds when given a config with an API endpoint", func() {
		config.SetApiEndpoint("api.example.com")
		req := NewApiEndpointRequirement(ui, config)
		success := req.Execute()
		Expect(success).To(BeTrue())
	})

	It("fails when given a config without an API endpoint", func() {
		req := NewApiEndpointRequirement(ui, config)
		success := req.Execute()
		Expect(success).To(BeFalse())

		Expect(ui.Outputs).To(ContainSubstrings([]string{"No API endpoint"}))
	})
})
