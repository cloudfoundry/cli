package organization_test

import (
	"cf"
	. "cf/commands/organization"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestSetQuotaFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	orgRepo := &testapi.FakeOrgRepository{}

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
	orgRepo := &testapi.FakeOrgRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callSetQuota([]string{"my-org", "my-quota"}, reqFactory, orgRepo)

	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.OrganizationName, "my-org")

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
	callSetQuota([]string{"my-org", "my-quota"}, reqFactory, orgRepo)

	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestSetQuota(t *testing.T) {
	org := cf.Organization{Name: "my-org", Guid: "my-org-guid"}
	quota := cf.Quota{Name: "my-found-quota", Guid: "my-quota-guid"}
	orgRepo := &testapi.FakeOrgRepository{FindQuotaByNameQuota: quota}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Organization: org}

	ui := callSetQuota([]string{"my-org", "my-quota"}, reqFactory, orgRepo)

	assert.Equal(t, orgRepo.FindQuotaByNameName, "my-quota")

	assert.Contains(t, ui.Outputs[0], "Setting quota")
	assert.Contains(t, ui.Outputs[0], "my-found-quota")
	assert.Contains(t, ui.Outputs[0], "my-org")

	assert.Equal(t, orgRepo.UpdateQuotaOrg, org)
	assert.Equal(t, orgRepo.UpdateQuotaQuota, quota)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callSetQuota(args []string, reqFactory *testreq.FakeReqFactory, orgRepo *testapi.FakeOrgRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-quota", args)
	cmd := NewSetQuota(ui, orgRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
