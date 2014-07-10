package requirements_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/cf/requirements"
	testassert "github.com/cloudfoundry/cli/testhelpers/assert"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SpaceRequirement", func() {
	var (
		ui *testterm.FakeUI
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
	})

	Context("when a space with the given name exists", func() {
		It("succeeds", func() {
			space := models.Space{}
			space.Name = "awesome-sauce-space"
			space.Guid = "my-space-guid"
			spaceRepo := &testapi.FakeSpaceRepository{Spaces: []models.Space{space}}

			spaceReq := NewSpaceRequirement("awesome-sauce-space", ui, spaceRepo)

			Expect(spaceReq.Execute()).To(BeTrue())
			Expect(spaceRepo.FindByNameName).To(Equal("awesome-sauce-space"))
			Expect(spaceReq.GetSpace()).To(Equal(space))
		})
	})

	Context("when a space with the given name does not exist", func() {
		It("fails", func() {
			spaceRepo := &testapi.FakeSpaceRepository{FindByNameNotFound: true}
			testassert.AssertPanic(testterm.QuietPanic, func() {
				NewSpaceRequirement("foo", ui, spaceRepo).Execute()
			})
		})
	})
})
