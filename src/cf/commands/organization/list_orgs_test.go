package organization_test

import (
	"cf/commands/organization"
	"cf/configuration"
	"cf/models"
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

func callListOrgs(config configuration.Reader, reqFactory *testreq.FakeReqFactory, orgRepo *testapi.FakeOrgRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("orgs", []string{})
	cmd := organization.NewListOrgs(fakeUI, config, orgRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {

	It("TestListOrgsRequirements", func() {
		orgRepo := &testapi.FakeOrgRepository{}
		config := testconfig.NewRepository()

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		callListOrgs(config, reqFactory, orgRepo)
		assert.True(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
		callListOrgs(config, reqFactory, orgRepo)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)
	})

	It("TestListAllPagesOfOrgs", func() {
		org1 := models.Organization{}
		org1.Name = "Organization-1"

		org2 := models.Organization{}
		org2.Name = "Organization-2"

		org3 := models.Organization{}
		org3.Name = "Organization-3"

		orgRepo := &testapi.FakeOrgRepository{
			Organizations: []models.Organization{org1, org2, org3},
		}

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		tokenInfo := configuration.TokenInfo{Username: "my-user"}
		config := testconfig.NewRepositoryWithAccessToken(tokenInfo)

		ui := callListOrgs(config, reqFactory, orgRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Getting orgs as my-user"},
			{"Organization-1"},
			{"Organization-2"},
			{"Organization-3"},
		})
	})

	It("TestListNoOrgs", func() {
		orgs := []models.Organization{}
		orgRepo := &testapi.FakeOrgRepository{
			Organizations: orgs,
		}

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		tokenInfo := configuration.TokenInfo{Username: "my-user"}
		config := testconfig.NewRepositoryWithAccessToken(tokenInfo)

		ui := callListOrgs(config, reqFactory, orgRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Getting orgs as my-user"},
			{"No orgs found"},
		})
	})
})
