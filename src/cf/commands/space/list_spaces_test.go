package space_test

import (
	"cf"
	"cf/api"
	. "cf/commands/space"
	"cf/configuration"
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

func callSpaces(args []string, reqFactory *testreq.FakeReqFactory, config *configuration.Configuration, spaceRepo api.SpaceRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("spaces", args)

	cmd := NewListSpaces(ui, config, spaceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSpacesRequirements", func() {
			spaceRepo := &testapi.FakeSpaceRepository{}
			config := &configuration.Configuration{}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
			callSpaces([]string{}, reqFactory, config, spaceRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
			callSpaces([]string{}, reqFactory, config, spaceRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
			callSpaces([]string{}, reqFactory, config, spaceRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestListingSpaces", func() {

			space := cf.Space{}
			space.Name = "space1"
			space2 := cf.Space{}
			space2.Name = "space2"
			space3 := cf.Space{}
			space3.Name = "space3"
			spaceRepo := &testapi.FakeSpaceRepository{
				Spaces: []cf.Space{space, space2, space3},
			}
			token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
				Username: "my-user",
			})

			assert.NoError(mr.T(), err)
			org := cf.OrganizationFields{}
			org.Name = "my-org"
			config := &configuration.Configuration{
				OrganizationFields: org,
				AccessToken:        token,
			}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
			ui := callSpaces([]string{}, reqFactory, config, spaceRepo)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting spaces in org", "my-org", "my-user"},
				{"space1"},
				{"space2"},
				{"space3"},
			})
		})
		It("TestListingSpacesWhenNoSpaces", func() {

			spaceRepo := &testapi.FakeSpaceRepository{
				Spaces: []cf.Space{},
			}
			token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
				Username: "my-user",
			})

			assert.NoError(mr.T(), err)
			org2 := cf.OrganizationFields{}
			org2.Name = "my-org"
			config := &configuration.Configuration{
				OrganizationFields: org2,
				AccessToken:        token,
			}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
			ui := callSpaces([]string{}, reqFactory, config, spaceRepo)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting spaces in org", "my-org", "my-user"},
				{"No spaces found"},
			})
		})
	})
}
