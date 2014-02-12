package requirements_test

import (
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testassert "testhelpers/assert"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestLoginRequirement", func() {

		ui := new(testterm.FakeUI)
		config := testconfig.NewRepositoryWithDefaults()

		req := NewLoginRequirement(ui, config)
		success := req.Execute()
		Expect(success).To(BeTrue())

		config.SetAccessToken("")
		req = NewLoginRequirement(ui, config)
		success = req.Execute()
		Expect(success).To(BeFalse())

		testassert.SliceContains(ui.Outputs, testassert.Lines{{"Not logged in."}})
	})
})
