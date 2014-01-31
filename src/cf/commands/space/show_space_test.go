package space_test

import (
	"cf"
	. "cf/commands/space"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callShowSpace(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("space", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		AccessToken:        token,
		OrganizationFields: org,
	}

	cmd := NewShowSpace(ui, config)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestShowSpaceRequirements", func() {
			args := []string{"my-space"}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
			callShowSpace(mr.T(), args, reqFactory)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
			callShowSpace(mr.T(), args, reqFactory)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
			callShowSpace(mr.T(), args, reqFactory)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestShowSpaceInfoSuccess", func() {

			org := cf.OrganizationFields{}
			org.Name = "my-org"

			app := cf.ApplicationFields{}
			app.Name = "app1"
			app.Guid = "app1-guid"
			apps := []cf.ApplicationFields{app}

			domain := cf.DomainFields{}
			domain.Name = "domain1"
			domain.Guid = "domain1-guid"
			domains := []cf.DomainFields{domain}

			serviceInstance := cf.ServiceInstanceFields{}
			serviceInstance.Name = "service1"
			serviceInstance.Guid = "service1-guid"
			services := []cf.ServiceInstanceFields{serviceInstance}

			space := cf.Space{}
			space.Name = "space1"
			space.Organization = org
			space.Applications = apps
			space.Domains = domains
			space.ServiceInstances = services

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, Space: space}
			ui := callShowSpace(mr.T(), []string{"space1"}, reqFactory)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting info for space", "space1", "my-org", "my-user"},
				{"OK"},
				{"space1"},
				{"Org", "my-org"},
				{"Apps", "app1"},
				{"Domains", "domain1"},
				{"Services", "service1"},
			})
		})
	})
}
