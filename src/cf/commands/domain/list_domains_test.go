package domain_test

import (
	"cf"
	"cf/commands/domain"
	"cf/configuration"
	"cf/net"
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

func callListDomains(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("domains", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	spaceFields := cf.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := cf.OrganizationFields{}
	orgFields.Name = "my-org"

	config := &configuration.Configuration{
		SpaceFields:        spaceFields,
		OrganizationFields: orgFields,
		AccessToken:        token,
	}

	cmd := domain.NewListDomains(fakeUI, config, domainRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestListDomainsRequirements", func() {
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
			domainRepo := &testapi.FakeDomainRepository{}

			callListDomains(mr.T(), []string{}, reqFactory, domainRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
			callListDomains(mr.T(), []string{}, reqFactory, domainRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
			callListDomains(mr.T(), []string{}, reqFactory, domainRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestListDomainsFailsWithUsage", func() {

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
			domainRepo := &testapi.FakeDomainRepository{}

			ui := callListDomains(mr.T(), []string{"foo"}, reqFactory, domainRepo)
			assert.True(mr.T(), ui.FailedWithUsage)
		})
		It("TestListDomains", func() {

			orgFields := cf.OrganizationFields{}
			orgFields.Name = "my-org"
			orgFields.Guid = "my-org-guid"

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, OrganizationFields: orgFields}
			domain1 := cf.Domain{}
			domain1.Shared = true
			domain1.Name = "Domain1"

			domain2 := cf.Domain{}
			domain2.Shared = false
			domain2.Name = "Domain2"

			domain3 := cf.Domain{}
			domain3.Shared = false
			domain3.Name = "Domain3"

			domainRepo := &testapi.FakeDomainRepository{
				ListSharedDomainsDomains: []cf.Domain{domain1},
				ListDomainsForOrgDomains: []cf.Domain{domain2, domain3},
			}

			ui := callListDomains(mr.T(), []string{}, reqFactory, domainRepo)

			assert.Equal(mr.T(), domainRepo.ListDomainsForOrgDomainsGuid, "my-org-guid")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting domains in org", "my-org", "my-user"},
				{"name", "status"},
				{"Domain1", "shared"},
				{"Domain2", "owned"},
				{"Domain3", "owned"},
			})
		})
		It("TestListDomainsWhenThereAreNone", func() {

			orgFields := cf.OrganizationFields{}
			orgFields.Name = "my-org"
			orgFields.Guid = "my-org-guid"

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, OrganizationFields: orgFields}
			domainRepo := &testapi.FakeDomainRepository{}

			ui := callListDomains(mr.T(), []string{}, reqFactory, domainRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting domains in org", "my-org", "my-user"},
				{"No domains found"},
			})
		})
		It("TestListDomainsSharedDomainsFails", func() {

			orgFields := cf.OrganizationFields{}
			orgFields.Name = "my-org"
			orgFields.Guid = "my-org-guid"

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, OrganizationFields: orgFields}

			domainRepo := &testapi.FakeDomainRepository{
				ListSharedDomainsApiResponse: net.NewApiResponseWithMessage("borked!"),
			}
			ui := callListDomains(mr.T(), []string{}, reqFactory, domainRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting domains in org", "my-org", "my-user"},
				{"failed"},
				{"shared domains"},
				{"borked!"},
			})
		})
		It("TestListDomainsSharedDomainsTriesOldEndpointOn404", func() {

			orgFields := cf.OrganizationFields{}
			orgFields.Name = "my-org"
			orgFields.Guid = "my-org-guid"

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, OrganizationFields: orgFields}

			domain := cf.Domain{}
			domain.Name = "ze-domain"

			domainRepo := &testapi.FakeDomainRepository{
				ListSharedDomainsApiResponse: net.NewNotFoundApiResponse("whoops! misplaced yr domainz"),
				ListDomainsForOrgDomains:     []cf.Domain{domain},
			}
			ui := callListDomains(mr.T(), []string{}, reqFactory, domainRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting domains in org", "my-org", "my-user"},
				{"ze-domain"},
			})
		})
		It("TestListDomainsOrgDomainsFails", func() {

			orgFields := cf.OrganizationFields{}
			orgFields.Name = "my-org"
			orgFields.Guid = "my-org-guid"

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, OrganizationFields: orgFields}

			domainRepo := &testapi.FakeDomainRepository{
				ListDomainsForOrgApiResponse: net.NewApiResponseWithMessage("borked!"),
			}
			ui := callListDomains(mr.T(), []string{}, reqFactory, domainRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting domains in org", "my-org", "my-user"},
				{"failed"},
				{"private domains"},
				{"borked!"},
			})
		})
	})
}
