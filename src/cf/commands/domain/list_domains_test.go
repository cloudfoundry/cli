package domain_test

import (
	"cf"
	"cf/commands/domain"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestListDomainsRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	domainRepo := &testapi.FakeDomainRepository{}

	callListDomains(t, []string{}, reqFactory, domainRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
	callListDomains(t, []string{}, reqFactory, domainRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	callListDomains(t, []string{}, reqFactory, domainRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestListDomainsFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	domainRepo := &testapi.FakeDomainRepository{}

	ui := callListDomains(t, []string{"foo"}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)
}

func TestListDomains(t *testing.T) {
	orgFields := cf.OrganizationFields{}
	orgFields.Name = "my-org"
	orgFields.Guid = "my-org-guid"

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, OrganizationFields: orgFields}
	domain1 := cf.Domain{}
	domain1.Shared = true
	domain1.Name = "Domain1"

	domain2 := cf.Domain{}
	domain2.Shared = false
	domain2.Name = "Domain2"

	domain3 := cf.Domain{}
	domain3.Shared = false
	domain3.Name = "Domain3"

	domainRepo := &testapi.FakeDomainRepository{
		ListSharedDomainsDomains: []cf.Domain{domain1},
		ListDomainsForOrgDomains: []cf.Domain{domain2, domain3},
	}

	ui := callListDomains(t, []string{}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.ListDomainsForOrgDomainsGuid, "my-org-guid")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Getting domains in org", "my-org", "my-user"},
		{"name", "status"},
		{"Domain1", "shared"},
		{"Domain2", "owned"},
		{"Domain3", "owned"},
	})
}

func TestListDomainsWhenThereAreNone(t *testing.T) {
	orgFields := cf.OrganizationFields{}
	orgFields.Name = "my-org"
	orgFields.Guid = "my-org-guid"

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, OrganizationFields: orgFields}
	domainRepo := &testapi.FakeDomainRepository{}

	ui := callListDomains(t, []string{}, reqFactory, domainRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Getting domains in org", "my-org", "my-user"},
		{"No domains found"},
	})
}

func TestListDomainsSharedDomainsFails(t *testing.T) {
	orgFields := cf.OrganizationFields{}
	orgFields.Name = "my-org"
	orgFields.Guid = "my-org-guid"

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, OrganizationFields: orgFields}

	domainRepo := &testapi.FakeDomainRepository{
		ListSharedDomainsApiResponse: net.NewApiResponseWithMessage("borked!"),
	}
	ui := callListDomains(t, []string{}, reqFactory, domainRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Getting domains in org", "my-org", "my-user"},
		{"failed"},
		{"shared domains"},
		{"borked!"},
	})
}

func TestListDomainsOrgDomainsFails(t *testing.T) {
	orgFields := cf.OrganizationFields{}
	orgFields.Name = "my-org"
	orgFields.Guid = "my-org-guid"

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, OrganizationFields: orgFields}

	domainRepo := &testapi.FakeDomainRepository{
		ListDomainsForOrgApiResponse: net.NewApiResponseWithMessage("borked!"),
	}
	ui := callListDomains(t, []string{}, reqFactory, domainRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Getting domains in org", "my-org", "my-user"},
		{"failed"},
		{"private domains"},
		{"borked!"},
	})
}

func callListDomains(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("domains", args)

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

	cmd := domain.NewListDomains(fakeUI, config, domainRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
