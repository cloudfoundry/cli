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
			orgRepo := &testapi.FakeOrganizationRepository{Organizations: []models.Organization{org}}
			orgReq := NewOrganizationRequirement("my-org-name", ui, orgRepo)

			Expect(orgReq.Execute()).To(BeTrue())
			Expect(orgRepo.FindByNameName).To(Equal("my-org-name"))
			Expect(orgReq.GetOrganization()).To(Equal(org))
		})
	})

	It("fails when the org with the given name does not exist", func() {
		orgRepo := &testapi.FakeOrganizationRepository{FindByNameNotFound: true}

		testassert.AssertPanic(testterm.QuietPanic, func() {
			NewOrganizationRequirement("foo", ui, orgRepo).Execute()
		})
	})
})
