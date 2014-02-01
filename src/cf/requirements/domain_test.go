package requirements

import (
	"cf"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
	"testing"
)

func TestDomainReqExecute(t *testing.T) {
	domain := cf.Domain{}
	domain.Name = "example.com"
	domain.Guid = "domain-guid"
	domainRepo := &testapi.FakeDomainRepository{FindByNameDomain: domain}
	ui := new(testterm.FakeUI)

	domainReq := newDomainRequirement("example.com", ui, domainRepo)
	success := domainReq.Execute()

	assert.True(t, success)
	assert.Equal(t, domainRepo.FindByNameInCurrentSpaceName, "example.com")
	assert.Equal(t, domainReq.GetDomain(), domain)
}

func TestDomainReqWhenDomainDoesNotExist(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{FindByNameNotFound: true}
	ui := new(testterm.FakeUI)

	domainReq := newDomainRequirement("example.com", ui, domainRepo)

	testassert.AssertPanic(t, testterm.FailedWasCalled, func() {
		domainReq.Execute()
	})
}

func TestDomainReqOnError(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{FindByNameErr: true}
	ui := new(testterm.FakeUI)

	domainReq := newDomainRequirement("example.com", ui, domainRepo)

	testassert.AssertPanic(t, testterm.FailedWasCalled, func() {
		domainReq.Execute()
	})
}
