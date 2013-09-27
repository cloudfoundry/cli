package commands_test

import (
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestCreateSharedDomain(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	domainRepo := &testhelpers.FakeDomainRepository{}
	fakeUI := callCreateSharedDomain([]string{"example.com", "--shared"}, reqFactory, domainRepo)

	assert.Equal(t, len(fakeUI.Outputs), 2)
	assert.Equal(t, domainRepo.CreateDomainName, "example.com")
	assert.Contains(t, fakeUI.Outputs[0], "Creating shared domain")
	assert.Contains(t, fakeUI.Outputs[0], "example.com")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestCreateSharedDomainFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	domainRepo := &testhelpers.FakeDomainRepository{}
	ui := callCreateSharedDomain([]string{"--shared"}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateSharedDomain([]string{"example.com", "--shared"}, reqFactory, domainRepo)
	assert.False(t, ui.FailedWithUsage)
}

func callCreateSharedDomain(args []string, reqFactory *testhelpers.FakeReqFactory, domainRepo *testhelpers.FakeDomainRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("create-domain", args)
	cmd := NewCreateSharedDomain(fakeUI, domainRepo)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
