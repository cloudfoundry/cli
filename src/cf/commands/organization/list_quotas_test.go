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

func callListQuotas(t mr.TestingT, reqFactory *testreq.FakeReqFactory, quotaRepo *testapi.FakeQuotaRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("quotas", []string{})

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

	cmd := organization.NewListQuotas(fakeUI, config, quotaRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestListQuotasRequirements", func() {
			quotaRepo := &testapi.FakeQuotaRepository{}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			callListQuotas(mr.T(), reqFactory, quotaRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
			callListQuotas(mr.T(), reqFactory, quotaRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestListQuotas", func() {

			quota := cf.QuotaFields{}
			quota.Name = "quota-name"
			quota.MemoryLimit = 1024

			quotaRepo := &testapi.FakeQuotaRepository{FindAllQuotas: []cf.QuotaFields{quota}}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			ui := callListQuotas(mr.T(), reqFactory, quotaRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting quotas as", "my-user"},
				{"OK"},
				{"name", "memory limit"},
				{"quota-name", "1g"},
			})
		})
	})
}
