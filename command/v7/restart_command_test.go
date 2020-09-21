package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/shared/sharedfakes"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("restart Command", func() {
	var (
		cmd             v7.RestartCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		fakeAppStager   *sharedfakes.FakeAppStager

		binaryName string
		executeErr error
		app        resources.Application
		strategy   constant.DeploymentStrategy
		noWait     bool
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		fakeAppStager = new(sharedfakes.FakeAppStager)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = resources.Application{Name: "app-name", GUID: "app-guid"}
		strategy = constant.DeploymentStrategyDefault
		noWait = false

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
		})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: "some-space-guid",
		})
		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
		fakeActor.GetApplicationByNameAndSpaceReturns(app, v7action.Warnings{"get-app-warning"}, nil)

		cmd = v7.RestartCommand{
			RequiredArgs: flag.AppName{AppName: app.Name},
			Strategy:     flag.DeploymentStrategy{Name: strategy},
			NoWait:       noWait,

			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			Stager: fakeAppStager,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	It("gets the application", func() {
		Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
		inputAppName, inputSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
		Expect(inputAppName).To(Equal(app.Name))
		Expect(inputSpaceGUID).To(Equal("some-space-guid"))
	})

	When("Getting the application fails", func() {
		BeforeEach(func() {
			fakeActor.GetApplicationByNameAndSpaceReturns(
				resources.Application{},
				v7action.Warnings{"get-app-warning"},
				errors.New("get-app-error"),
			)
		})

		It("displays all warnings and returns an error", func() {
			Expect(executeErr).To(MatchError("get-app-error"))
			Expect(testUI.Err).To(Say("get-app-warning"))
		})
	})

	Context("a new package is available", func() {
		BeforeEach(func() {
			fakeActor.GetUnstagedNewestPackageGUIDReturns("package-guid", v7action.Warnings{}, nil)
		})

		It("stages the new package and starts the app with the new droplet", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakeAppStager.StageAndStartCallCount()).To(Equal(1))

			inputApp, inputSpace, inputOrg, inputPkgGUID, inputStrategy, inputNoWait, inputAppAction := fakeAppStager.StageAndStartArgsForCall(0)
			Expect(inputApp).To(Equal(app))
			Expect(inputSpace).To(Equal(cmd.Config.TargetedSpace()))
			Expect(inputOrg).To(Equal(cmd.Config.TargetedOrganization()))
			Expect(inputPkgGUID).To(Equal("package-guid"))
			Expect(inputStrategy).To(Equal(strategy))
			Expect(inputNoWait).To(Equal(noWait))
			Expect(inputAppAction).To(Equal(constant.ApplicationRestarting))
		})

		Context("staging and starting the app returns an error", func() {
			BeforeEach(func() {
				fakeAppStager.StageAndStartReturns(errors.New("stage-and-start-error"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("stage-and-start-error"))
			})
		})
	})

	Context("no new package is available", func() {
		BeforeEach(func() {
			fakeActor.GetUnstagedNewestPackageGUIDReturns("", v7action.Warnings{}, nil)
		})

		It("starts the app with the current droplet", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakeAppStager.StartAppCallCount()).To(Equal(1))

			inputApp, inputDropletGuid, inputStrategy, inputNoWait, inputSpace, inputOrg, inputAppAction := fakeAppStager.StartAppArgsForCall(0)
			Expect(inputApp).To(Equal(app))
			Expect(inputDropletGuid).To(Equal(""))
			Expect(inputStrategy).To(Equal(strategy))
			Expect(inputNoWait).To(Equal(noWait))
			Expect(inputSpace).To(Equal(cmd.Config.TargetedSpace()))
			Expect(inputOrg).To(Equal(cmd.Config.TargetedOrganization()))
			Expect(inputAppAction).To(Equal(constant.ApplicationRestarting))
		})

		When("starting the app returns an error", func() {
			BeforeEach(func() {
				fakeAppStager.StartAppReturns(errors.New("start-error"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("start-error"))
			})
		})
	})

})
