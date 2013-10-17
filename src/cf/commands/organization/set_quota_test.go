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
	quotaRepo := &testapi.FakeQuotaRepository{}

	fakeUI := callSetQuota([]string{}, reqFactory, quotaRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callSetQuota([]string{"org"}, reqFactory, quotaRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callSetQuota([]string{"org", "quota"}, reqFactory, quotaRepo)
	assert.False(t, fakeUI.FailedWithUsage)

	fakeUI = callSetQuota([]string{"org", "quota", "extra-stuff"}, reqFactory, quotaRepo)
	assert.True(t, fakeUI.FailedWithUsage)
}

func TestSetQuotaRequirements(t *testing.T) {
	quotaRepo := &testapi.FakeQuotaRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callSetQuota([]string{"my-org", "my-quota"}, reqFactory, quotaRepo)

	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.OrganizationName, "my-org")

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
	callSetQuota([]string{"my-org", "my-quota"}, reqFactory, quotaRepo)

	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestSetQuota(t *testing.T) {
	org := cf.Organization{Name: "my-org", Guid: "my-org-guid"}
	quota := cf.Quota{Name: "my-found-quota", Guid: "my-quota-guid"}
	quotaRepo := &testapi.FakeQuotaRepository{FindByNameQuota: quota}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Organization: org}

	ui := callSetQuota([]string{"my-org", "my-quota"}, reqFactory, quotaRepo)

	assert.Equal(t, quotaRepo.FindByNameName, "my-quota")

	assert.Contains(t, ui.Outputs[0], "Setting quota")
	assert.Contains(t, ui.Outputs[0], "my-found-quota")
	assert.Contains(t, ui.Outputs[0], "my-org")

	assert.Equal(t, quotaRepo.UpdateOrg, org)
	assert.Equal(t, quotaRepo.UpdateQuota, quota)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callSetQuota(args []string, reqFactory *testreq.FakeReqFactory, quotaRepo *testapi.FakeQuotaRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-quota", args)
	cmd := NewSetQuota(ui, quotaRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
