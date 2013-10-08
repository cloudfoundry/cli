package domain_test

import (
	. "cf/commands/domain"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestShareDomainRequirements(t *testing.T) {
	domainRepo := &testhelpers.FakeDomainRepository{}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	callShareDomain([]string{"example.com"}, reqFactory, domainRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false}
	callShareDomain([]string{"example.com"}, reqFactory, domainRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestShareDomainFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	domainRepo := &testhelpers.FakeDomainRepository{}
	ui := callShareDomain([]string{}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callShareDomain([]string{"example.com"}, reqFactory, domainRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestShareDomain(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	domainRepo := &testhelpers.FakeDomainRepository{}
	fakeUI := callShareDomain([]string{"example.com"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.ShareDomainDomainToCreate.Name, "example.com")
	assert.Contains(t, fakeUI.Outputs[0], "example.com")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func callShareDomain(args []string, reqFactory *testhelpers.FakeReqFactory, domainRepo *testhelpers.FakeDomainRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("share-domain", args)
	cmd := NewShareDomain(fakeUI, domainRepo)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
