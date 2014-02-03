package user_test

import (
	"cf"
	. "cf/commands/user"
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

func callOrgUsers(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org3 := cf.OrganizationFields{}
	org3.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org3,
		AccessToken:        token,
	}

	cmd := NewOrgUsers(ui, config, userRepo)
	ctxt := testcmd.NewContext("org-users", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestOrgUsersFailsWithUsage", func() {
			reqFactory := &testreq.FakeReqFactory{}
			userRepo := &testapi.FakeUserRepository{}
			ui := callOrgUsers(mr.T(), []string{}, reqFactory, userRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callOrgUsers(mr.T(), []string{"Org1"}, reqFactory, userRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestOrgUsersRequirements", func() {

			reqFactory := &testreq.FakeReqFactory{}
			userRepo := &testapi.FakeUserRepository{}
			args := []string{"Org1"}

			reqFactory.LoginSuccess = false
			callOrgUsers(mr.T(), args, reqFactory, userRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory.LoginSuccess = true
			callOrgUsers(mr.T(), args, reqFactory, userRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			assert.Equal(mr.T(), "Org1", reqFactory.OrganizationName)
		})
		It("TestOrgUsers", func() {

			org := cf.Organization{}
			org.Name = "Found Org"
			org.Guid = "found-org-guid"

			userRepo := &testapi.FakeUserRepository{}
			user := cf.UserFields{}
			user.Username = "user1"
			user2 := cf.UserFields{}
			user2.Username = "user2"
			user3 := cf.UserFields{}
			user3.Username = "user3"
			user4 := cf.UserFields{}
			user4.Username = "user4"
			userRepo.ListUsersByRole = map[string][]cf.UserFields{
				cf.ORG_MANAGER:     []cf.UserFields{user, user2},
				cf.BILLING_MANAGER: []cf.UserFields{user4},
				cf.ORG_AUDITOR:     []cf.UserFields{user3},
			}

			reqFactory := &testreq.FakeReqFactory{
				LoginSuccess: true,
				Organization: org,
			}

			ui := callOrgUsers(mr.T(), []string{"Org1"}, reqFactory, userRepo)

			assert.Equal(mr.T(), userRepo.ListUsersOrganizationGuid, "found-org-guid")
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting users in org", "Found Org", "my-user"},
				{"ORG MANAGER"},
				{"user1"},
				{"user2"},
				{"BILLING MANAGER"},
				{"user4"},
				{"ORG AUDITOR"},
				{"user3"},
			})
		})
	})
}
