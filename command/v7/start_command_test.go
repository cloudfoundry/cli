package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("start Command", func() {
	var (
		cmd                v7.StartCommand
		testUI             *ui.UI
		fakeConfig         *commandfakes.FakeConfig
		fakeSharedActor    *commandfakes.FakeSharedActor
		fakeActor          *v7fakes.FakeActor
		fakeLogCacheClient *sharedactionfakes.FakeLogCacheClient
		fakeAppStager      *sharedfakes.FakeAppStager

		binaryName string
		executeErr error
		app        resources.Application
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		fakeLogCacheClient = new(sharedactionfakes.FakeLogCacheClient)
		fakeAppStager = new(sharedfakes.FakeAppStager)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = resources.Application{Name: "app-name", GUID: "app-guid"}

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
		})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: "some-space-guid",
		})
		fakeActor.GetCurrentUserReturns(configv3.User{Name: "steve"}, nil)
		fakeActor.GetApplicationByNameAndSpaceReturns(app, v7action.Warnings{"get-app-warning"}, nil)

		cmd = v7.StartCommand{
			RequiredArgs: flag.AppName{AppName: app.Name},
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			LogCacheClient: fakeLogCacheClient,
			Stager:         fakeAppStager,
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
			fakeActor.GetCurrentUserReturns(configv3.User{}, expectedErr)
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
			fakeActor.GetApplicationByNameAndSpaceReturns(app, v7action.Warnings{"get-app-warning"}, nil)
		})

		When("the app is stopped", func() {
			BeforeEach(func() {
				app = resources.Application{Name: "app-name", GUID: "app-guid", State: constant.ApplicationStopped}
				fakeActor.GetApplicationByNameAndSpaceReturns(app, v7action.Warnings{"get-app-warning"}, nil)
			})

			It("stages the new package and starts the app with the new droplet", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeAppStager.StageAndStartCallCount()).To(Equal(1))

				inputApp, inputSpace, inputOrg, inputPkgGUID, opts := fakeAppStager.StageAndStartArgsForCall(0)
				Expect(inputApp).To(Equal(app))
				Expect(inputSpace).To(Equal(cmd.Config.TargetedSpace()))
				Expect(inputOrg).To(Equal(cmd.Config.TargetedOrganization()))
				Expect(inputPkgGUID).To(Equal("package-guid"))
				Expect(opts.Strategy).To(Equal(constant.DeploymentStrategyDefault))
				Expect(opts.NoWait).To(Equal(false))
				Expect(opts.AppAction).To(Equal(constant.ApplicationStarting))
			})

			When("staging and starting the app returns an error", func() {
				BeforeEach(func() {
					fakeAppStager.StageAndStartReturns(errors.New("stage-and-start-error"))
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("stage-and-start-error"))
				})
			})
		})

		When("the app is started", func() {
			BeforeEach(func() {
				app = resources.Application{Name: "app-name", GUID: "app-guid", State: constant.ApplicationStarted}
				fakeActor.GetApplicationByNameAndSpaceReturns(app, v7action.Warnings{"get-app-warning"}, nil)
			})

			It("starts the app with the current droplet", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeAppStager.StartAppCallCount()).To(Equal(1))

				inputApp, inputSpace, inputOrg, inputDropletGuid, opts := fakeAppStager.StartAppArgsForCall(0)
				Expect(inputApp).To(Equal(app))
				Expect(inputDropletGuid).To(Equal(""))
				Expect(inputSpace).To(Equal(cmd.Config.TargetedSpace()))
				Expect(inputOrg).To(Equal(cmd.Config.TargetedOrganization()))
				Expect(opts.Strategy).To(Equal(constant.DeploymentStrategyDefault))
				Expect(opts.NoWait).To(Equal(false))
				Expect(opts.AppAction).To(Equal(constant.ApplicationStarting))
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

	Context("no new package is available", func() {
		BeforeEach(func() {
			fakeActor.GetUnstagedNewestPackageGUIDReturns("", v7action.Warnings{}, nil)
		})

		It("starts the app with the current droplet", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakeAppStager.StartAppCallCount()).To(Equal(1))

			inputApp, inputSpace, inputOrg, inputDropletGuid, opts := fakeAppStager.StartAppArgsForCall(0)
			Expect(inputApp).To(Equal(app))
			Expect(inputDropletGuid).To(Equal(""))
			Expect(inputSpace).To(Equal(cmd.Config.TargetedSpace()))
			Expect(inputOrg).To(Equal(cmd.Config.TargetedOrganization()))
			Expect(opts.Strategy).To(Equal(constant.DeploymentStrategyDefault))
			Expect(opts.NoWait).To(Equal(false))
			Expect(opts.AppAction).To(Equal(constant.ApplicationStarting))
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

	When("getting attempting to get the unstaged package returns an error", func() {
		var expectedErr error
		BeforeEach(func() {
			expectedErr = errors.New("error getting package")
			fakeActor.GetUnstagedNewestPackageGUIDReturns("", v7action.Warnings{"needs-stage-warnings"}, expectedErr)
		})

		It("errors", func() {
			Expect(testUI.Err).To(Say("needs-stage-warnings"))
			Expect(executeErr).To(MatchError(expectedErr))
			Expect(fakeActor.GetUnstagedNewestPackageGUIDCallCount()).To(Equal(1))
			appGUID := fakeActor.GetUnstagedNewestPackageGUIDArgsForCall(0)
			Expect(appGUID).To(Equal("app-guid"))
		})
	})

})
