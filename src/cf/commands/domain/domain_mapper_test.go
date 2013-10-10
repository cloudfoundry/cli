package domain_test

import (
	"cf"
	. "cf/commands/domain"
	"cf/net"
	"errors"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestMapDomainRequirements(t *testing.T) {
	ui := &testterm.FakeUI{}
	domainRepo := &testapi.FakeDomainRepository{}
	cmd := NewDomainMapper(ui, domainRepo, true)

	ctxt := testcmd.NewContext("map-domain", []string{"foo.com", "my-space"})
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

func TestMapDomainSuccess(t *testing.T) {
	ctxt := testcmd.NewContext("map-domain", []string{"my-space", "foo.com"})
	ui := &testterm.FakeUI{}
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com"},
	}

	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:       true,
		TargetedOrgSuccess: true,
		Organization:       cf.Organization{Name: "my-org", Guid: "my-org-guid"},
		Space:              cf.Space{Name: "my-space"},
	}

	cmd := NewDomainMapper(ui, domainRepo, true)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	assert.Equal(t, domainRepo.MapDomainDomain.Name, "foo.com")
	assert.Equal(t, domainRepo.MapDomainSpace.Name, "my-space")
	assert.Contains(t, ui.Outputs[0], "Mapping domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestMapDomainDomainNotFound(t *testing.T) {
	ctxt := testcmd.NewContext("map-domain", []string{"my-space", "foo.com"})
	ui := &testterm.FakeUI{}
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgApiResponse: net.NewNotFoundApiResponse("%s %s not found", "Domain", "foo.com"),
	}

	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:       true,
		TargetedOrgSuccess: true,
		Organization:       cf.Organization{Name: "my-org", Guid: "my-org-guid"},
		Space:              cf.Space{Name: "my-space"},
	}

	cmd := NewDomainMapper(ui, domainRepo, true)

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[0], "Mapping domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "foo.com")
}

func TestMapDomainMappingFails(t *testing.T) {
	ctxt := testcmd.NewContext("map-domain", []string{"my-space", "foo.com"})
	ui := &testterm.FakeUI{}
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com"},
		MapDomainApiResponse:  net.NewApiResponseWithError("Did not work %s", errors.New("bummer")),
	}

	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:       true,
		TargetedOrgSuccess: true,
		Organization:       cf.Organization{Name: "my-org", Guid: "my-org-guid"},
		Space:              cf.Space{Name: "my-space"},
	}

	cmd := NewDomainMapper(ui, domainRepo, true)

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[0], "Mapping domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Did not work")
	assert.Contains(t, ui.Outputs[2], "bummer")
}

func TestUnmapDomainSuccess(t *testing.T) {
	ctxt := testcmd.NewContext("unmap-domain", []string{"my-space", "foo.com"})
	ui := &testterm.FakeUI{}
	domainRepo := &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com"},
	}

	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:       true,
		TargetedOrgSuccess: true,
		Organization:       cf.Organization{Name: "my-org", Guid: "my-org-guid"},
		Space:              cf.Space{Name: "my-space"},
	}

	cmd := NewDomainMapper(ui, domainRepo, false)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.UnmapDomainDomain.Name, "foo.com")
	assert.Equal(t, domainRepo.UnmapDomainSpace.Name, "my-space")
	assert.Contains(t, ui.Outputs[0], "Unmapping domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[1], "OK")
}
