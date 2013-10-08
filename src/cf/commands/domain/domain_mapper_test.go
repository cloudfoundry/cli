package domain_test

import (
	"cf"
	. "cf/commands/domain"
	"cf/net"
	"errors"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestMapDomainRequirements(t *testing.T) {
	ui := &testhelpers.FakeUI{}
	domainRepo := &testhelpers.FakeDomainRepository{}
	cmd := NewDomainMapper(ui, domainRepo, true)

	ctxt := testhelpers.NewContext("map-domain", []string{"foo.com", "my-space"})
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

func TestMapDomainSuccess(t *testing.T) {
	ctxt := testhelpers.NewContext("map-domain", []string{"my-space", "foo.com"})
	ui := &testhelpers.FakeUI{}
	domainRepo := &testhelpers.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com"},
	}

	reqFactory := &testhelpers.FakeReqFactory{
		LoginSuccess:       true,
		TargetedOrgSuccess: true,
		Organization:       cf.Organization{Name: "my-org", Guid: "my-org-guid"},
		Space:              cf.Space{Name: "my-space"},
	}

	cmd := NewDomainMapper(ui, domainRepo, true)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	assert.Equal(t, domainRepo.MapDomainDomain.Name, "foo.com")
	assert.Equal(t, domainRepo.MapDomainSpace.Name, "my-space")
	assert.Contains(t, ui.Outputs[0], "Mapping domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestMapDomainDomainNotFound(t *testing.T) {
	ctxt := testhelpers.NewContext("map-domain", []string{"my-space", "foo.com"})
	ui := &testhelpers.FakeUI{}
	domainRepo := &testhelpers.FakeDomainRepository{
		FindByNameInOrgApiResponse: net.NewNotFoundApiResponse("Domain", "foo.com"),
	}

	reqFactory := &testhelpers.FakeReqFactory{
		LoginSuccess:       true,
		TargetedOrgSuccess: true,
		Organization:       cf.Organization{Name: "my-org", Guid: "my-org-guid"},
		Space:              cf.Space{Name: "my-space"},
	}

	cmd := NewDomainMapper(ui, domainRepo, true)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[0], "Mapping domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "foo.com")
}

func TestMapDomainMappingFails(t *testing.T) {
	ctxt := testhelpers.NewContext("map-domain", []string{"my-space", "foo.com"})
	ui := &testhelpers.FakeUI{}
	domainRepo := &testhelpers.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com"},
		MapDomainApiResponse:  net.NewApiResponseWithError("Did not work %s", errors.New("bummer")),
	}

	reqFactory := &testhelpers.FakeReqFactory{
		LoginSuccess:       true,
		TargetedOrgSuccess: true,
		Organization:       cf.Organization{Name: "my-org", Guid: "my-org-guid"},
		Space:              cf.Space{Name: "my-space"},
	}

	cmd := NewDomainMapper(ui, domainRepo, true)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[0], "Mapping domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Did not work")
	assert.Contains(t, ui.Outputs[2], "bummer")
}

func TestUnmapDomainSuccess(t *testing.T) {
	ctxt := testhelpers.NewContext("unmap-domain", []string{"my-space", "foo.com"})
	ui := &testhelpers.FakeUI{}
	domainRepo := &testhelpers.FakeDomainRepository{
		FindByNameInOrgDomain: cf.Domain{Name: "foo.com"},
	}

	reqFactory := &testhelpers.FakeReqFactory{
		LoginSuccess:       true,
		TargetedOrgSuccess: true,
		Organization:       cf.Organization{Name: "my-org", Guid: "my-org-guid"},
		Space:              cf.Space{Name: "my-space"},
	}

	cmd := NewDomainMapper(ui, domainRepo, false)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, domainRepo.UnmapDomainDomain.Name, "foo.com")
	assert.Equal(t, domainRepo.UnmapDomainSpace.Name, "my-space")
	assert.Contains(t, ui.Outputs[0], "Unmapping domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[1], "OK")
}
