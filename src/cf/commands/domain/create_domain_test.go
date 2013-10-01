package domain_test

import (
	"cf"
	. "cf/commands/domain"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestCreateDomain(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	domainRepo := &testhelpers.FakeDomainRepository{}
	orgRepo := &testhelpers.FakeOrgRepository{FindByNameOrganization: cf.Organization{Name: "myOrg", Guid: "myOrg-guid"}}
	fakeUI := callCreateDomain([]string{"example.com", "myOrg"}, reqFactory, domainRepo, orgRepo)

	assert.Equal(t, domainRepo.CreateDomainDomainToCreate.Name, "example.com")
	assert.Equal(t, domainRepo.CreateDomainOwningOrg.Name, "myOrg")
	assert.Contains(t, fakeUI.Outputs[0], "example.com")
	assert.Contains(t, fakeUI.Outputs[0], "myOrg")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestCreateDomainFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	domainRepo := &testhelpers.FakeDomainRepository{}
	orgRepo := &testhelpers.FakeOrgRepository{}
	ui := callCreateDomain([]string{""}, reqFactory, domainRepo, orgRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateDomain([]string{"example.com"}, reqFactory, domainRepo, orgRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateDomain([]string{"example.com", "org1"}, reqFactory, domainRepo, orgRepo)
	assert.False(t, ui.FailedWithUsage)
}

func callCreateDomain(args []string, reqFactory *testhelpers.FakeReqFactory, domainRepo *testhelpers.FakeDomainRepository, orgRepo *testhelpers.FakeOrgRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("create-domain", args)
	cmd := NewCreateDomain(fakeUI, domainRepo, orgRepo)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
