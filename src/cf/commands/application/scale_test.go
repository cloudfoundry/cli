package application_test

import (
	. "cf/commands/application"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	"testhelpers/maker"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	var (
		reqFactory *testreq.FakeReqFactory
		restarter  *testcmd.FakeAppRestarter
		appRepo    *testapi.FakeApplicationRepository
		ui         *testterm.FakeUI
		configRepo configuration.Repository
		cmd        *Scale
	)

	BeforeEach(func() {
		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		restarter = &testcmd.FakeAppRestarter{}
		appRepo = &testapi.FakeApplicationRepository{}
		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
		cmd = NewScale(ui, configRepo, restarter, appRepo)
	})

	Describe("requirements", func() {
		It("requires the user to be logged in with a targed space", func() {
			args := []string{"-m", "1G", "my-app"}

			reqFactory.LoginSuccess = false
			reqFactory.TargetedSpaceSuccess = true

			testcmd.RunCommand(cmd, testcmd.NewContext("scale", args), reqFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

			reqFactory.LoginSuccess = true
			reqFactory.TargetedSpaceSuccess = false

			testcmd.RunCommand(cmd, testcmd.NewContext("scale", args), reqFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

			reqFactory.LoginSuccess = true
			reqFactory.TargetedSpaceSuccess = true

			testcmd.RunCommand(cmd, testcmd.NewContext("scale", args), reqFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
			Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		})

		It("requires an app to be specified", func() {
			testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"-m", "1G"}), reqFactory)

			Expect(ui.FailedWithUsage).To(BeTrue())
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("does not require any flags", func() {
			reqFactory.LoginSuccess = true
			reqFactory.TargetedSpaceSuccess = true

			testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"my-app"}), reqFactory)

			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})
	})

	Describe("scaling an app", func() {
		BeforeEach(func() {
			app := maker.NewApp(maker.Overrides{"name": "my-app", "guid": "my-app-guid"})
			reqFactory.Application = app
			appRepo.UpdateAppResult = app
		})

		Context("when no flags are specified", func() {
			It("prints a description of the app's limits", func() {
				testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"my-app"}), reqFactory)

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"memory", "256M"},
					{"disk", "1G"},
					{"instances", "42"},
				})
			})
		})

		It("can set an app's instance count, memory limit and disk limit", func() {
			testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"-i", "5", "-m", "512M", "-k", "2G", "my-app"}), reqFactory)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Scaling", "my-app", "my-org", "my-space", "my-user"},
				{"OK"},
			})

			Expect(restarter.AppToRestart.Guid).To(Equal("my-app-guid"))
			Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
			Expect(*appRepo.UpdateParams.Memory).To(Equal(uint64(512)))
			Expect(*appRepo.UpdateParams.InstanceCount).To(Equal(5))
			Expect(*appRepo.UpdateParams.DiskQuota).To(Equal(uint64(2048)))
		})

		It("does not scale the memory and disk limits if they are not specified", func() {
			testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"-i", "5", "my-app"}), reqFactory)

			Expect(restarter.AppToRestart.Guid).To(Equal(""))
			Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
			Expect(*appRepo.UpdateParams.InstanceCount).To(Equal(5))
			Expect(appRepo.UpdateParams.DiskQuota).To(BeNil())
			Expect(appRepo.UpdateParams.Memory).To(BeNil())
		})

		It("does not scale the app's instance count if it is not specified", func() {
			testcmd.RunCommand(cmd, testcmd.NewContext("scale", []string{"-m", "512M", "my-app"}), reqFactory)

			Expect(restarter.AppToRestart.Guid).To(Equal("my-app-guid"))
			Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
			Expect(*appRepo.UpdateParams.Memory).To(Equal(uint64(512)))
			Expect(appRepo.UpdateParams.DiskQuota).To(BeNil())
			Expect(appRepo.UpdateParams.InstanceCount).To(BeNil())
		})
	})
})
