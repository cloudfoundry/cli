package domain_test

import (
	"cf"
	. "cf/commands/domain"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestListDomainsRequirements(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	domainRepo := &testhelpers.FakeDomainRepository{}

	callListDomains([]string{}, reqFactory, domainRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
	callListDomains([]string{}, reqFactory, domainRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	callListDomains([]string{}, reqFactory, domainRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestListDomainsFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	domainRepo := &testhelpers.FakeDomainRepository{}

	ui := callListDomains([]string{"foo"}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)
}

func TestListDomains(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, Organization: cf.Organization{Name: "my-org", Guid: "my-org-guid"}}
	domainRepo := &testhelpers.FakeDomainRepository{
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

func callListDomains(args []string, reqFactory *testhelpers.FakeReqFactory, domainRepo *testhelpers.FakeDomainRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("domains", args)
	cmd := NewListDomains(fakeUI, domainRepo)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
