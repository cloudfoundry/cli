package domain_test

import (
	"cf"
	. "cf/commands/domain"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestCreateDomainRequirements(t *testing.T) {
	domainRepo := &testhelpers.FakeDomainRepository{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}

	callCreateDomain([]string{"example.com", "my-org"}, reqFactory, domainRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.OrganizationName, "my-org")

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false}

	callCreateDomain([]string{"example.com", "my-org"}, reqFactory, domainRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestCreateDomainFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	domainRepo := &testhelpers.FakeDomainRepository{}
	ui := callCreateDomain([]string{""}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateDomain([]string{"example.com"}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateDomain([]string{"example.com", "org1"}, reqFactory, domainRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestCreateDomain(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, Organization: cf.Organization{Name: "myOrg", Guid: "myOrg-guid"}}
	domainRepo := &testhelpers.FakeDomainRepository{}
	fakeUI := callCreateDomain([]string{"example.com", "myOrg"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.CreateDomainDomainToCreate.Name, "example.com")
	assert.Equal(t, domainRepo.CreateDomainOwningOrg.Name, "myOrg")
	assert.Contains(t, fakeUI.Outputs[0], "example.com")
	assert.Contains(t, fakeUI.Outputs[0], "myOrg")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func callCreateDomain(args []string, reqFactory *testhelpers.FakeReqFactory, domainRepo *testhelpers.FakeDomainRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("create-domain", args)
	cmd := NewCreateDomain(fakeUI, domainRepo)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
