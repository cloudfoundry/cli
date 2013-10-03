package requirements_test

import (
	"cf"
	. "cf/requirements"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestDomainReqExecute(t *testing.T) {
	domain := cf.Domain{Name: "example.com", Guid: "domain-guid"}
	domainRepo := &testhelpers.FakeDomainRepository{FindByNameDomain: domain}
	ui := new(testhelpers.FakeUI)

	domainReq := NewDomainRequirement("example.com", ui, domainRepo)
	success := domainReq.Execute()

	assert.True(t, success)
	assert.Equal(t, domainRepo.FindByNameName, "example.com")
	assert.Equal(t, domainReq.GetDomain(), domain)
}

func TestDomainReqWhenDomainDoesNotExist(t *testing.T) {
	domainRepo := &testhelpers.FakeDomainRepository{FindByNameNotFound: true}
	ui := new(testhelpers.FakeUI)

	domainReq := NewDomainRequirement("example.com", ui, domainRepo)
	success := domainReq.Execute()

	assert.False(t, success)
}

func TestDomainReqOnError(t *testing.T) {
	domainRepo := &testhelpers.FakeDomainRepository{FindByNameErr: true}
	ui := new(testhelpers.FakeUI)

	domainReq := NewDomainRequirement("example.com", ui, domainRepo)
	success := domainReq.Execute()

	assert.False(t, success)
}
