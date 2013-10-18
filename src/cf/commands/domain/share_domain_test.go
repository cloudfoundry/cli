package domain_test

import (
	. "cf/commands/domain"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestShareDomainRequirements(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callShareDomain([]string{"example.com"}, reqFactory, domainRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
	callShareDomain([]string{"example.com"}, reqFactory, domainRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestShareDomainFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	domainRepo := &testapi.FakeDomainRepository{}
	ui := callShareDomain([]string{}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callShareDomain([]string{"example.com"}, reqFactory, domainRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestShareDomain(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	domainRepo := &testapi.FakeDomainRepository{}
	fakeUI := callShareDomain([]string{"example.com"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.CreateSharedDomainDomain.Name, "example.com")
	assert.Contains(t, fakeUI.Outputs[0], "example.com")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func callShareDomain(args []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("share-domain", args)
	cmd := NewShareDomain(fakeUI, domainRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
