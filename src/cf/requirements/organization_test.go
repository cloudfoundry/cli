package requirements_test

import (
	"cf/models"
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestOrgReqExecute", func() {

		org := models.Organization{}
		org.Name = "my-org-name"
		org.Guid = "my-org-guid"
		orgRepo := &testapi.FakeOrgRepository{Organizations: []models.Organization{org}}
		ui := new(testterm.FakeUI)

		orgReq := NewOrganizationRequirement("my-org-name", ui, orgRepo)
		success := orgReq.Execute()

		Expect(success).To(BeTrue())
		Expect(orgRepo.FindByNameName).To(Equal("my-org-name"))
		Expect(orgReq.GetOrganization()).To(Equal(org))
	})

	It("TestOrgReqWhenOrgDoesNotExist", func() {

		orgRepo := &testapi.FakeOrgRepository{FindByNameNotFound: true}
		ui := new(testterm.FakeUI)

		orgReq := NewOrganizationRequirement("foo", ui, orgRepo)

		testassert.AssertPanic(testterm.FailedWasCalled, func() {
			orgReq.Execute()
		})
	})
})
