package domain_test

import (
	. "cf/commands/domain"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestShareDomainRequirements", func() {
		domainRepo := &testapi.FakeDomainRepository{}

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		callShareDomain([]string{"example.com"}, reqFactory, domainRepo)
		assert.True(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
		callShareDomain([]string{"example.com"}, reqFactory, domainRepo)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)
	})
	It("TestShareDomainFailsWithUsage", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		domainRepo := &testapi.FakeDomainRepository{}
		ui := callShareDomain([]string{}, reqFactory, domainRepo)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callShareDomain([]string{"example.com"}, reqFactory, domainRepo)
		assert.False(mr.T(), ui.FailedWithUsage)
	})
	It("TestShareDomain", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		domainRepo := &testapi.FakeDomainRepository{}
		ui := callShareDomain([]string{"example.com"}, reqFactory, domainRepo)

		Expect(domainRepo.CreateSharedDomainName).To(Equal("example.com"))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Creating shared domain", "example.com", "my-user"},
			{"OK"},
		})
	})
})

func callShareDomain(args []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-shared-domain", args)
	configRepo := testconfig.NewRepositoryWithAccessToken(configuration.TokenInfo{Username: "my-user"})
	cmd := NewCreateSharedDomain(fakeUI, configRepo, domainRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
