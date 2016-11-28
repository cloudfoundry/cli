package requirements_test

import (
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	. "code.cloudfoundry.org/cli/cf/requirements"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DomainRequirement", func() {
	var config coreconfig.ReadWriter

	BeforeEach(func() {
		config = testconfig.NewRepository()
		config.SetOrganizationFields(models.OrganizationFields{GUID: "the-org-guid"})
	})

	It("succeeds when the domain is found", func() {
		domain := models.DomainFields{Name: "example.com", GUID: "domain-guid"}
		domainRepo := new(apifakes.FakeDomainRepository)
		domainRepo.FindByNameInOrgReturns(domain, nil)
		domainReq := NewDomainRequirement("example.com", config, domainRepo)
		err := domainReq.Execute()

		Expect(err).NotTo(HaveOccurred())
		orgName, orgGUID := domainRepo.FindByNameInOrgArgsForCall(0)
		Expect(orgName).To(Equal("example.com"))
		Expect(orgGUID).To(Equal("the-org-guid"))
		Expect(domainReq.GetDomain()).To(Equal(domain))
	})

	It("fails when the domain is not found", func() {
		domainRepo := new(apifakes.FakeDomainRepository)
		domainRepo.FindByNameInOrgReturns(models.DomainFields{}, errors.NewModelNotFoundError("Domain", ""))
		domainReq := NewDomainRequirement("example.com", config, domainRepo)

		err := domainReq.Execute()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Domain"))
		Expect(err.Error()).To(ContainSubstring("not found"))
	})

	It("fails when an error occurs fetching the domain", func() {
		domainRepo := new(apifakes.FakeDomainRepository)
		domainRepo.FindByNameInOrgReturns(models.DomainFields{}, errors.New("an-error"))
		domainReq := NewDomainRequirement("example.com", config, domainRepo)

		err := domainReq.Execute()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("an-error"))
	})
})
