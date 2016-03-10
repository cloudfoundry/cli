package requirements_test

import (
	"errors"

	test_org "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OrganizationRequirement", func() {
	Context("when an org with the given name exists", func() {
		It("succeeds", func() {
			org := models.Organization{}
			org.Name = "my-org-name"
			org.Guid = "my-org-guid"
			orgRepo := &test_org.FakeOrganizationRepository{}
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
		orgRepo := &test_org.FakeOrganizationRepository{}

		orgError := errors.New("not found")
		orgRepo.FindByNameReturns(models.Organization{}, orgError)

		err := NewOrganizationRequirement("foo", orgRepo).Execute()
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(orgError))
	})
})
