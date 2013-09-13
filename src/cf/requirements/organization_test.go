package requirements_test

import (
	"cf"
	. "cf/requirements"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestOrgReqExecute(t *testing.T) {
	org := cf.Organization{Name: "my-org", Guid: "my-org-guid"}
	orgRepo := &testhelpers.FakeOrgRepository{FindByNameOrganization: org}
	ui := new(testhelpers.FakeUI)

	orgReq := NewOrganizationRequirement("foo", ui, orgRepo)
	success := orgReq.Execute()

	assert.True(t, success)
	assert.Equal(t, orgRepo.FindByNameName, "foo")
	assert.Equal(t, orgReq.GetOrganization(), org)
}
