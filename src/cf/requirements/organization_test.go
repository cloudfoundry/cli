package requirements

import (
	"cf"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestOrgReqExecute", func() {

			org := cf.Organization{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"
			orgRepo := &testapi.FakeOrgRepository{FindByNameOrganization: org}
			ui := new(testterm.FakeUI)

			orgReq := newOrganizationRequirement("foo", ui, orgRepo)
			success := orgReq.Execute()

			assert.True(mr.T(), success)
			assert.Equal(mr.T(), orgRepo.FindByNameName, "foo")
			assert.Equal(mr.T(), orgReq.GetOrganization(), org)
		})
		It("TestOrgReqWhenOrgDoesNotExist", func() {

			orgRepo := &testapi.FakeOrgRepository{FindByNameNotFound: true}
			ui := new(testterm.FakeUI)

			orgReq := newOrganizationRequirement("foo", ui, orgRepo)
			success := orgReq.Execute()

			assert.False(mr.T(), success)
		})
	})
}
