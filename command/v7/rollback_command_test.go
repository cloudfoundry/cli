package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/command/commandfakes"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
	v7 "code.cloudfoundry.org/cli/v8/command/v7"
	"code.cloudfoundry.org/cli/v8/command/v7/shared/sharedfakes"
	"code.cloudfoundry.org/cli/v8/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/types"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("rollback Command", func() {
	var (
		appName         string
		binaryName      string
		executeErr      error
		fakeActor       *v7fakes.FakeActor
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		input           *Buffer
		testUI          *ui.UI
		app             resources.Application

		fakeAppStager *sharedfakes.FakeAppStager
		cmd           v7.RollbackCommand
	)

	BeforeEach(func() {
		appName = "some-app"
		binaryName = "faceman"
		fakeActor = new(v7fakes.FakeActor)
		fakeAppStager = new(sharedfakes.FakeAppStager)
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())

		revisions := []resources.Revision{
			resources.Revision{Version: 2},
			resources.Revision{Version: 1},
		}
		app = resources.Application{
			GUID: "123",
			Name: "some-app",
		}

		fakeActor.GetRevisionsByApplicationNameAndSpaceReturns(
			revisions, v7action.Warnings{"warning-2"}, nil,
		)

		fakeConfig.BinaryNameReturns(binaryName)
		fakeActor.GetCurrentUserReturns(configv3.User{Name: "steve"}, nil)
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
			GUID: "some-org-guid",
		})

		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: "some-space-guid",
		})

		cmd = v7.RollbackCommand{
			RequiredArgs: flag.AppName{AppName: appName},
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				Actor:       fakeActor,
				SharedActor: fakeSharedActor,
			},
			Stager: fakeAppStager,
		}
		maxInFlight := 5
		cmd.MaxInFlight = &maxInFlight
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

		It("returns an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("failing to retrieve the app", func() {
		BeforeEach(func() {
			fakeActor.GetApplicationByNameAndSpaceReturns(resources.Application{}, v7action.Warnings{"warning-1", "warning-2"}, errors.New("oh no"))
		})

		It("returns an error and all warnings", func() {
			Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1), "GetApplicationByNameAndSpace call count")
			Expect(executeErr).To(MatchError("oh no"))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("there is a failure fetching the revision", func() {
		BeforeEach(func() {
			fakeActor.GetRevisionByApplicationAndVersionReturns(resources.Revision{}, v7action.Warnings{"warning-1", "warning-2"}, errors.New("oh no"))
		})

		It("returns an error and all warnings", func() {
			Expect(fakeActor.GetRevisionByApplicationAndVersionCallCount()).To(Equal(1), "GetRevisionByApplicationAndVersion call count")
			Expect(executeErr).To(MatchError("oh no"))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("the first revision is set as the rollback target", func() {
		BeforeEach(func() {
			cmd.Version = flag.Revision{NullInt: types.NullInt{Value: 1, IsSet: true}}
		})

		When("the app has at least one revision", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(
					app,
					v7action.Warnings{"app-warning-1"},
					nil,
				)

				fakeActor.GetRevisionByApplicationAndVersionReturns(
					resources.Revision{Version: 1, GUID: "some-1-guid"},
					v7action.Warnings{"revision-warning-3"},
					nil,
				)

				fakeAppStager.StartAppReturns(
					nil,
				)
			})

			It("fetches the app and revision revision", func() {
				Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1), "GetApplicationByNameAndSpace call count")
				appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(appName).To(Equal(appName))
				Expect(spaceGUID).To(Equal("some-space-guid"))

				Expect(fakeActor.GetRevisionByApplicationAndVersionCallCount()).To(Equal(1), "GetRevisionByApplicationAndVersion call count")
				appGUID, version := fakeActor.GetRevisionByApplicationAndVersionArgsForCall(0)
				Expect(appGUID).To(Equal("123"))
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(version).To(Equal(1))
			})

			When("the user passes the force flag", func() {
				BeforeEach(func() {
					cmd.Force = true
				})

				It("skips the prompt and executes the rollback", func() {
					Expect(fakeAppStager.StartAppCallCount()).To(Equal(1), "GetStartApp call count")

					application, _, _, revisionGUID, opts := fakeAppStager.StartAppArgsForCall(0)
					Expect(application.GUID).To(Equal("123"))
					Expect(revisionGUID).To(Equal("some-1-guid"))
					Expect(opts.AppAction).To(Equal(constant.ApplicationRollingBack))

					Expect(testUI.Out).ToNot(Say("Rolling '%s' back to revision '1' will create a new revision. The new revision '3' will use the settings from revision '1'.", appName))
					Expect(testUI.Out).ToNot(Say("Are you sure you want to continue?"))

					Expect(testUI.Out).To(Say("Rolling back to revision 1 for app some-app in org some-org / space some-space as steve..."))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
					Expect(testUI.Err).To(Say("warning-3"))

					Expect(testUI.Out).To(Say("OK"))
				})
			})

			When("user says yes to prompt", func() {
				BeforeEach(func() {
					_, err := input.Write([]byte("y\n"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("successfully executes the command and outputs warnings", func() {
					Expect(fakeAppStager.StartAppCallCount()).To(Equal(1), "GetStartApp call count")

					application, _, _, revisionGUID, opts := fakeAppStager.StartAppArgsForCall(0)
					Expect(application.GUID).To(Equal("123"))
					Expect(revisionGUID).To(Equal("some-1-guid"))
					Expect(opts.AppAction).To(Equal(constant.ApplicationRollingBack))
					Expect(opts.Strategy).To(Equal(constant.DeploymentStrategyRolling))
					Expect(opts.MaxInFlight).To(Equal(5))

					Expect(testUI.Out).To(Say("Rolling '%s' back to revision '1' will create a new revision. The new revision will use the settings from revision '1'.", appName))
					Expect(testUI.Out).To(Say("Are you sure you want to continue?"))
					Expect(testUI.Out).To(Say("Rolling back to revision 1 for app some-app in org some-org / space some-space as steve..."))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
					Expect(testUI.Err).To(Say("warning-3"))

					Expect(testUI.Out).To(Say("OK"))
				})
			})

			When("the strategy flag is passed", func() {
				BeforeEach(func() {
					cmd.Strategy.Name = constant.DeploymentStrategyCanary

					_, err := input.Write([]byte("y\n"))
					Expect(err).NotTo(HaveOccurred())
				})
				It("uses the specified strategy to rollback", func() {
					Expect(fakeAppStager.StartAppCallCount()).To(Equal(1), "GetStartApp call count")

					application, _, _, revisionGUID, opts := fakeAppStager.StartAppArgsForCall(0)
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(application.GUID).To(Equal("123"))
					Expect(revisionGUID).To(Equal("some-1-guid"))
					Expect(opts.AppAction).To(Equal(constant.ApplicationRollingBack))
					Expect(opts.Strategy).To(Equal(constant.DeploymentStrategyCanary))
				})
			})

			When("canary strategy is provided", func() {
				BeforeEach(func() {
					cmd.Strategy = flag.DeploymentStrategy{Name: constant.DeploymentStrategyCanary}
					cmd.InstanceSteps = "1,2,4"

					fakeConfig = &commandfakes.FakeConfig{}
					fakeConfig.APIVersionReturns("4.0.0")
					cmd.Config = fakeConfig

					_, err := input.Write([]byte("y\n"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("starts the app with the current droplet", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeAppStager.StartAppCallCount()).To(Equal(1))

					inputApp, inputSpace, inputOrg, inputDropletGuid, opts := fakeAppStager.StartAppArgsForCall(0)
					Expect(inputApp).To(Equal(app))
					Expect(inputDropletGuid).To(Equal("some-1-guid"))
					Expect(inputSpace).To(Equal(cmd.Config.TargetedSpace()))
					Expect(inputOrg).To(Equal(cmd.Config.TargetedOrganization()))
					Expect(opts.Strategy).To(Equal(constant.DeploymentStrategyCanary))
					Expect(opts.AppAction).To(Equal(constant.ApplicationRollingBack))
					Expect(opts.CanarySteps).To(Equal([]resources.CanaryStep{{InstanceWeight: 1}, {InstanceWeight: 2}, {InstanceWeight: 4}}))
				})
			})

			When("user says no to prompt", func() {
				BeforeEach(func() {
					_, err := input.Write([]byte("n\n"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("does not execute the command and outputs warnings", func() {
					Expect(fakeAppStager.StartAppCallCount()).To(Equal(0), "GetStartApp call count")

					Expect(testUI.Out).To(Say("Rolling '%s' back to revision '1' will create a new revision. The new revision will use the settings from revision '1'.", appName))
					Expect(testUI.Out).To(Say("App '%s' has not been rolled back to revision '1'.", appName))

					Expect(testUI.Err).To(Say("app-warning-1"))
					Expect(testUI.Err).To(Say("revision-warning-3"))
				})
			})

			When("the user chooses the default", func() {
				BeforeEach(func() {
					_, err := input.Write([]byte("\n"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("cancels the rollback", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(fakeAppStager.StartAppCallCount()).To(Equal(0), "GetStartApp call count")

					Expect(testUI.Out).To(Say("App '%s' has not been rolled back to revision '1'.", appName))
				})
			})
		})
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

		Entry("max-in-flight is smaller than 1",
			func() {
				maxInFlight := 0
				cmd.MaxInFlight = &maxInFlight
			},
			translatableerror.IncorrectUsageError{
				Message: "--max-in-flight must be greater than or equal to 1",
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
