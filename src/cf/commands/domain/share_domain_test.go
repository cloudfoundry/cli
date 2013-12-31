package domain_test

import (
	. "cf/commands/domain"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestShareDomainRequirements(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callShareDomain(t, []string{"example.com"}, reqFactory, domainRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
	callShareDomain(t, []string{"example.com"}, reqFactory, domainRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestShareDomainFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	domainRepo := &testapi.FakeDomainRepository{}
	ui := callShareDomain(t, []string{}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callShareDomain(t, []string{"example.com"}, reqFactory, domainRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestShareDomain(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	domainRepo := &testapi.FakeDomainRepository{}
	ui := callShareDomain(t, []string{"example.com"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.CreateSharedDomainName, "example.com")
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Sharing domain", "example.com", "my-user"},
		{"OK"},
	})
}

func callShareDomain(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("share-domain", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		AccessToken: token,
	}

	cmd := NewShareDomain(fakeUI, config, domainRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
