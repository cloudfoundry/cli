package requirements_test

import (
	"errors"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SpaceRequirement", func() {
	Context("when a space with the given name exists", func() {
		It("succeeds", func() {
			space := models.Space{}
			space.Name = "awesome-sauce-space"
			space.Guid = "my-space-guid"
			spaceRepo := &testapi.FakeSpaceRepository{}
			spaceRepo.FindByNameReturns(space, nil)

			spaceReq := NewSpaceRequirement("awesome-sauce-space", spaceRepo)

			err := spaceReq.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(spaceReq.GetSpace()).To(Equal(space))
			Expect(spaceRepo.FindByNameArgsForCall(0)).To(Equal("awesome-sauce-space"))
		})
	})

	Context("when a space with the given name does not exist", func() {
		It("errors", func() {
			spaceRepo := &testapi.FakeSpaceRepository{}
			spaceError := errors.New("space-repo-err")
			spaceRepo.FindByNameReturns(models.Space{}, spaceError)

			err := NewSpaceRequirement("foo", spaceRepo).Execute()

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(spaceError))
		})
	})
})
