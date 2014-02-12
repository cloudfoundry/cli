package application_test

import (
	. "cf/commands/application"
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
	It("TestScaleRequirements", func() {
		args := []string{"-m", "1G", "my-app"}
		deps := getScaleDependencies()

		deps.reqFactory.LoginSuccess = false
		deps.reqFactory.TargetedSpaceSuccess = true
		callScale(args, deps)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		deps.reqFactory.LoginSuccess = true
		deps.reqFactory.TargetedSpaceSuccess = false
		callScale(args, deps)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		deps.reqFactory.LoginSuccess = true
		deps.reqFactory.TargetedSpaceSuccess = true
		callScale(args, deps)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		Expect(deps.reqFactory.ApplicationName).To(Equal("my-app"))
	})
	It("TestScaleFailsWithUsage", func() {

		deps := getScaleDependencies()

		ui := callScale([]string{}, deps)

		Expect(ui.FailedWithUsage).To(BeTrue())
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestScaleFailsWithoutFlags", func() {

		args := []string{"my-app"}
		deps := getScaleDependencies()
		deps.reqFactory.LoginSuccess = true
		deps.reqFactory.TargetedSpaceSuccess = true

		callScale(args, deps)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestScaleAll", func() {

		app := maker.NewApp(maker.Overrides{"name": "my-app", "guid": "my-app-guid"})
		deps := getScaleDependencies()
		deps.reqFactory.Application = app
		deps.appRepo.UpdateAppResult = app

		ui := callScale([]string{"-i", "5", "-m", "512M", "my-app"}, deps)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Scaling", "my-app", "my-org", "my-space", "my-user"},
			{"OK"},
		})

		Expect(deps.restarter.AppToRestart.Guid).To(Equal("my-app-guid"))
		Expect(deps.appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
		Expect(*deps.appRepo.UpdateParams.Memory).To(Equal(uint64(512)))
		Expect(*deps.appRepo.UpdateParams.InstanceCount).To(Equal(5))
	})
	It("TestScaleOnlyInstances", func() {

		app := maker.NewApp(maker.Overrides{"name": "my-app", "guid": "my-app-guid"})
		deps := getScaleDependencies()
		deps.reqFactory.Application = app
		deps.appRepo.UpdateAppResult = app

		callScale([]string{"-i", "5", "my-app"}, deps)

		Expect(deps.restarter.AppToRestart.Guid).To(Equal(""))
		Expect(deps.appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
		Expect(*deps.appRepo.UpdateParams.InstanceCount).To(Equal(5))
		Expect(deps.appRepo.UpdateParams.DiskQuota).To(BeNil())
		Expect(deps.appRepo.UpdateParams.Memory).To(BeNil())
	})
	It("TestScaleOnlyMemory", func() {

		app := maker.NewApp(maker.Overrides{"name": "my-app", "guid": "my-app-guid"})
		deps := getScaleDependencies()
		deps.reqFactory.Application = app
		deps.appRepo.UpdateAppResult = app

		callScale([]string{"-m", "512M", "my-app"}, deps)

		Expect(deps.restarter.AppToRestart.Guid).To(Equal("my-app-guid"))
		Expect(deps.appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
		Expect(*deps.appRepo.UpdateParams.Memory).To(Equal(uint64(512)))
		Expect(deps.appRepo.UpdateParams.DiskQuota).To(BeNil())
		Expect(deps.appRepo.UpdateParams.InstanceCount).To(BeNil())
	})
})

type scaleDependencies struct {
	reqFactory *testreq.FakeReqFactory
	restarter  *testcmd.FakeAppRestarter
	appRepo    *testapi.FakeApplicationRepository
}

func getScaleDependencies() (deps scaleDependencies) {
	deps = scaleDependencies{
		reqFactory: &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true},
		restarter:  &testcmd.FakeAppRestarter{},
		appRepo:    &testapi.FakeApplicationRepository{},
	}
	return
}

func callScale(args []string, deps scaleDependencies) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("scale", args)
	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewScale(ui, configRepo, deps.restarter, deps.appRepo)
	testcmd.RunCommand(cmd, ctxt, deps.reqFactory)
	return
}
