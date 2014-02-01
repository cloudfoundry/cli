package requirements

import (
	"cf"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
	"testing"
)

func TestOrgReqExecute(t *testing.T) {
	org := cf.Organization{}
	org.Name = "my-org-name"
	org.Guid = "my-org-guid"
	orgRepo := &testapi.FakeOrgRepository{Organizations: []cf.Organization{org}}
	ui := new(testterm.FakeUI)

	orgReq := newOrganizationRequirement("my-org-name", ui, orgRepo)
	success := orgReq.Execute()

	assert.True(t, success)
	assert.Equal(t, orgRepo.FindByNameName, "my-org-name")
	assert.Equal(t, orgReq.GetOrganization(), org)
}

func TestOrgReqWhenOrgDoesNotExist(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{FindByNameNotFound: true}
	ui := new(testterm.FakeUI)

	orgReq := newOrganizationRequirement("foo", ui, orgRepo)

	testassert.AssertPanic(t, testterm.FailedWasCalled, func() {
		orgReq.Execute()
	})
}
