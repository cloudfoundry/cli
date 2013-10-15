package requirements

import (
	"cf"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testterm "testhelpers/terminal"
	"testing"
)

func TestDomainReqExecute(t *testing.T) {
	domain := cf.Domain{Name: "example.com", Guid: "domain-guid"}
	domainRepo := &testapi.FakeDomainRepository{FindByNameDomain: domain}
	ui := new(testterm.FakeUI)

	domainReq := newDomainRequirement("example.com", ui, domainRepo)
	success := domainReq.Execute()

	assert.True(t, success)
	assert.Equal(t, domainRepo.FindByNameName, "example.com")
	assert.Equal(t, domainReq.GetDomain(), domain)
}

func TestDomainReqWhenDomainDoesNotExist(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{FindByNameNotFound: true}
	ui := new(testterm.FakeUI)

	domainReq := newDomainRequirement("example.com", ui, domainRepo)
	success := domainReq.Execute()

	assert.False(t, success)
}

func TestDomainReqOnError(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{FindByNameErr: true}
	ui := new(testterm.FakeUI)

	domainReq := newDomainRequirement("example.com", ui, domainRepo)
	success := domainReq.Execute()

	assert.False(t, success)
}
