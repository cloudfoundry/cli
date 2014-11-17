package commands_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("password command", func() {
	var deps passwordDeps

	BeforeEach(func() {
		deps = getPasswordDeps()
	})

	It("does not pass requirements if you are not logged in", func() {
		deps.ReqFactory.LoginSuccess = false
		_, passed := callPassword([]string{}, deps)
		Expect(passed).To(BeFalse())
	})

	Context("when logged in successfully", func() {
		BeforeEach(func() {
			deps.ReqFactory.LoginSuccess = true
		})

		It("passes requirements", func() {
			_, passed := callPassword([]string{"", "", ""}, deps)
			Expect(passed).To(BeTrue())
		})

		It("changes your password when given a new password", func() {
			deps.PwdRepo.UpdateUnauthorized = false
			ui, _ := callPassword([]string{"old-password", "new-password", "new-password"}, deps)

			Expect(ui.PasswordPrompts).To(ContainSubstrings(
				[]string{"Current Password"},
				[]string{"New Password"},
				[]string{"Verify Password"},
			))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Changing password..."},
				[]string{"OK"},
				[]string{"Please log in again"},
			))

			Expect(deps.PwdRepo.UpdateNewPassword).To(Equal("new-password"))
			Expect(deps.PwdRepo.UpdateOldPassword).To(Equal("old-password"))

			Expect(deps.Config.AccessToken()).To(Equal(""))
			Expect(deps.Config.OrganizationFields()).To(Equal(models.OrganizationFields{}))
			Expect(deps.Config.SpaceFields()).To(Equal(models.SpaceFields{}))
		})

		It("fails when the password verification does not match", func() {
			ui, _ := callPassword([]string{"old-password", "new-password", "new-password-with-error"}, deps)

			Expect(ui.PasswordPrompts).To(ContainSubstrings(
				[]string{"Current Password"},
				[]string{"New Password"},
				[]string{"Verify Password"},
			))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"Password verification does not match"},
			))

			Expect(deps.PwdRepo.UpdateNewPassword).To(Equal(""))
		})

		It("fails when the current password does not match", func() {
			deps.PwdRepo.UpdateUnauthorized = true
			ui, _ := callPassword([]string{"old-password", "new-password", "new-password"}, deps)

			Expect(ui.PasswordPrompts).To(ContainSubstrings(
				[]string{"Current Password"},
				[]string{"New Password"},
				[]string{"Verify Password"},
			))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Changing password..."},
				[]string{"FAILED"},
				[]string{"Current password did not match"},
			))

			Expect(deps.PwdRepo.UpdateNewPassword).To(Equal("new-password"))
			Expect(deps.PwdRepo.UpdateOldPassword).To(Equal("old-password"))
		})
	})
})

type passwordDeps struct {
	ReqFactory *testreq.FakeReqFactory
	PwdRepo    *testapi.FakePasswordRepo
	Config     core_config.ReadWriter
}

func getPasswordDeps() passwordDeps {
	return passwordDeps{
		ReqFactory: &testreq.FakeReqFactory{LoginSuccess: true},
		PwdRepo:    &testapi.FakePasswordRepo{UpdateUnauthorized: true},
		Config:     testconfig.NewRepository(),
	}
}

func callPassword(inputs []string, deps passwordDeps) (*testterm.FakeUI, bool) {
	ui := &testterm.FakeUI{Inputs: inputs}
	cmd := NewPassword(ui, deps.PwdRepo, deps.Config)
	passed := testcmd.RunCommand(cmd, []string{}, deps.ReqFactory)

	return ui, passed
}
