package domain_test

import (
	"cf"
	. "cf/commands/domain"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestListDomainsRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	domainRepo := &testapi.FakeDomainRepository{}

	callListDomains([]string{}, reqFactory, domainRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
	callListDomains([]string{}, reqFactory, domainRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	callListDomains([]string{}, reqFactory, domainRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestListDomainsFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	domainRepo := &testapi.FakeDomainRepository{}

	ui := callListDomains([]string{"foo"}, reqFactory, domainRepo)
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
	fakeUI := callListDomains([]string{}, reqFactory, domainRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Getting domains in org")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[1], "OK")

	assert.Contains(t, fakeUI.Outputs[3], "Domain1")
	assert.Contains(t, fakeUI.Outputs[3], "shared")

	assert.Contains(t, fakeUI.Outputs[4], "Domain2")
	assert.Contains(t, fakeUI.Outputs[4], "owned")
	assert.Contains(t, fakeUI.Outputs[4], "my-space, my-other-space")

	assert.Contains(t, fakeUI.Outputs[5], "Domain3")
	assert.Contains(t, fakeUI.Outputs[5], "reserved")
}

func callListDomains(args []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("domains", args)
	cmd := NewListDomains(fakeUI, domainRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
