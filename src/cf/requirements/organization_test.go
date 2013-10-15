package requirements

import (
	"cf"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testterm "testhelpers/terminal"
	"testing"
)

func TestOrgReqExecute(t *testing.T) {
	org := cf.Organization{Name: "my-org", Guid: "my-org-guid"}
	orgRepo := &testapi.FakeOrgRepository{FindByNameOrganization: org}
	ui := new(testterm.FakeUI)

	orgReq := newOrganizationRequirement("foo", ui, orgRepo)
	success := orgReq.Execute()

	assert.True(t, success)
	assert.Equal(t, orgRepo.FindByNameName, "foo")
	assert.Equal(t, orgReq.GetOrganization(), org)
}

func TestOrgReqWhenOrgDoesNotExist(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{FindByNameNotFound: true}
	ui := new(testterm.FakeUI)

	orgReq := newOrganizationRequirement("foo", ui, orgRepo)
	success := orgReq.Execute()

	assert.False(t, success)
}
