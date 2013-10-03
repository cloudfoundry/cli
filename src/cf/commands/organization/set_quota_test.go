package organization_test

import (
	"cf"
	. "cf/commands/organization"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestSetQuotaFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}
	orgRepo := &testhelpers.FakeOrgRepository{}

	fakeUI := callSetQuota([]string{}, reqFactory, orgRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callSetQuota([]string{"org"}, reqFactory, orgRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callSetQuota([]string{"org", "quota"}, reqFactory, orgRepo)
	assert.False(t, fakeUI.FailedWithUsage)

	fakeUI = callSetQuota([]string{"org", "quota", "extra-stuff"}, reqFactory, orgRepo)
	assert.True(t, fakeUI.FailedWithUsage)
}

func TestSetQuotaRequirements(t *testing.T) {
	orgRepo := &testhelpers.FakeOrgRepository{}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	callSetQuota([]string{"my-org", "my-quota"}, reqFactory, orgRepo)

	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.OrganizationName, "my-org")

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false}
	callSetQuota([]string{"my-org", "my-quota"}, reqFactory, orgRepo)

	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestSetQuota(t *testing.T) {
	org := cf.Organization{Name: "my-org", Guid: "my-org-guid"}
	quota := cf.Quota{Name: "my-found-quota", Guid: "my-quota-guid"}
	orgRepo := &testhelpers.FakeOrgRepository{FindQuotaByNameQuota: quota}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, Organization: org}

	ui := callSetQuota([]string{"my-org", "my-quota"}, reqFactory, orgRepo)

	assert.Equal(t, orgRepo.FindQuotaByNameName, "my-quota")

	assert.Contains(t, ui.Outputs[0], "Setting quota")
	assert.Contains(t, ui.Outputs[0], "my-found-quota")
	assert.Contains(t, ui.Outputs[0], "my-org")

	assert.Equal(t, orgRepo.UpdateQuotaOrg, org)
	assert.Equal(t, orgRepo.UpdateQuotaQuota, quota)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callSetQuota(args []string, reqFactory *testhelpers.FakeReqFactory, orgRepo *testhelpers.FakeOrgRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("set-quota", args)
	cmd := NewSetQuota(ui, orgRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
