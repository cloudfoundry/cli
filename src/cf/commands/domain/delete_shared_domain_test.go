package domain_test

import (
	"cf"
	"cf/commands/domain"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestGetDeleteSharedDomainRequirements(t *testing.T) {
	deps := getDeleteSharedDomainDeps()
	deps.requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	callDeleteSharedDomain(t, []string{"foo.com"}, []string{"y"}, deps)
	assert.True(t, testcmd.CommandDidPassRequirements)

	deps.requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	callDeleteSharedDomain(t, []string{"foo.com"}, []string{"y"}, deps)
	assert.False(t, testcmd.CommandDidPassRequirements)

	deps.requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
	callDeleteSharedDomain(t, []string{"foo.com"}, []string{"y"}, deps)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestDeleteSharedDomainNotFound(t *testing.T) {
	deps := getDeleteSharedDomainDeps()
	deps.domainRepo.FindByNameInOrgApiResponse = net.NewNotFoundApiResponse("%s %s not found", "Domain", "foo.com")
	ui := callDeleteSharedDomain(t, []string{"foo.com"}, []string{"y"}, deps)

	assert.Equal(t, deps.domainRepo.DeleteDomainGuid, "")
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting domain", "foo.com"},
		{"OK"},
		{"foo.com", "not found"},
	})
}

func TestDeleteSharedDomainFindError(t *testing.T) {
	deps := getDeleteSharedDomainDeps()
	deps.domainRepo.FindByNameInOrgApiResponse = net.NewApiResponseWithMessage("couldn't find the droids you're lookin for")
	ui := callDeleteSharedDomain(t, []string{"foo.com"}, []string{"y"}, deps)

	assert.Equal(t, deps.domainRepo.DeleteDomainGuid, "")
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting domain", "foo.com"},
		{"FAILED"},
		{"foo.com"},
		{"couldn't find the droids you're lookin for"},
	})
}

func TestDeleteSharedDomainDeleteError(t *testing.T) {
	deps := getDeleteSharedDomainDeps()
	deps.domainRepo.DeleteSharedApiResponse = net.NewApiResponseWithMessage("failed badly")
	ui := callDeleteSharedDomain(t, []string{"foo.com"}, []string{"y"}, deps)

	assert.Equal(t, deps.domainRepo.DeleteSharedDomainGuid, "foo-guid")
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting domain", "foo.com"},
		{"FAILED"},
		{"foo.com"},
		{"failed badly"},
	})
}

func TestDeleteSharedDomainHasConfirmation(t *testing.T) {
	deps := getDeleteSharedDomainDeps()
	ui := callDeleteSharedDomain(t, []string{"foo.com"}, []string{"y"}, deps)

	assert.Equal(t, deps.domainRepo.DeleteSharedDomainGuid, "foo-guid")
	testassert.SliceContains(t, ui.Prompts, testassert.Lines{
		{"shared", "foo.com"},
	})
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting domain", "foo.com"},
		{"OK"},
	})
}

func TestDeleteSharedDomainForceFlagSkipsConfirmation(t *testing.T) {
	deps := getDeleteSharedDomainDeps()
	ui := callDeleteSharedDomain(t, []string{"-f", "foo.com"}, []string{}, deps)

	assert.Equal(t, deps.domainRepo.DeleteSharedDomainGuid, "foo-guid")
	assert.Equal(t, len(ui.Prompts), 0)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting domain", "foo.com"},
		{"OK"},
	})
}

func fakeDomainRepo() *testapi.FakeDomainRepository {
	domain := cf.Domain{}
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

func callDeleteSharedDomain(t *testing.T, args []string, inputs []string, deps deleteSharedDomainDependencies) (ui *testterm.FakeUI) {
	ctxt := testcmd.NewContext("delete-domain", args)
	ui = &testterm.FakeUI{
		Inputs: inputs,
	}

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

	cmd := domain.NewDeleteSharedDomain(ui, config, deps.domainRepo)
	testcmd.RunCommand(cmd, ctxt, deps.requirementsFactory)
	return
}
