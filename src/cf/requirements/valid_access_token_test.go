package requirements

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestValidAccessRequirement", func() {

			ui := new(testterm.FakeUI)
			appRepo := &testapi.FakeApplicationRepository{
				ReadAuthErr: true,
			}

			req := newValidAccessTokenRequirement(ui, appRepo)
			success := req.Execute()
			assert.False(mr.T(), success)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{{"Not logged in."}})

			appRepo.ReadAuthErr = false

			req = newValidAccessTokenRequirement(ui, appRepo)
			success = req.Execute()
			assert.True(mr.T(), success)
		})
	})
}
