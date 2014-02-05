package requirements_test

import (
	"cf"
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestDomainReqExecute", func() {

			domain := cf.Domain{}
			domain.Name = "example.com"
			domain.Guid = "domain-guid"
			domainRepo := &testapi.FakeDomainRepository{FindByNameDomain: domain}
			ui := new(testterm.FakeUI)

			domainReq := NewDomainRequirement("example.com", ui, domainRepo)
			success := domainReq.Execute()

			assert.True(mr.T(), success)
			assert.Equal(mr.T(), domainRepo.FindByNameInCurrentSpaceName, "example.com")
			assert.Equal(mr.T(), domainReq.GetDomain(), domain)
		})
		It("TestDomainReqWhenDomainDoesNotExist", func() {

			domainRepo := &testapi.FakeDomainRepository{FindByNameNotFound: true}
			ui := new(testterm.FakeUI)

			domainReq := NewDomainRequirement("example.com", ui, domainRepo)

			testassert.AssertPanic(mr.T(), testterm.FailedWasCalled, func() {
				domainReq.Execute()
			})
		})
		It("TestDomainReqOnError", func() {

			domainRepo := &testapi.FakeDomainRepository{FindByNameErr: true}
			ui := new(testterm.FakeUI)

			domainReq := NewDomainRequirement("example.com", ui, domainRepo)

			testassert.AssertPanic(mr.T(), testterm.FailedWasCalled, func() {
				domainReq.Execute()
			})
		})
	})
}
