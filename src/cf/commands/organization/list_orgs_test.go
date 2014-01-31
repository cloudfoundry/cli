package organization_test

import (
	"cf"
	"cf/commands/organization"
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

func callListOrgs(config *configuration.Configuration, reqFactory *testreq.FakeReqFactory, orgRepo *testapi.FakeOrgRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("orgs", []string{})
	cmd := organization.NewListOrgs(fakeUI, config, orgRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestListOrgsRequirements", func() {
			orgRepo := &testapi.FakeOrgRepository{}
			config := &configuration.Configuration{}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			callListOrgs(config, reqFactory, orgRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
			callListOrgs(config, reqFactory, orgRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestListAllPagesOfOrgs", func() {

			org1 := cf.Organization{}
			org1.Name = "Organization-1"

			org2 := cf.Organization{}
			org2.Name = "Organization-2"

			org3 := cf.Organization{}
			org3.Name = "Organization-3"

			orgRepo := &testapi.FakeOrgRepository{
				Organizations: []cf.Organization{org1, org2, org3},
			}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

			tokenInfo := configuration.TokenInfo{Username: "my-user"}
			accessToken, err := testconfig.CreateAccessTokenWithTokenInfo(tokenInfo)
			assert.NoError(mr.T(), err)
			config := &configuration.Configuration{AccessToken: accessToken}

			ui := callListOrgs(config, reqFactory, orgRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting orgs as my-user"},
				{"Organization-1"},
				{"Organization-2"},
				{"Organization-3"},
			})
		})
		It("TestListNoOrgs", func() {

			orgs := []cf.Organization{}
			orgRepo := &testapi.FakeOrgRepository{
				Organizations: orgs,
			}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

			tokenInfo := configuration.TokenInfo{Username: "my-user"}
			accessToken, err := testconfig.CreateAccessTokenWithTokenInfo(tokenInfo)
			assert.NoError(mr.T(), err)
			config := &configuration.Configuration{AccessToken: accessToken}

			ui := callListOrgs(config, reqFactory, orgRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting orgs as my-user"},
				{"No orgs found"},
			})
		})
	})
}
