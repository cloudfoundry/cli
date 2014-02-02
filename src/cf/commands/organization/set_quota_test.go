package organization_test

import (
	"cf"
	"cf/commands/organization"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callSetQuota(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, quotaRepo *testapi.FakeQuotaRepository) (ui *testterm.FakeUI) {
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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSetQuotaFailsWithUsage", func() {
			reqFactory := &testreq.FakeReqFactory{}
			quotaRepo := &testapi.FakeQuotaRepository{}

			ui := callSetQuota(mr.T(), []string{}, reqFactory, quotaRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callSetQuota(mr.T(), []string{"org"}, reqFactory, quotaRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callSetQuota(mr.T(), []string{"org", "quota"}, reqFactory, quotaRepo)
			assert.False(mr.T(), ui.FailedWithUsage)

			ui = callSetQuota(mr.T(), []string{"org", "quota", "extra-stuff"}, reqFactory, quotaRepo)
			assert.True(mr.T(), ui.FailedWithUsage)
		})
		It("TestSetQuotaRequirements", func() {

			quotaRepo := &testapi.FakeQuotaRepository{}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			callSetQuota(mr.T(), []string{"my-org", "my-quota"}, reqFactory, quotaRepo)

			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
			assert.Equal(mr.T(), reqFactory.OrganizationName, "my-org")

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
			callSetQuota(mr.T(), []string{"my-org", "my-quota"}, reqFactory, quotaRepo)

			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestSetQuota", func() {

			org := cf.Organization{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"

			quota := cf.QuotaFields{}
			quota.Name = "my-found-quota"
			quota.Guid = "my-quota-guid"

			quotaRepo := &testapi.FakeQuotaRepository{FindByNameQuota: quota}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Organization: org}

			ui := callSetQuota(mr.T(), []string{"my-org", "my-quota"}, reqFactory, quotaRepo)

			assert.Equal(mr.T(), quotaRepo.FindByNameName, "my-quota")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Setting quota", "my-found-quota", "my-org", "my-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), quotaRepo.UpdateOrgGuid, "my-org-guid")
			assert.Equal(mr.T(), quotaRepo.UpdateQuotaGuid, "my-quota-guid")
		})
	})
}
