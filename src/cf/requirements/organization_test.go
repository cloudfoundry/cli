package requirements_test

import (
	"cf/models"
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
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

		assert.True(mr.T(), success)
		Expect(orgRepo.FindByNameName).To(Equal("my-org-name"))
		Expect(orgReq.GetOrganization()).To(Equal(org))
	})

	It("TestOrgReqWhenOrgDoesNotExist", func() {

		orgRepo := &testapi.FakeOrgRepository{FindByNameNotFound: true}
		ui := new(testterm.FakeUI)

		orgReq := NewOrganizationRequirement("foo", ui, orgRepo)

		testassert.AssertPanic(mr.T(), testterm.FailedWasCalled, func() {
			orgReq.Execute()
		})
	})
})
