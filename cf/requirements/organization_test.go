package requirements_test

import (
	"errors"

	test_org "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/cf/requirements"
	testassert "github.com/cloudfoundry/cli/testhelpers/assert"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OrganizationRequirement", func() {
	var (
		ui *testterm.FakeUI
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
	})

	Context("when an org with the given name exists", func() {
		It("succeeds", func() {
			org := models.Organization{}
			org.Name = "my-org-name"
			org.Guid = "my-org-guid"
			orgRepo := &test_org.FakeOrganizationRepository{}
			orgReq := NewOrganizationRequirement("my-org-name", ui, orgRepo)

			orgRepo.ListOrgsReturns([]models.Organization{org}, nil)
			orgRepo.FindByNameReturns(org, nil)

			Expect(orgReq.Execute()).To(BeTrue())
			Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-org-name"))
			Expect(orgReq.GetOrganization()).To(Equal(org))
		})
	})

	It("fails when the org with the given name does not exist", func() {
		orgRepo := &test_org.FakeOrganizationRepository{}

		orgRepo.FindByNameReturns(models.Organization{}, errors.New("not found"))

		testassert.AssertPanic(testterm.QuietPanic, func() {
			NewOrganizationRequirement("foo", ui, orgRepo).Execute()
		})
	})
})
