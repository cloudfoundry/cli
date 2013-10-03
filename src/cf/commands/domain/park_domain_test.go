package domain_test

import (
	"cf"
	. "cf/commands/domain"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestParkDomainRequirements(t *testing.T) {
	domainRepo := &testhelpers.FakeDomainRepository{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}

	callParkDomain([]string{"example.com", "my-org"}, reqFactory, domainRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.OrganizationName, "my-org")

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false}

	callParkDomain([]string{"example.com", "my-org"}, reqFactory, domainRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestParkDomainFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	domainRepo := &testhelpers.FakeDomainRepository{}
	ui := callParkDomain([]string{""}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callParkDomain([]string{"example.com"}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callParkDomain([]string{"example.com", "org1"}, reqFactory, domainRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestParkDomain(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, Organization: cf.Organization{Name: "myOrg", Guid: "myOrg-guid"}}
	domainRepo := &testhelpers.FakeDomainRepository{}
	fakeUI := callParkDomain([]string{"example.com", "myOrg"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.ParkDomainDomainToCreate.Name, "example.com")
	assert.Equal(t, domainRepo.ParkDomainOwningOrg.Name, "myOrg")
	assert.Contains(t, fakeUI.Outputs[0], "example.com")
	assert.Contains(t, fakeUI.Outputs[0], "myOrg")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func callParkDomain(args []string, reqFactory *testhelpers.FakeReqFactory, domainRepo *testhelpers.FakeDomainRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("park-domain", args)
	cmd := NewParkDomain(fakeUI, domainRepo)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
