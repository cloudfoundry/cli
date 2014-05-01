package requirements_test

import (
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("LoginRequirement", func() {
	var ui *testterm.FakeUI

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
	})

	It("succeeds when given a config with an API endpoint and authentication", func() {
		config := testconfig.NewRepositoryWithAccessToken(configuration.TokenInfo{Username: "my-user"})
		config.SetApiEndpoint("api.example.com")
		req := NewLoginRequirement(ui, config)
		success := req.Execute()
		Expect(success).To(BeTrue())
	})

	It("fails when given a config with only an API endpoint", func() {
		config := testconfig.NewRepository()
		config.SetApiEndpoint("api.example.com")
		req := NewLoginRequirement(ui, config)
		success := req.Execute()
		Expect(success).To(BeFalse())

		Expect(ui.Outputs).To(ContainSubstrings([]string{"Not logged in."}))
	})

	It("fails when given a config with neither an API endpoint nor authentication", func() {
		config := testconfig.NewRepository()
		req := NewLoginRequirement(ui, config)
		success := req.Execute()
		Expect(success).To(BeFalse())

		Expect(ui.Outputs).To(ContainSubstrings([]string{"No API endpoint"}))
		Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"Not logged in."}))
	})
})
