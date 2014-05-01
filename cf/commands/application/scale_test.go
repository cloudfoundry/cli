package application_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/configuration"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/cli/testhelpers/maker"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("scale command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		restarter           *testcmd.FakeAppRestarter
		appRepo             *testapi.FakeApplicationRepository
		ui                  *testterm.FakeUI
		configRepo          configuration.Repository
		cmd                 *Scale
	)

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		restarter = &testcmd.FakeAppRestarter{}
		appRepo = &testapi.FakeApplicationRepository{}
		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
		cmd = NewScale(ui, configRepo, restarter, appRepo)
	})

	Describe("requirements", func() {
		It("requires the user to be logged in with a targed space", func() {
			args := []string{"-m", "1G", "my-app"}

			requirementsFactory.LoginSuccess = false
			requirementsFactory.TargetedSpaceSuccess = true

			testcmd.RunCommand(cmd, testcmd.NewContext("scale", args), requirementsFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = false

			testcmd.RunCommand(cmd, testcmd.NewContext("scale", args), requirementsFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("requires an app to be specified", func() {
			testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"-m", "1G"}), requirementsFactory)

			Expect(ui.FailedWithUsage).To(BeTrue())
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("does not require any flags", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true

			testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"my-app"}), requirementsFactory)

			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})
	})

	Describe("checking for bad flags", func() {
		It("fails when non-positive value is given for memory limit", func() {
			testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"-m", "0M", "my-app"}), requirementsFactory)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"memory"},
				[]string{"positive integer"},
			))
		})

		It("fails when non-positive value is given for instances", func() {
			testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"-i", "-15", "my-app"}), requirementsFactory)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"Invalid instance count"},
				[]string{"must be a positive integer"},
			))
		})

		It("fails when non-positive value is given for disk quota", func() {
			testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"-k", "-1G", "my-app"}), requirementsFactory)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"disk quota"},
				[]string{"positive integer"},
			))
		})
	})

	Describe("scaling an app", func() {
		BeforeEach(func() {
			app := maker.NewApp(maker.Overrides{"name": "my-app", "guid": "my-app-guid"})
			app.InstanceCount = 42
			app.DiskQuota = 1024
			app.Memory = 256

			requirementsFactory.Application = app
			appRepo.UpdateAppResult = app
		})

		Context("when no flags are specified", func() {
			It("prints a description of the app's limits", func() {
				testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"my-app"}), requirementsFactory)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Showing", "my-app", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{"memory", "256M"},
					[]string{"disk", "1G"},
					[]string{"instances", "42"},
				))

				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"Scaling", "my-app", "my-org", "my-space", "my-user"}))
			})
		})

		Context("when the user does not confirm 'yes'", func() {
			It("does not restart the app", func() {
				ui.Inputs = []string{"whatever"}
				testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"-i", "5", "-m", "512M", "-k", "2G", "my-app"}), requirementsFactory)

				Expect(restarter.AppToRestart.Guid).To(Equal(""))
			})
		})

		Context("when the user provides the -f flag", func() {
			It("does not prompt the user", func() {
				testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"-f", "-i", "5", "-m", "512M", "-k", "2G", "my-app"}), requirementsFactory)
				Expect(restarter.AppToRestart.Guid).To(Equal("my-app-guid"))
			})
		})

		Context("when the user confirms they want to restart", func() {
			BeforeEach(func() {
				ui.Inputs = []string{"yes"}
			})

			It("can set an app's instance count, memory limit and disk limit", func() {
				testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"-i", "5", "-m", "512M", "-k", "2G", "my-app"}), requirementsFactory)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Scaling", "my-app", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))

				Expect(ui.Prompts).To(ContainSubstrings([]string{"This will cause the app to restart", "Are you sure", "my-app"}))
				Expect(restarter.AppToRestart.Guid).To(Equal("my-app-guid"))
				Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
				Expect(*appRepo.UpdateParams.Memory).To(Equal(uint64(512)))
				Expect(*appRepo.UpdateParams.InstanceCount).To(Equal(5))
				Expect(*appRepo.UpdateParams.DiskQuota).To(Equal(uint64(2048)))
			})

			It("does not scale the memory and disk limits if they are not specified", func() {
				testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"-i", "5", "my-app"}), requirementsFactory)

				Expect(restarter.AppToRestart.Guid).To(Equal(""))
				Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
				Expect(*appRepo.UpdateParams.InstanceCount).To(Equal(5))
				Expect(appRepo.UpdateParams.DiskQuota).To(BeNil())
				Expect(appRepo.UpdateParams.Memory).To(BeNil())
			})

			It("does not scale the app's instance count if it is not specified", func() {
				testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"-m", "512M", "my-app"}), requirementsFactory)

				Expect(restarter.AppToRestart.Guid).To(Equal("my-app-guid"))
				Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
				Expect(*appRepo.UpdateParams.Memory).To(Equal(uint64(512)))
				Expect(appRepo.UpdateParams.DiskQuota).To(BeNil())
				Expect(appRepo.UpdateParams.InstanceCount).To(BeNil())
			})
		})
	})
})
