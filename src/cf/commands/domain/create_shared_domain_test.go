package domain_test

import (
	. "cf/commands/domain"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callShareDomain(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-shared-domain", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		AccessToken: token,
	}

	cmd := NewCreateSharedDomain(fakeUI, config, domainRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestShareDomainRequirements", func() {
			domainRepo := &testapi.FakeDomainRepository{}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			callShareDomain(mr.T(), []string{"example.com"}, reqFactory, domainRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
			callShareDomain(mr.T(), []string{"example.com"}, reqFactory, domainRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestShareDomainFailsWithUsage", func() {

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			domainRepo := &testapi.FakeDomainRepository{}
			ui := callShareDomain(mr.T(), []string{}, reqFactory, domainRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callShareDomain(mr.T(), []string{"example.com"}, reqFactory, domainRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestShareDomain", func() {

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			domainRepo := &testapi.FakeDomainRepository{}
			ui := callShareDomain(mr.T(), []string{"example.com"}, reqFactory, domainRepo)

			assert.Equal(mr.T(), domainRepo.CreateSharedDomainName, "example.com")
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating shared domain", "example.com", "my-user"},
				{"OK"},
			})
		})
	})
}
