package requirements_test

import (
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestValidAccessRequirement", func() {
		ui := new(testterm.FakeUI)
		appRepo := &testapi.FakeApplicationRepository{
			ReadAuthErr: true,
		}

		req := NewValidAccessTokenRequirement(ui, appRepo)
		success := req.Execute()
		Expect(success).To(BeFalse())
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{{"Not logged in."}})

		appRepo.ReadAuthErr = false

		req = NewValidAccessTokenRequirement(ui, appRepo)
		success = req.Execute()
		Expect(success).To(BeTrue())
	})
})
