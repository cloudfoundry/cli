package domain_test

import (
	"cf"
	. "cf/commands/domain"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestReserveDomainRequirements(t *testing.T) {
	domainRepo := &testhelpers.FakeDomainRepository{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}

	callReserveDomain([]string{"my-org", "example.com"}, reqFactory, domainRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.OrganizationName, "my-org")

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false}

	callReserveDomain([]string{"my-org", "example.com"}, reqFactory, domainRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestReserveDomainFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	domainRepo := &testhelpers.FakeDomainRepository{}
	ui := callReserveDomain([]string{""}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callReserveDomain([]string{"org1"}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callReserveDomain([]string{"org1", "example.com"}, reqFactory, domainRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestReserveDomain(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, Organization: cf.Organization{Name: "myOrg", Guid: "myOrg-guid"}}
	domainRepo := &testhelpers.FakeDomainRepository{}
	fakeUI := callReserveDomain([]string{"myOrg", "example.com"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.ReserveDomainDomainToCreate.Name, "example.com")
	assert.Equal(t, domainRepo.ReserveDomainOwningOrg.Name, "myOrg")
	assert.Contains(t, fakeUI.Outputs[0], "example.com")
	assert.Contains(t, fakeUI.Outputs[0], "myOrg")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func callReserveDomain(args []string, reqFactory *testhelpers.FakeReqFactory, domainRepo *testhelpers.FakeDomainRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("reserve-domain", args)
	cmd := NewReserveDomain(fakeUI, domainRepo)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
