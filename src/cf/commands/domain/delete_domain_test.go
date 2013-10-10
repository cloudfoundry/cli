package domain_test

import (
	"cf"
	. "cf/commands/domain"
	"cf/net"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestGetRequirements(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"y"}}
	domainRepo := &testapi.FakeDomainRepository{}
	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testcmd.NewContext("delete-domain", []string{"foo.com"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.False(t, testcmd.CommandDidPassRequirements)

	ctxt = testcmd.NewContext("map-domain", []string{})
	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestDeleteDomainSuccess(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"y"}}
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com"},
	}

	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testcmd.NewContext("delete-domain", []string{"foo.com"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.DeleteDomainDomain.Name, "foo.com")

	assert.Contains(t, ui.Prompts[0], "delete")
	assert.Contains(t, ui.Prompts[0], "foo.com")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteDomainNoConfirmation(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"no"}}
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com"},
	}

	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testcmd.NewContext("delete-domain", []string{"foo.com"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.DeleteDomainDomain.Name, "")

	assert.Contains(t, ui.Prompts[0], "delete")
	assert.Contains(t, ui.Prompts[0], "foo.com")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")

	assert.Equal(t, len(ui.Outputs), 1)
}

func TestDeleteDomainNotFound(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"y"}}
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgApiResponse: net.NewNotFoundApiResponse("%s %s not found", "Domain", "foo.com"),
	}

	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testcmd.NewContext("delete-domain", []string{"foo.com"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.DeleteDomainDomain.Name, "")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "foo.com")
	assert.Contains(t, ui.Outputs[2], "not found")
}

func TestDeleteDomainFindError(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"y"}}
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgApiResponse: net.NewApiResponseWithMessage("failed badly"),
	}

	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testcmd.NewContext("delete-domain", []string{"foo.com"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.DeleteDomainDomain.Name, "")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "foo.com")
	assert.Contains(t, ui.Outputs[2], "failed badly")
}

func TestDeleteDomainDeleteError(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"y"}}
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain:   cf.Domain{Name: "foo.com"},
		DeleteDomainApiResponse: net.NewApiResponseWithMessage("failed badly"),
	}

	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testcmd.NewContext("delete-domain", []string{"foo.com"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.DeleteDomainDomain.Name, "foo.com")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "foo.com")
	assert.Contains(t, ui.Outputs[2], "failed badly")
}

func TestDeleteDomainDeleteSharedHasSharedConfirmation(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"y"}}
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com", Shared: true},
	}

	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testcmd.NewContext("delete-domain", []string{"foo.com"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.DeleteDomainDomain.Name, "foo.com")

	assert.Contains(t, ui.Prompts[0], "shared")
	assert.Contains(t, ui.Prompts[0], "foo.com")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteDomainForceFlagSkipsConfirmation(t *testing.T) {
	ui := &testterm.FakeUI{}
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com", Shared: true},
	}

	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testcmd.NewContext("delete-domain", []string{"-f", "foo.com"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.DeleteDomainDomain.Name, "foo.com")

	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "OK")
}
