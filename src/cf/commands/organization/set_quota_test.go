package organization_test

import (
	"cf"
	"cf/commands/organization"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestSetQuotaFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	quotaRepo := &testapi.FakeQuotaRepository{}

	ui := callSetQuota(t, []string{}, reqFactory, quotaRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetQuota(t, []string{"org"}, reqFactory, quotaRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetQuota(t, []string{"org", "quota"}, reqFactory, quotaRepo)
	assert.False(t, ui.FailedWithUsage)

	ui = callSetQuota(t, []string{"org", "quota", "extra-stuff"}, reqFactory, quotaRepo)
	assert.True(t, ui.FailedWithUsage)
}

func TestSetQuotaRequirements(t *testing.T) {
	quotaRepo := &testapi.FakeQuotaRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callSetQuota(t, []string{"my-org", "my-quota"}, reqFactory, quotaRepo)

	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.OrganizationName, "my-org")

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
	callSetQuota(t, []string{"my-org", "my-quota"}, reqFactory, quotaRepo)

	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestSetQuota(t *testing.T) {
	org := cf.Organization{}
	org.Name = "my-org"
	org.Guid = "my-org-guid"

	quota := cf.QuotaFields{}
	quota.Name = "my-found-quota"
	quota.Guid = "my-quota-guid"

	quotaRepo := &testapi.FakeQuotaRepository{FindByNameQuota: quota}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Organization: org}

	ui := callSetQuota(t, []string{"my-org", "my-quota"}, reqFactory, quotaRepo)

	assert.Equal(t, quotaRepo.FindByNameName, "my-quota")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Setting quota", "my-found-quota", "my-org", "my-user"},
		{"OK"},
	})

	assert.Equal(t, quotaRepo.UpdateOrgGuid, "my-org-guid")
	assert.Equal(t, quotaRepo.UpdateQuotaGuid, "my-quota-guid")
}

func callSetQuota(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, quotaRepo *testapi.FakeQuotaRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-quota", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	spaceFields := cf.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := cf.OrganizationFields{}
	orgFields.Name = "my-org"

	config := &configuration.Configuration{
		SpaceFields:        spaceFields,
		OrganizationFields: orgFields,
		AccessToken:        token,
	}

	cmd := organization.NewSetQuota(ui, config, quotaRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
