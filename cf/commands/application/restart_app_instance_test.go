package application_test

import (
	"errors"

	testApplication "github.com/cloudfoundry/cli/cf/api/app_instances/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("restart-app-instance", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.Repository
		appInstancesRepo    *testApplication.FakeAppInstancesRepository
		requirementsFactory *testreq.FakeReqFactory
		application         models.Application
		deps                command_registry.Dependency
	)

	BeforeEach(func() {
		application = models.Application{}
		application.Name = "my-app"
		application.Guid = "my-app-guid"
		application.InstanceCount = 1

		ui = &testterm.FakeUI{}
		appInstancesRepo = &testApplication.FakeAppInstancesRepository{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
			Application:          application,
		}
	})

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = config
		deps.RepoLocator = deps.RepoLocator.SetAppInstancesRepository(appInstancesRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("restart-app-instance").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("restart-app-instance", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails if not logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand("my-app", "0")).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false
			Expect(runCommand("my-app", "0")).To(BeFalse())
		})

		It("fails when there is not exactly two arguments", func() {
			Expect(runCommand("my-app")).To(BeFalse())
			Expect(runCommand("my-app", "0", "0")).To(BeFalse())
			Expect(runCommand()).To(BeFalse())
		})
	})

	Describe("restarting an instance of an application", func() {
		It("correctly 'restarts' the desired instance", func() {
			runCommand("my-app", "0")

			app_guid, instance := appInstancesRepo.DeleteInstanceArgsForCall(0)
			Expect(app_guid).To(Equal(application.Guid))
			Expect(instance).To(Equal(0))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Restarting instance 0 of application my-app as my-user"},
				[]string{"OK"},
			))
		})

		Context("when deleting the app instance fails", func() {
			BeforeEach(func() {
				appInstancesRepo.DeleteInstanceReturns(errors.New("deletion failed"))
			})
			It("fails", func() {
				runCommand("my-app", "0")

				app_guid, instance := appInstancesRepo.DeleteInstanceArgsForCall(0)
				Expect(app_guid).To(Equal(application.Guid))
				Expect(instance).To(Equal(0))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"deletion failed"},
				))
			})
		})

		Context("when the instance passed is not an non-negative integer", func() {
			It("fails when it is a string", func() {
				runCommand("my-app", "some-silly-thing")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Instance must be a non-negative integer"},
				))
			})
		})
	})
})
