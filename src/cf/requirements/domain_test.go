package requirements_test

import (
	"cf/configuration"
	"cf/models"
	"cf/net"
	. "cf/requirements"
	"errors"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	var config *configuration.Configuration
	var ui *testterm.FakeUI

	BeforeEach(func() {
		config = &configuration.Configuration{
			OrganizationFields: models.OrganizationFields{Guid: "the-org-guid"},
		}
		ui = new(testterm.FakeUI)
	})

	It("succeeds when the domain is found", func() {
		domain := models.DomainFields{Name: "example.com", Guid: "domain-guid"}
		domainRepo := &testapi.FakeDomainRepository{FindByNameInOrgDomain: domain}
		domainReq := NewDomainRequirement("example.com", ui, config, domainRepo)
		success := domainReq.Execute()

		assert.True(mr.T(), success)
		assert.Equal(mr.T(), domainRepo.FindByNameInOrgName, "example.com")
		assert.Equal(mr.T(), domainRepo.FindByNameInOrgGuid, "the-org-guid")
		assert.Equal(mr.T(), domainReq.GetDomain(), domain)
	})

	It("fails when the domain is not found", func() {
		domainRepo := &testapi.FakeDomainRepository{FindByNameInOrgApiResponse: net.NewNotFoundApiResponse("")}
		domainReq := NewDomainRequirement("example.com", ui, config, domainRepo)

		testassert.AssertPanic(mr.T(), testterm.FailedWasCalled, func() {
			domainReq.Execute()
		})
	})

	It("fails when an error occurs fetching the domain", func() {
		domainRepo := &testapi.FakeDomainRepository{FindByNameInOrgApiResponse: net.NewApiResponseWithError("", errors.New(""))}
		domainReq := NewDomainRequirement("example.com", ui, config, domainRepo)

		testassert.AssertPanic(mr.T(), testterm.FailedWasCalled, func() {
			domainReq.Execute()
		})
	})
})
