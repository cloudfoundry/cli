package requirements_test

import (
	"cf/configuration"
	"cf/models"
	"cf/net"
	. "cf/requirements"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	var config configuration.ReadWriter
	var ui *testterm.FakeUI

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		config = testconfig.NewRepository()
		config.SetOrganizationFields(models.OrganizationFields{Guid: "the-org-guid"})
	})

	It("succeeds when the domain is found", func() {
		domain := models.DomainFields{Name: "example.com", Guid: "domain-guid"}
		domainRepo := &testapi.FakeDomainRepository{FindByNameInOrgDomain: domain}
		domainReq := NewDomainRequirement("example.com", ui, config, domainRepo)
		success := domainReq.Execute()

		Expect(success).To(BeTrue())
		Expect(domainRepo.FindByNameInOrgName).To(Equal("example.com"))
		Expect(domainRepo.FindByNameInOrgGuid).To(Equal("the-org-guid"))
		Expect(domainReq.GetDomain()).To(Equal(domain))
	})

	It("fails when the domain is not found", func() {
		domainRepo := &testapi.FakeDomainRepository{FindByNameInOrgApiResponse: net.NewNotFoundApiResponse("")}
		domainReq := NewDomainRequirement("example.com", ui, config, domainRepo)

		testassert.AssertPanic(testterm.FailedWasCalled, func() {
			domainReq.Execute()
		})
	})

	It("fails when an error occurs fetching the domain", func() {
		domainRepo := &testapi.FakeDomainRepository{FindByNameInOrgApiResponse: net.NewApiResponseWithError("", errors.New(""))}
		domainReq := NewDomainRequirement("example.com", ui, config, domainRepo)

		testassert.AssertPanic(testterm.FailedWasCalled, func() {
			domainReq.Execute()
		})
	})
})
