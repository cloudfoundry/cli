package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/shared/sharedfakes"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("restage Command", func() {
	var (
		cmd             v7.RestageCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		fakeAppStager   *sharedfakes.FakeAppStager

		executeErr  error
		expectedErr error
		app         v7action.Application
	)

	BeforeEach(func() {
		app = v7action.Application{Name: "some-app", GUID: "app-guid"}

		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		fakeAppStager = new(sharedfakes.FakeAppStager)

		cmd = v7.RestageCommand{
			RequiredArgs: flag.AppName{AppName: app.Name},
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			Stager: fakeAppStager,
		}

		fakeConfig.BinaryNameReturns("some-binary-name")
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
		})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: "some-space-guid",
		})
		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
		fakeSharedActor.CheckTargetReturns(nil)
		fakeActor.GetApplicationByNameAndSpaceReturns(
			app,
			v7action.Warnings{"get-app-warning"},
			nil,
		)
		fakeActor.GetNewestReadyPackageForApplicationReturns(
			v7action.Package{GUID: "earliest-package-guid"},
			v7action.Warnings{"get-package-warning"},
			nil,
		)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: "binary"})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: "binary"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is not logged in", func() {
		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("it's NOT a rolling deploy", func() {
		BeforeEach(func() {
			cmd.Strategy.Name = constant.DeploymentStrategyDefault
		})
		It("warns that there will be app downtime", func() {
			Expect(testUI.Err).To(Say("This action will cause app downtime."))
		})
	})

	It("displays that it's restaging", func() {
		Expect(testUI.Out).To(Say("Restaging app some-app in org some-org / space some-space as steve..."))
	})

	It("gets the application and displays all warnings", func() {
		Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
		appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
		Expect(appName).To(Equal("some-app"))
		Expect(spaceGUID).To(Equal("some-space-guid"))

		Expect(testUI.Err).To(Say("get-app-warning"))
	})

	When("getting the application fails", func() {
		BeforeEach(func() {
			fakeActor.GetApplicationByNameAndSpaceReturns(
				v7action.Application{},
				v7action.Warnings{"get-app-warning"},
				errors.New("get-app-error"),
			)
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("get-app-error"))
		})
	})

	It("gets the newest package and displays all warnings", func() {
		Expect(fakeActor.GetNewestReadyPackageForApplicationCallCount()).To(Equal(1))
		appGUID := fakeActor.GetNewestReadyPackageForApplicationArgsForCall(0)
		Expect(appGUID).To(Equal("app-guid"))

		Expect(testUI.Err).To(Say("get-package-warning"))
	})

	When("getting the package fails", func() {
		BeforeEach(func() {
			fakeActor.GetNewestReadyPackageForApplicationReturns(
				v7action.Package{},
				v7action.Warnings{"get-package-warning"},
				errors.New("get-package-error"),
			)
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("get-package-error"))
		})

		When("there are no packages available to stage", func() {
			BeforeEach(func() {
				fakeActor.GetNewestReadyPackageForApplicationReturns(
					v7action.Package{},
					v7action.Warnings{"get-package-warning"},
					actionerror.PackageNotFoundInAppError{},
				)
			})

			It("displays all warnings and returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.PackageNotFoundInAppError{
					AppName: app.Name, BinaryName: "some-binary-name"}))
				Expect(testUI.Err).To(Say("get-package-warning"))
			})
		})
	})

	It("stages and starts the app", func() {
		Expect(fakeAppStager.StageAndStartCallCount()).To(Equal(1))
		returnedApp, pkgGUID, strategy, noWait := fakeAppStager.StageAndStartArgsForCall(0)
		Expect(returnedApp).To(Equal(app))
		Expect(pkgGUID).To(Equal("earliest-package-guid"))
		Expect(strategy).To(Equal(constant.DeploymentStrategyDefault))
		Expect(noWait).To(Equal(false))
	})

	When("staging and starting the app fails", func() {
		BeforeEach(func() {
			fakeAppStager.StageAndStartReturns(errors.New("stage-and-start-error"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("stage-and-start-error"))
		})
	})

	It("succeeds", func() {
		Expect(executeErr).To(Not(HaveOccurred()))
	})
})
