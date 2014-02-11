package domain_test

import (
	"cf/commands/domain"
	"cf/configuration"
	"cf/models"
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

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestGetRequirements", func() {
			domainRepo := &testapi.FakeDomainRepository{}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

			callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
			callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
			callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestDeleteDomainSuccess", func() {

			domain := models.DomainFields{}
			domain.Name = "foo.com"
			domain.Guid = "foo-guid"
			domainRepo := &testapi.FakeDomainRepository{
				FindByNameInOrgDomain: domain,
			}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

			ui := callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

			assert.Equal(mr.T(), domainRepo.DeleteDomainGuid, "foo-guid")

			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
				{"delete", "foo.com"},
			})

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting domain", "foo.com", "my-user"},
				{"OK"},
			})
		})
		It("TestDeleteDomainNoConfirmation", func() {

			domain := models.DomainFields{}
			domain.Name = "foo.com"
			domain.Guid = "foo-guid"
			domainRepo := &testapi.FakeDomainRepository{
				FindByNameInOrgDomain: domain,
			}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

			ui := callDeleteDomain([]string{"foo.com"}, []string{"no"}, reqFactory, domainRepo)

			assert.Equal(mr.T(), domainRepo.DeleteDomainGuid, "")

			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
				{"delete", "foo.com"},
			})

			assert.Equal(mr.T(), len(ui.Outputs), 1)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting domain", "foo.com"},
			})
		})
		It("TestDeleteDomainNotFound", func() {

			domainRepo := &testapi.FakeDomainRepository{
				FindByNameInOrgApiResponse: net.NewNotFoundApiResponse("%s %s not found", "Domain", "foo.com"),
			}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

			ui := callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

			assert.Equal(mr.T(), domainRepo.DeleteDomainGuid, "")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting domain", "foo.com"},
				{"OK"},
				{"foo.com", "not found"},
			})
		})
		It("TestDeleteDomainFindError", func() {

			domainRepo := &testapi.FakeDomainRepository{
				FindByNameInOrgApiResponse: net.NewApiResponseWithMessage("failed badly"),
			}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

			ui := callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

			assert.Equal(mr.T(), domainRepo.DeleteDomainGuid, "")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting domain", "foo.com"},
				{"FAILED"},
				{"foo.com"},
				{"failed badly"},
			})
		})
		It("TestDeleteDomainDeleteError", func() {

			domain := models.DomainFields{}
			domain.Name = "foo.com"
			domain.Guid = "foo-guid"
			domainRepo := &testapi.FakeDomainRepository{
				FindByNameInOrgDomain: domain,
				DeleteApiResponse:     net.NewApiResponseWithMessage("failed badly"),
			}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

			ui := callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

			assert.Equal(mr.T(), domainRepo.DeleteDomainGuid, "foo-guid")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting domain", "foo.com"},
				{"FAILED"},
				{"foo.com"},
				{"failed badly"},
			})
		})
		It("TestDeleteDomainForceFlagSkipsConfirmation", func() {

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

			domain := models.DomainFields{}
			domain.Name = "foo.com"
			domain.Guid = "foo-guid"
			domainRepo := &testapi.FakeDomainRepository{
				FindByNameInOrgDomain: domain,
			}
			ui := callDeleteDomain([]string{"-f", "foo.com"}, []string{}, reqFactory, domainRepo)

			assert.Equal(mr.T(), domainRepo.DeleteDomainGuid, "foo-guid")
			assert.Equal(mr.T(), len(ui.Prompts), 0)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting domain", "foo.com"},
				{"OK"},
			})
		})
	})
}

func callDeleteDomain(args []string, inputs []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (ui *testterm.FakeUI) {
	ctxt := testcmd.NewContext("delete-domain", args)
	ui = &testterm.FakeUI{
		Inputs: inputs,
	}

	configRepo := testconfig.NewRepositoryWithAccessToken(configuration.TokenInfo{Username: "my-user"})

	spaceFields := models.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := models.OrganizationFields{}
	orgFields.Name = "my-org"
	configRepo.SetSpaceFields(spaceFields)
	configRepo.SetOrganizationFields(orgFields)

	cmd := domain.NewDeleteDomain(ui, configRepo, domainRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
