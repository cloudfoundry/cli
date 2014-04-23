package requirements_test

import (
	"cf/configuration"
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"

	. "testhelpers/matchers"
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
