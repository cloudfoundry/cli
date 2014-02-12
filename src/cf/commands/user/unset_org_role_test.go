package user_test

import (
	. "cf/commands/user"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestUnsetOrgRoleFailsWithUsage", func() {
		userRepo := &testapi.FakeUserRepository{}
		reqFactory := &testreq.FakeReqFactory{}

		ui := callUnsetOrgRole(mr.T(), []string{}, userRepo, reqFactory)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callUnsetOrgRole(mr.T(), []string{"username"}, userRepo, reqFactory)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callUnsetOrgRole(mr.T(), []string{"username", "org"}, userRepo, reqFactory)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callUnsetOrgRole(mr.T(), []string{"username", "org", "role"}, userRepo, reqFactory)
		assert.False(mr.T(), ui.FailedWithUsage)
	})
	It("TestUnsetOrgRoleRequirements", func() {

		userRepo := &testapi.FakeUserRepository{}
		reqFactory := &testreq.FakeReqFactory{}
		args := []string{"username", "org", "role"}

		reqFactory.LoginSuccess = false
		callUnsetOrgRole(mr.T(), args, userRepo, reqFactory)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory.LoginSuccess = true
		callUnsetOrgRole(mr.T(), args, userRepo, reqFactory)
		assert.True(mr.T(), testcmd.CommandDidPassRequirements)

		Expect(reqFactory.UserUsername).To(Equal("username"))
		assert.Equal(mr.T(), reqFactory.OrganizationName, "org")
	})
	It("TestUnsetOrgRole", func() {

		userRepo := &testapi.FakeUserRepository{}
		user := models.UserFields{}
		user.Username = "some-user"
		user.Guid = "some-user-guid"
		org := models.Organization{}
		org.Name = "some-org"
		org.Guid = "some-org-guid"
		reqFactory := &testreq.FakeReqFactory{
			LoginSuccess: true,
			UserFields:   user,
			Organization: org,
		}
		args := []string{"my-username", "my-org", "OrgManager"}

		ui := callUnsetOrgRole(mr.T(), args, userRepo, reqFactory)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Removing role", "OrgManager", "my-username", "my-org", "my-user"},
			{"OK"},
		})

		assert.Equal(mr.T(), userRepo.UnsetOrgRoleRole, models.ORG_MANAGER)
		assert.Equal(mr.T(), userRepo.UnsetOrgRoleUserGuid, "some-user-guid")
		assert.Equal(mr.T(), userRepo.UnsetOrgRoleOrganizationGuid, "some-org-guid")
	})
})

func callUnsetOrgRole(t mr.TestingT, args []string, userRepo *testapi.FakeUserRepository, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("unset-org-role", args)

	configRepo := testconfig.NewRepositoryWithDefaults()

	cmd := NewUnsetOrgRole(ui, configRepo, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
