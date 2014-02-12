package requirements_test

import (
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
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
		assert.True(mr.T(), success)

		config.SetAccessToken("")
		req = NewLoginRequirement(ui, config)
		success = req.Execute()
		assert.False(mr.T(), success)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{{"Not logged in."}})
	})
})
