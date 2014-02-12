package organization_test

import (
	"cf/commands/organization"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	spaceFields := models.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := models.OrganizationFields{}
	orgFields.Name = "my-org"

	token := configuration.TokenInfo{Username: "my-user"}
	config := testconfig.NewRepositoryWithAccessToken(token)
	config.SetSpaceFields(spaceFields)
	config.SetOrganizationFields(orgFields)

	cmd := organization.NewListQuotas(fakeUI, config, quotaRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestListQuotasRequirements", func() {
		quotaRepo := &testapi.FakeQuotaRepository{}

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		callListQuotas(mr.T(), reqFactory, quotaRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
		callListQuotas(mr.T(), reqFactory, quotaRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestListQuotas", func() {

		quota := models.QuotaFields{}
		quota.Name = "quota-name"
		quota.MemoryLimit = 1024

		quotaRepo := &testapi.FakeQuotaRepository{FindAllQuotas: []models.QuotaFields{quota}}
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
