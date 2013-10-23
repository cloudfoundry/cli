package domain_test

import (
	"cf"
	. "cf/commands/domain"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
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
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, Organization: cf.Organization{Name: "my-org", Guid: "my-org-guid"}}
	domainRepo := &testapi.FakeDomainRepository{
		FindAllByOrgDomains: []cf.Domain{
			{Name: "Domain1", Shared: true, Spaces: []cf.Space{}},
			{Name: "Domain2", Shared: false, Spaces: []cf.Space{
				{Name: "my-space"},
				{Name: "my-other-space"},
			}},
			{Name: "Domain3", Shared: false, Spaces: []cf.Space{}},
		},
	}
	fakeUI := callListDomains(t, []string{}, reqFactory, domainRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Getting domains in org")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[0], "my-user")
	assert.Contains(t, fakeUI.Outputs[1], "OK")

	assert.Contains(t, fakeUI.Outputs[4], "Domain1")
	assert.Contains(t, fakeUI.Outputs[4], "shared")

	assert.Contains(t, fakeUI.Outputs[5], "Domain2")
	assert.Contains(t, fakeUI.Outputs[5], "owned")
	assert.Contains(t, fakeUI.Outputs[5], "my-space, my-other-space")

	assert.Contains(t, fakeUI.Outputs[6], "Domain3")
	assert.Contains(t, fakeUI.Outputs[6], "reserved")
}

func callListDomains(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("domains", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewListDomains(fakeUI, config, domainRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
