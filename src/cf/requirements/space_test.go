package requirements_test

import (
	"cf/models"
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestSpaceReqExecute", func() {

		space := models.Space{}
		space.Name = "awesome-sauce-space"
		space.Guid = "my-space-guid"
		spaceRepo := &testapi.FakeSpaceRepository{Spaces: []models.Space{space}}
		ui := new(testterm.FakeUI)

		spaceReq := NewSpaceRequirement("awesome-sauce-space", ui, spaceRepo)
		success := spaceReq.Execute()

		Expect(success).To(BeTrue())
		Expect(spaceRepo.FindByNameName).To(Equal("awesome-sauce-space"))
		Expect(spaceReq.GetSpace()).To(Equal(space))
	})

	It("TestSpaceReqExecuteWhenSpaceNotFound", func() {

		spaceRepo := &testapi.FakeSpaceRepository{FindByNameNotFound: true}
		ui := new(testterm.FakeUI)

		testassert.AssertPanic(mr.T(), testterm.FailedWasCalled, func() {
			NewSpaceRequirement("foo", ui, spaceRepo).Execute()
		})
	})
})
