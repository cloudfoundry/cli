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

func TestGetRequirements(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	callDeleteDomain(t, []string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	callDeleteDomain(t, []string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
	callDeleteDomain(t, []string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestDeleteDomainSuccess(t *testing.T) {
	domain := cf.Domain{}
	domain.Name = "foo.com"
	domain.Guid = "foo-guid"
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: domain,
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	ui := callDeleteDomain(t, []string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.DeleteDomainGuid, "foo-guid")

	testassert.SliceContains(t, ui.Prompts, testassert.Lines{
		{"delete", "foo.com"},
	})

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting domain", "foo.com", "my-user"},
		{"OK"},
	})
}

func TestDeleteDomainNoConfirmation(t *testing.T) {
	domain := cf.Domain{}
	domain.Name = "foo.com"
	domain.Guid = "foo-guid"
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: domain,
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	ui := callDeleteDomain(t, []string{"foo.com"}, []string{"no"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.DeleteDomainGuid, "")

	testassert.SliceContains(t, ui.Prompts, testassert.Lines{
		{"delete", "foo.com"},
	})

	assert.Equal(t, len(ui.Outputs), 1)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting domain", "foo.com"},
	})
}

func TestDeleteDomainNotFound(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgApiResponse: net.NewNotFoundApiResponse("%s %s not found", "Domain", "foo.com"),
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	ui := callDeleteDomain(t, []string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.DeleteDomainGuid, "")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting domain", "foo.com"},
		{"OK"},
		{"foo.com", "not found"},
	})
}

func TestDeleteDomainFindError(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgApiResponse: net.NewApiResponseWithMessage("failed badly"),
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	ui := callDeleteDomain(t, []string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.DeleteDomainGuid, "")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting domain", "foo.com"},
		{"FAILED"},
		{"foo.com"},
		{"failed badly"},
	})
}

func TestDeleteDomainDeleteError(t *testing.T) {
	domain := cf.Domain{}
	domain.Name = "foo.com"
	domain.Guid = "foo-guid"
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: domain,
		DeleteApiResponse:     net.NewApiResponseWithMessage("failed badly"),
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	ui := callDeleteDomain(t, []string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.DeleteDomainGuid, "foo-guid")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting domain", "foo.com"},
		{"FAILED"},
		{"foo.com"},
		{"failed badly"},
	})
}

func TestDeleteDomainForceFlagSkipsConfirmation(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	domain := cf.Domain{}
	domain.Name = "foo.com"
	domain.Guid = "foo-guid"
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: domain,
	}
	ui := callDeleteDomain(t, []string{"-f", "foo.com"}, []string{}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.DeleteDomainGuid, "foo-guid")
	assert.Equal(t, len(ui.Prompts), 0)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting domain", "foo.com"},
		{"OK"},
	})
}

func callDeleteDomain(t *testing.T, args []string, inputs []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (ui *testterm.FakeUI) {
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

	cmd := domain.NewDeleteDomain(ui, config, domainRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
