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

func fakeDomainRepo() *testapi.FakeDomainRepository {
	domain := models.Domain{}
	domain.Name = "foo.com"
	domain.Guid = "foo-guid"
	domain.Shared = true

	return &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: domain,
	}
}

type deleteSharedDomainDependencies struct {
	requirementsFactory *testreq.FakeReqFactory
	domainRepo          *testapi.FakeDomainRepository
}

func getDeleteSharedDomainDeps() deleteSharedDomainDependencies {
	return deleteSharedDomainDependencies{
		requirementsFactory: &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true},
		domainRepo:          fakeDomainRepo(),
	}
}

func callDeleteSharedDomain(t mr.TestingT, args []string, inputs []string, deps deleteSharedDomainDependencies) (ui *testterm.FakeUI) {
	ctxt := testcmd.NewContext("delete-domain", args)
	ui = &testterm.FakeUI{
		Inputs: inputs,
	}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	spaceFields := models.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := models.OrganizationFields{}
	orgFields.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        spaceFields,
		OrganizationFields: orgFields,
		AccessToken:        token,
	}

	cmd := domain.NewDeleteSharedDomain(ui, config, deps.domainRepo)
	testcmd.RunCommand(cmd, ctxt, deps.requirementsFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestGetDeleteSharedDomainRequirements", func() {
			deps := getDeleteSharedDomainDeps()
			deps.requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

			callDeleteSharedDomain(mr.T(), []string{"foo.com"}, []string{"y"}, deps)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			deps.requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
			callDeleteSharedDomain(mr.T(), []string{"foo.com"}, []string{"y"}, deps)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			deps.requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
			callDeleteSharedDomain(mr.T(), []string{"foo.com"}, []string{"y"}, deps)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestDeleteSharedDomainNotFound", func() {

			deps := getDeleteSharedDomainDeps()
			deps.domainRepo.FindByNameInOrgApiResponse = net.NewNotFoundApiResponse("%s %s not found", "Domain", "foo.com")
			ui := callDeleteSharedDomain(mr.T(), []string{"foo.com"}, []string{"y"}, deps)

			assert.Equal(mr.T(), deps.domainRepo.DeleteDomainGuid, "")
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting domain", "foo.com"},
				{"OK"},
				{"foo.com", "not found"},
			})
		})
		It("TestDeleteSharedDomainFindError", func() {

			deps := getDeleteSharedDomainDeps()
			deps.domainRepo.FindByNameInOrgApiResponse = net.NewApiResponseWithMessage("couldn't find the droids you're lookin for")
			ui := callDeleteSharedDomain(mr.T(), []string{"foo.com"}, []string{"y"}, deps)

			assert.Equal(mr.T(), deps.domainRepo.DeleteDomainGuid, "")
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting domain", "foo.com"},
				{"FAILED"},
				{"foo.com"},
				{"couldn't find the droids you're lookin for"},
			})
		})
		It("TestDeleteSharedDomainDeleteError", func() {

			deps := getDeleteSharedDomainDeps()
			deps.domainRepo.DeleteSharedApiResponse = net.NewApiResponseWithMessage("failed badly")
			ui := callDeleteSharedDomain(mr.T(), []string{"foo.com"}, []string{"y"}, deps)

			assert.Equal(mr.T(), deps.domainRepo.DeleteSharedDomainGuid, "foo-guid")
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting domain", "foo.com"},
				{"FAILED"},
				{"foo.com"},
				{"failed badly"},
			})
		})
		It("TestDeleteSharedDomainHasConfirmation", func() {

			deps := getDeleteSharedDomainDeps()
			ui := callDeleteSharedDomain(mr.T(), []string{"foo.com"}, []string{"y"}, deps)

			assert.Equal(mr.T(), deps.domainRepo.DeleteSharedDomainGuid, "foo-guid")
			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
				{"shared", "foo.com"},
			})
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting domain", "foo.com"},
				{"OK"},
			})
		})
		It("TestDeleteSharedDomainForceFlagSkipsConfirmation", func() {

			deps := getDeleteSharedDomainDeps()
			ui := callDeleteSharedDomain(mr.T(), []string{"-f", "foo.com"}, []string{}, deps)

			assert.Equal(mr.T(), deps.domainRepo.DeleteSharedDomainGuid, "foo-guid")
			assert.Equal(mr.T(), len(ui.Prompts), 0)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting domain", "foo.com"},
				{"OK"},
			})
		})
	})
}
