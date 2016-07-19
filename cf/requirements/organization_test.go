package requirements_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/organizations/organizationsfakes"
	"code.cloudfoundry.org/cli/cf/models"
	. "code.cloudfoundry.org/cli/cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OrganizationRequirement", func() {
	var orgRepo *organizationsfakes.FakeOrganizationRepository
	BeforeEach(func() {
		orgRepo = new(organizationsfakes.FakeOrganizationRepository)
	})

	Context("when an org with the given name exists", func() {
		It("succeeds", func() {
			org := models.Organization{}
			org.Name = "my-org-name"
			org.GUID = "my-org-guid"
			orgReq := NewOrganizationRequirement("my-org-name", orgRepo)

			orgRepo.ListOrgsReturns([]models.Organization{org}, nil)
			orgRepo.FindByNameReturns(org, nil)

			err := orgReq.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-org-name"))
			Expect(orgReq.GetOrganization()).To(Equal(org))
		})
	})

	It("fails when the org with the given name does not exist", func() {
		orgError := errors.New("not found")
		orgRepo.FindByNameReturns(models.Organization{}, orgError)

		err := NewOrganizationRequirement("foo", orgRepo).Execute()
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(orgError))
	})
})
