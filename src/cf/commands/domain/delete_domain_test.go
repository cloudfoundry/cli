package domain_test

import (
	"cf"
	. "cf/commands/domain"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestGetRequirements(t *testing.T) {
	ui := &testhelpers.FakeUI{Inputs: []string{"y"}}
	domainRepo := &testhelpers.FakeDomainRepository{}
	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testhelpers.NewContext("delete-domain", []string{"foo.com"})
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.False(t, testhelpers.CommandDidPassRequirements)

	ctxt = testhelpers.NewContext("map-domain", []string{})
	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestDeleteDomainSuccess(t *testing.T) {
	ui := &testhelpers.FakeUI{Inputs: []string{"y"}}
	domainRepo := &testhelpers.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com"},
	}

	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testhelpers.NewContext("delete-domain", []string{"foo.com"})
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.DeleteDomainDomain.Name, "foo.com")

	assert.Contains(t, ui.Prompts[0], "delete")
	assert.Contains(t, ui.Prompts[0], "foo.com")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteDomainNoConfirmation(t *testing.T) {
	ui := &testhelpers.FakeUI{Inputs: []string{"no"}}
	domainRepo := &testhelpers.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com"},
	}

	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testhelpers.NewContext("delete-domain", []string{"foo.com"})
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.DeleteDomainDomain.Name, "")

	assert.Contains(t, ui.Prompts[0], "delete")
	assert.Contains(t, ui.Prompts[0], "foo.com")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")

	assert.Equal(t, len(ui.Outputs), 1)
}

func TestDeleteDomainNotFound(t *testing.T) {
	ui := &testhelpers.FakeUI{Inputs: []string{"y"}}
	domainRepo := &testhelpers.FakeDomainRepository{
		FindByNameInOrgApiResponse: net.NewNotFoundApiResponse("%s %s not found", "Domain", "foo.com"),
	}

	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testhelpers.NewContext("delete-domain", []string{"foo.com"})
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.DeleteDomainDomain.Name, "")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "foo.com")
	assert.Contains(t, ui.Outputs[2], "not found")
}

func TestDeleteDomainFindError(t *testing.T) {
	ui := &testhelpers.FakeUI{Inputs: []string{"y"}}
	domainRepo := &testhelpers.FakeDomainRepository{
		FindByNameInOrgApiResponse: net.NewApiResponseWithMessage("failed badly"),
	}

	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testhelpers.NewContext("delete-domain", []string{"foo.com"})
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.DeleteDomainDomain.Name, "")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "foo.com")
	assert.Contains(t, ui.Outputs[2], "failed badly")
}

func TestDeleteDomainDeleteError(t *testing.T) {
	ui := &testhelpers.FakeUI{Inputs: []string{"y"}}
	domainRepo := &testhelpers.FakeDomainRepository{
		FindByNameInOrgDomain:   cf.Domain{Name: "foo.com"},
		DeleteDomainApiResponse: net.NewApiResponseWithMessage("failed badly"),
	}

	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testhelpers.NewContext("delete-domain", []string{"foo.com"})
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.DeleteDomainDomain.Name, "foo.com")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "foo.com")
	assert.Contains(t, ui.Outputs[2], "failed badly")
}

func TestDeleteDomainDeleteSharedHasSharedConfirmation(t *testing.T) {
	ui := &testhelpers.FakeUI{Inputs: []string{"y"}}
	domainRepo := &testhelpers.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com", Shared: true},
	}

	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testhelpers.NewContext("delete-domain", []string{"foo.com"})
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.DeleteDomainDomain.Name, "foo.com")

	assert.Contains(t, ui.Prompts[0], "shared")
	assert.Contains(t, ui.Prompts[0], "foo.com")

	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteDomainForceFlagSkipsConfirmation(t *testing.T) {
	ui := &testhelpers.FakeUI{}
	domainRepo := &testhelpers.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com", Shared: true},
	}

	cmd := NewDeleteDomain(ui, domainRepo)

	ctxt := testhelpers.NewContext("delete-domain", []string{"-f", "foo.com"})
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.DeleteDomainDomain.Name, "foo.com")

	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "OK")
}
