package requirements_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/spaces/spacesfakes"
	"code.cloudfoundry.org/cli/cf/models"
	. "code.cloudfoundry.org/cli/cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SpaceRequirement", func() {
	var spaceRepo *spacesfakes.FakeSpaceRepository
	BeforeEach(func() {
		spaceRepo = new(spacesfakes.FakeSpaceRepository)
	})

	Context("when a space with the given name exists", func() {
		It("succeeds", func() {
			space := models.Space{}
			space.Name = "awesome-sauce-space"
			space.GUID = "my-space-guid"
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
			spaceError := errors.New("space-repo-err")
			spaceRepo.FindByNameReturns(models.Space{}, spaceError)

			err := NewSpaceRequirement("foo", spaceRepo).Execute()

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(spaceError))
		})
	})
})
