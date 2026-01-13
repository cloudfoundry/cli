package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v9/command/commandfakes"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
	v7 "code.cloudfoundry.org/cli/v9/command/v7"
	"code.cloudfoundry.org/cli/v9/command/v7/shared/sharedfakes"
	"code.cloudfoundry.org/cli/v9/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/util/configv3"
	"code.cloudfoundry.org/cli/v9/util/ui"
	. "github.com/onsi/ginkgo/v2"
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
		app         resources.Application
	)

	BeforeEach(func() {
		app = resources.Application{Name: "some-app", GUID: "app-guid"}

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
		fakeActor.GetCurrentUserReturns(configv3.User{Name: "steve"}, nil)
		fakeSharedActor.CheckTargetReturns(nil)
		fakeActor.GetApplicationByNameAndSpaceReturns(
			app,
			v7action.Warnings{"get-app-warning"},
			nil,
		)
		fakeActor.GetNewestReadyPackageForApplicationReturns(
			resources.Package{GUID: "earliest-package-guid"},
			v7action.Warnings{"get-package-warning"},
			nil,
		)

		cmd.Strategy = flag.DeploymentStrategy{Name: constant.DeploymentStrategyRolling}
		maxInFlight := 4
		cmd.MaxInFlight = &maxInFlight
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
			fakeActor.GetCurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("No strategy flag is given", func() {
		BeforeEach(func() {
			cmd.Strategy.Name = constant.DeploymentStrategyDefault
			cmd.MaxInFlight = nil
		})
		It("warns that there will be app downtime", func() {
			Expect(testUI.Err).To(Say("This action will cause app downtime."))
		})
	})

	When("canary strategy is provided", func() {
		BeforeEach(func() {
			cmd.Strategy = flag.DeploymentStrategy{Name: constant.DeploymentStrategyCanary}
			cmd.InstanceSteps = "1,2,4"
			fakeConfig = &commandfakes.FakeConfig{}
			fakeConfig.APIVersionReturns("4.0.0")
			cmd.Config = fakeConfig
		})

		It("starts the app with the current droplet", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakeAppStager.StageAndStartCallCount()).To(Equal(1))

			inputApp, inputSpace, inputOrg, inputDropletGuid, opts := fakeAppStager.StageAndStartArgsForCall(0)
			Expect(inputApp).To(Equal(app))
			Expect(inputDropletGuid).To(Equal("earliest-package-guid"))
			Expect(inputSpace).To(Equal(cmd.Config.TargetedSpace()))
			Expect(inputOrg).To(Equal(cmd.Config.TargetedOrganization()))
			Expect(opts.Strategy).To(Equal(constant.DeploymentStrategyCanary))
			Expect(opts.AppAction).To(Equal(constant.ApplicationRestarting))
			Expect(opts.CanarySteps).To(Equal([]resources.CanaryStep{{InstanceWeight: 1}, {InstanceWeight: 2}, {InstanceWeight: 4}}))
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
				resources.Application{},
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
		app := fakeActor.GetNewestReadyPackageForApplicationArgsForCall(0)
		Expect(app).To(Equal(app))

		Expect(testUI.Err).To(Say("get-package-warning"))
	})

	When("getting the package fails", func() {
		BeforeEach(func() {
			fakeActor.GetNewestReadyPackageForApplicationReturns(
				resources.Package{},
				v7action.Warnings{"get-package-warning"},
				errors.New("get-package-error"),
			)
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("get-package-error"))
		})
	})

	It("stages and starts the app", func() {
		Expect(fakeAppStager.StageAndStartCallCount()).To(Equal(1))
		returnedApp, spaceForApp, orgForApp, pkgGUID, opts := fakeAppStager.StageAndStartArgsForCall(0)
		Expect(returnedApp).To(Equal(app))
		Expect(spaceForApp).To(Equal(fakeConfig.TargetedSpace()))
		Expect(orgForApp).To(Equal(fakeConfig.TargetedOrganization()))
		Expect(pkgGUID).To(Equal("earliest-package-guid"))
		Expect(opts.Strategy).To(Equal(constant.DeploymentStrategyRolling))
		Expect(opts.MaxInFlight).To(Equal(4))
		Expect(opts.NoWait).To(Equal(false))
		Expect(opts.AppAction).To(Equal(constant.ApplicationRestarting))
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

	DescribeTable("ValidateFlags returns an error",
		func(setup func(), expectedErr error) {
			setup()
			err := cmd.ValidateFlags()
			if expectedErr == nil {
				Expect(err).To(BeNil())
			} else {
				Expect(err).To(MatchError(expectedErr))
			}
		},
		Entry("max-in-flight is passed without strategy",
			func() {
				cmd.Strategy = flag.DeploymentStrategy{Name: constant.DeploymentStrategyDefault}
				maxInFlight := 10
				cmd.MaxInFlight = &maxInFlight
			},
			translatableerror.RequiredFlagsError{
				Arg1: "--max-in-flight",
				Arg2: "--strategy",
			}),

		Entry("max-in-flight is smaller than 1",
			func() {
				cmd.Strategy = flag.DeploymentStrategy{Name: constant.DeploymentStrategyRolling}
				maxInFlight := 0
				cmd.MaxInFlight = &maxInFlight
			},
			translatableerror.IncorrectUsageError{
				Message: "--max-in-flight must be greater than or equal to 1",
			}),

		Entry("instance-steps provided with rolling deployment",
			func() {
				cmd.Strategy = flag.DeploymentStrategy{Name: constant.DeploymentStrategyRolling}
				cmd.InstanceSteps = "1,2,3"
			},
			translatableerror.RequiredFlagsError{
				Arg1: "--instance-steps",
				Arg2: "--strategy=canary",
			}),

		Entry("instance-steps no strategy provided",
			func() {
				cmd.InstanceSteps = "1,2,3"
			},
			translatableerror.RequiredFlagsError{
				Arg1: "--instance-steps",
				Arg2: "--strategy=canary",
			}),

		Entry("instance-steps a valid list of ints",
			func() {
				cmd.Strategy = flag.DeploymentStrategy{Name: constant.DeploymentStrategyCanary}
				cmd.InstanceSteps = "some,thing,not,right"
			},
			translatableerror.ParseArgumentError{
				ArgumentName: "--instance-steps",
				ExpectedType: "list of weights",
			}),

		Entry("instance-steps used when CAPI does not support canary steps",
			func() {
				cmd.InstanceSteps = "1,2,3"
				cmd.Strategy.Name = constant.DeploymentStrategyCanary
				fakeConfig = &commandfakes.FakeConfig{}
				fakeConfig.APIVersionReturns("3.0.0")
				cmd.Config = fakeConfig
			},
			translatableerror.MinimumCFAPIVersionNotMetError{
				Command:        "--instance-steps",
				CurrentVersion: "3.0.0",
				MinimumVersion: "3.189.0",
			}),
	)
})
