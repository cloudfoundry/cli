package domain_test

import (
	"cf"
	"cf/commands/domain"
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

func TestCreateDomainRequirements(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	callCreateDomain(t, []string{"my-org", "example.com"}, reqFactory, domainRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.OrganizationName, "my-org")

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}

	callCreateDomain(t, []string{"my-org", "example.com"}, reqFactory, domainRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestCreateDomainFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	domainRepo := &testapi.FakeDomainRepository{}
	ui := callCreateDomain(t, []string{""}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateDomain(t, []string{"org1"}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateDomain(t, []string{"org1", "example.com"}, reqFactory, domainRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestCreateDomain(t *testing.T) {
	org := cf.Organization{}
	org.Name = "myOrg"
	org.Guid = "myOrg-guid"
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Organization: org}
	domainRepo := &testapi.FakeDomainRepository{}
	ui := callCreateDomain(t, []string{"myOrg", "example.com"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.CreateDomainName, "example.com")
	assert.Equal(t, domainRepo.CreateDomainOwningOrgGuid, "myOrg-guid")
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating domain", "example.com", "myOrg", "my-user"},
		{"OK"},
	})
}

func callCreateDomain(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (fakeUI *testterm.FakeUI) {
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
