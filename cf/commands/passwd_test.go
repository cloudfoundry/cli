package commands_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
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
	var (
		pwDeps passwordDeps
		ui     *testterm.FakeUI
		deps   command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = pwDeps.Config
		deps.RepoLocator = deps.RepoLocator.SetPasswordRepository(pwDeps.PwdRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("passwd").SetDependency(deps, pluginCall))
	}

	callPassword := func(inputs []string, pwDeps passwordDeps) (*testterm.FakeUI, bool) {
		ui = &testterm.FakeUI{Inputs: inputs}
		passed := testcmd.RunCliCommand("passwd", []string{}, pwDeps.ReqFactory, updateCommandDependency, false)
		return ui, passed
	}

	BeforeEach(func() {
		pwDeps = getPasswordDeps()
	})

	It("does not pass requirements if you are not logged in", func() {
		pwDeps.ReqFactory.LoginSuccess = false
		_, passed := callPassword([]string{}, pwDeps)
		Expect(passed).To(BeFalse())
	})

	Context("when logged in successfully", func() {
		BeforeEach(func() {
			pwDeps.ReqFactory.LoginSuccess = true
			pwDeps.PwdRepo.UpdateUnauthorized = false
		})

		It("passes requirements", func() {
			_, passed := callPassword([]string{"", "", ""}, pwDeps)
			Expect(passed).To(BeTrue())
		})

		It("changes your password when given a new password", func() {
			pwDeps.PwdRepo.UpdateUnauthorized = false
			ui, _ := callPassword([]string{"old-password", "new-password", "new-password"}, pwDeps)

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

			Expect(pwDeps.PwdRepo.UpdateNewPassword).To(Equal("new-password"))
			Expect(pwDeps.PwdRepo.UpdateOldPassword).To(Equal("old-password"))

			Expect(pwDeps.Config.AccessToken()).To(Equal(""))
			Expect(pwDeps.Config.OrganizationFields()).To(Equal(models.OrganizationFields{}))
			Expect(pwDeps.Config.SpaceFields()).To(Equal(models.SpaceFields{}))
		})

		It("fails when the password verification does not match", func() {
			ui, _ := callPassword([]string{"old-password", "new-password", "new-password-with-error"}, pwDeps)

			Expect(ui.PasswordPrompts).To(ContainSubstrings(
				[]string{"Current Password"},
				[]string{"New Password"},
				[]string{"Verify Password"},
			))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"Password verification does not match"},
			))

			Expect(pwDeps.PwdRepo.UpdateNewPassword).To(Equal(""))
		})

		It("fails when the current password does not match", func() {
			pwDeps.PwdRepo.UpdateUnauthorized = true
			ui, _ := callPassword([]string{"old-password", "new-password", "new-password"}, pwDeps)

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

			Expect(pwDeps.PwdRepo.UpdateNewPassword).To(Equal("new-password"))
			Expect(pwDeps.PwdRepo.UpdateOldPassword).To(Equal("old-password"))
		})
	})
})

type passwordDeps struct {
	ReqFactory *testreq.FakeReqFactory
	PwdRepo    *testapi.FakePasswordRepo
	Config     core_config.Repository
}

func getPasswordDeps() passwordDeps {
	return passwordDeps{
		ReqFactory: &testreq.FakeReqFactory{LoginSuccess: true},
		PwdRepo:    &testapi.FakePasswordRepo{UpdateUnauthorized: true},
		Config:     testconfig.NewRepository(),
	}
}
