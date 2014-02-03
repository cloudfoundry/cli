package requirements

import (
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestLoginRequirement", func() {

			ui := new(testterm.FakeUI)
			config := &configuration.Configuration{
				AccessToken: "foo bar token",
			}

			req := newLoginRequirement(ui, config)
			success := req.Execute()
			assert.True(mr.T(), success)

			config = &configuration.Configuration{
				AccessToken: "",
			}

			req = newLoginRequirement(ui, config)
			success = req.Execute()
			assert.False(mr.T(), success)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{{"Not logged in."}})
		})
	})
}
