package commands_test

import (
	. "cf/commands"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

type passwordDeps struct {
	ReqFactory *testreq.FakeReqFactory
	PwdRepo    *testapi.FakePasswordRepo
	Config     *configuration.Configuration
}

func getPasswordDeps() passwordDeps {
	return passwordDeps{
		ReqFactory: &testreq.FakeReqFactory{ValidAccessTokenSuccess: true},
		PwdRepo:    &testapi.FakePasswordRepo{UpdateUnauthorized: true},
		Config:     &configuration.Configuration{},
	}
}

func callPassword(inputs []string, deps passwordDeps) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{Inputs: inputs}

	ctxt := testcmd.NewContext("passwd", []string{})
	cmd := NewPassword(ui, deps.PwdRepo, deps.Config)
	testcmd.RunCommand(cmd, ctxt, deps.ReqFactory)

	return
}

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestPasswordRequiresValidAccessToken", func() {
			deps := getPasswordDeps()
			deps.ReqFactory.ValidAccessTokenSuccess = false
			callPassword([]string{}, deps)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			deps.ReqFactory.ValidAccessTokenSuccess = true
			deps.PwdRepo.UpdateUnauthorized = false

			callPassword([]string{"", "", ""}, deps)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
		})

		It("TestPasswordCanBeChanged", func() {
			deps := getPasswordDeps()
			deps.ReqFactory.ValidAccessTokenSuccess = true
			deps.PwdRepo.UpdateUnauthorized = false
			ui := callPassword([]string{"old-password", "new-password", "new-password"}, deps)

			testassert.SliceContains(mr.T(), ui.PasswordPrompts, testassert.Lines{
				{"Current Password"},
				{"New Password"},
				{"Verify Password"},
			})

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Changing password..."},
				{"OK"},
				{"Please log in again"},
			})

			assert.Equal(mr.T(), deps.PwdRepo.UpdateNewPassword, "new-password")
			assert.Equal(mr.T(), deps.PwdRepo.UpdateOldPassword, "old-password")

			assert.Empty(mr.T(), deps.Config.AccessToken)
			assert.Equal(mr.T(), deps.Config.OrganizationFields, models.OrganizationFields{})
			assert.Equal(mr.T(), deps.Config.SpaceFields, models.SpaceFields{})
		})

		It("TestPasswordVerification", func() {
			deps := getPasswordDeps()
			deps.ReqFactory.ValidAccessTokenSuccess = true
			ui := callPassword([]string{"old-password", "new-password", "new-password-with-error"}, deps)

			testassert.SliceContains(mr.T(), ui.PasswordPrompts, testassert.Lines{
				{"Current Password"},
				{"New Password"},
				{"Verify Password"},
			})
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"Password verification does not match"},
			})

			assert.Equal(mr.T(), deps.PwdRepo.UpdateNewPassword, "")
		})

		It("TestWhenCurrentPasswordDoesNotMatch", func() {
			deps := getPasswordDeps()
			deps.ReqFactory.ValidAccessTokenSuccess = true
			deps.PwdRepo.UpdateUnauthorized = true
			ui := callPassword([]string{"old-password", "new-password", "new-password"}, deps)

			testassert.SliceContains(mr.T(), ui.PasswordPrompts, testassert.Lines{
				{"Current Password"},
				{"New Password"},
				{"Verify Password"},
			})

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Changing password..."},
				{"FAILED"},
				{"Current password did not match"},
			})

			assert.Equal(mr.T(), deps.PwdRepo.UpdateNewPassword, "new-password")
			assert.Equal(mr.T(), deps.PwdRepo.UpdateOldPassword, "old-password")
		})
	})
}
