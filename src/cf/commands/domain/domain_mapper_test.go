package domain_test

import (
	"cf"
	"cf/commands/domain"
	"cf/configuration"
	"cf/net"
	"errors"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestMapDomainRequirements(t *testing.T) {
	reqFactory, domainRepo := getDomainMapperDeps()
	callDomainMapper(t, true, []string{"my-space", "foo.com"}, reqFactory, domainRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	reqFactory.TargetedOrgSuccess = false
	callDomainMapper(t, true, []string{"my-space", "foo.com"}, reqFactory, domainRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = false
	reqFactory.TargetedOrgSuccess = true
	callDomainMapper(t, true, []string{"my-space", "foo.com"}, reqFactory, domainRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	reqFactory.TargetedOrgSuccess = true
	callDomainMapper(t, true, []string{}, reqFactory, domainRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestMapDomainSuccess(t *testing.T) {
	reqFactory, domainRepo := getDomainMapperDeps()
	ui := callDomainMapper(t, true, []string{"my-space", "foo.com"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.MapDomainGuid, "foo-guid")
	assert.Equal(t, domainRepo.MapSpaceGuid, "my-space-guid")
	assert.Contains(t, ui.Outputs[0], "Mapping domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestMapDomainDomainNotFound(t *testing.T) {
	reqFactory, domainRepo := getDomainMapperDeps()
	domainRepo.FindByNameInOrgApiResponse = net.NewNotFoundApiResponse("Domain foo.com not found")
	ui := callDomainMapper(t, true, []string{"my-space", "foo.com"}, reqFactory, domainRepo)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[0], "Mapping domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "foo.com")
}

func TestMapDomainMappingFails(t *testing.T) {
	reqFactory, domainRepo := getDomainMapperDeps()
	domainRepo.MapApiResponse = net.NewApiResponseWithError("Did not work %s", errors.New("bummer"))

	ui := callDomainMapper(t, true, []string{"my-space", "foo.com"}, reqFactory, domainRepo)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[0], "Mapping domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Did not work")
	assert.Contains(t, ui.Outputs[2], "bummer")
}

func TestUnmapDomainSuccess(t *testing.T) {
	reqFactory, domainRepo := getDomainMapperDeps()
	ui := callDomainMapper(t, false, []string{"my-space", "foo.com"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.UnmapDomainGuid, "foo-guid")
	assert.Equal(t, domainRepo.UnmapSpaceGuid, "my-space-guid")
	assert.Contains(t, ui.Outputs[0], "Unmapping domain")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func getDomainMapperDeps() (reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) {
	domain := cf.Domain{}
	domain.Name = "foo.com"
	domain.Guid = "foo-guid"
	domainRepo = &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: domain,
	}

	org := cf.Organization{}
	org.Name = "my-org"
	org.Guid = "my-org-guid"

	space := cf.Space{}
	space.Name = "my-space"
	space.Guid = "my-space-guid"

	reqFactory = &testreq.FakeReqFactory{
		LoginSuccess:       true,
		TargetedOrgSuccess: true,
		Organization:       org,
		Space:              space,
	}
	return
}

func callDomainMapper(t *testing.T, shouldMap bool, args []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (ui *testterm.FakeUI) {
	cmdName := "map-domain"
	if !shouldMap {
		cmdName = "unmap-domain"
	}

	ctxt := testcmd.NewContext(cmdName, args)
	ui = &testterm.FakeUI{}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	orgFields := cf.OrganizationFields{}
	orgFields.Name = "my-org"

	spaceFields := cf.SpaceFields{}
	spaceFields.Name = "my-space"

	config := &configuration.Configuration{
		SpaceFields:        spaceFields,
		OrganizationFields: orgFields,
		AccessToken:        token,
	}

	cmd := domain.NewDomainMapper(ui, config, domainRepo, shouldMap)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
