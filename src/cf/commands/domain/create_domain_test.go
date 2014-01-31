package domain_test

import (
	"cf"
	"cf/commands/domain"
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

func callCreateDomain(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-domain", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		AccessToken: token,
	}

	cmd := domain.NewCreateDomain(fakeUI, config, domainRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestCreateDomainRequirements", func() {
			domainRepo := &testapi.FakeDomainRepository{}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

			callCreateDomain(mr.T(), []string{"my-org", "example.com"}, reqFactory, domainRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
			assert.Equal(mr.T(), reqFactory.OrganizationName, "my-org")

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}

			callCreateDomain(mr.T(), []string{"my-org", "example.com"}, reqFactory, domainRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestCreateDomainFailsWithUsage", func() {

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			domainRepo := &testapi.FakeDomainRepository{}
			ui := callCreateDomain(mr.T(), []string{""}, reqFactory, domainRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callCreateDomain(mr.T(), []string{"org1"}, reqFactory, domainRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callCreateDomain(mr.T(), []string{"org1", "example.com"}, reqFactory, domainRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestCreateDomain", func() {

			org := cf.Organization{}
			org.Name = "myOrg"
			org.Guid = "myOrg-guid"
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Organization: org}
			domainRepo := &testapi.FakeDomainRepository{}
			ui := callCreateDomain(mr.T(), []string{"myOrg", "example.com"}, reqFactory, domainRepo)

			assert.Equal(mr.T(), domainRepo.CreateDomainName, "example.com")
			assert.Equal(mr.T(), domainRepo.CreateDomainOwningOrgGuid, "myOrg-guid")
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating domain", "example.com", "myOrg", "my-user"},
				{"OK"},
			})
		})
	})
}
