package domain_test

import (
	"cf"
	. "cf/commands/domain"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
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
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com"},
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	ui := callDeleteDomain(t, []string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.DeleteDomain.Name, "foo.com")

	assert.Contains(t, ui.Prompts[0], "delete")
	assert.Contains(t, ui.Prompts[0], "foo.com")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteDomainNoConfirmation(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com"},
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	ui := callDeleteDomain(t, []string{"foo.com"}, []string{"no"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.DeleteDomain.Name, "")

	assert.Contains(t, ui.Prompts[0], "delete")
	assert.Contains(t, ui.Prompts[0], "foo.com")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")

	assert.Equal(t, len(ui.Outputs), 1)
}

func TestDeleteDomainNotFound(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgApiResponse: net.NewNotFoundApiResponse("%s %s not found", "Domain", "foo.com"),
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	ui := callDeleteDomain(t, []string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.DeleteDomain.Name, "")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "foo.com")
	assert.Contains(t, ui.Outputs[2], "not found")
}

func TestDeleteDomainFindError(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgApiResponse: net.NewApiResponseWithMessage("failed badly"),
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	ui := callDeleteDomain(t, []string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.DeleteDomain.Name, "")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "foo.com")
	assert.Contains(t, ui.Outputs[2], "failed badly")
}

func TestDeleteDomainDeleteError(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com"},
		DeleteApiResponse:     net.NewApiResponseWithMessage("failed badly"),
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	ui := callDeleteDomain(t, []string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.DeleteDomain.Name, "foo.com")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "foo.com")
	assert.Contains(t, ui.Outputs[2], "failed badly")
}

func TestDeleteDomainDeleteSharedHasSharedConfirmation(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com", Shared: true},
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	ui := callDeleteDomain(t, []string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.DeleteDomain.Name, "foo.com")

	assert.Contains(t, ui.Prompts[0], "shared")
	assert.Contains(t, ui.Prompts[0], "foo.com")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteDomainForceFlagSkipsConfirmation(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com", Shared: true},
	}
	ui := callDeleteDomain(t, []string{"-f", "foo.com"}, []string{}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.DeleteDomain.Name, "foo.com")

	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "OK")
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

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewDeleteDomain(ui, config, domainRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
