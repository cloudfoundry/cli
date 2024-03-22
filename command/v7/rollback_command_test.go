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
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("rollback Command", func() {
	var (
		app             string
		binaryName      string
		executeErr      error
		fakeActor       *v7fakes.FakeActor
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		input           *Buffer
		testUI          *ui.UI

		fakeAppStager *sharedfakes.FakeAppStager
		cmd           v7.RollbackCommand
	)

	BeforeEach(func() {
		app = "some-app"
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
			RequiredArgs: flag.AppName{AppName: app},
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				Actor:       fakeActor,
				SharedActor: fakeSharedActor,
			},
			Stager: fakeAppStager,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("displays the experimental warning", func() {
		Expect(testUI.Err).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
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
					resources.Application{GUID: "123"},
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
				Expect(appName).To(Equal(app))
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

					application, revisionGUID, _, _, _, _, appAction := fakeAppStager.StartAppArgsForCall(0)
					Expect(application.GUID).To(Equal("123"))
					Expect(revisionGUID).To(Equal("some-1-guid"))
					Expect(appAction).To(Equal(constant.ApplicationRollingBack))

					Expect(testUI.Out).ToNot(Say("Rolling '%s' back to revision '1' will create a new revision. The new revision '3' will use the settings from revision '1'.", app))
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

					application, revisionGUID, _, _, _, _, appAction := fakeAppStager.StartAppArgsForCall(0)
					Expect(application.GUID).To(Equal("123"))
					Expect(revisionGUID).To(Equal("some-1-guid"))
					Expect(appAction).To(Equal(constant.ApplicationRollingBack))

					Expect(testUI.Out).To(Say("Rolling '%s' back to revision '1' will create a new revision. The new revision will use the settings from revision '1'.", app))
					Expect(testUI.Out).To(Say("Are you sure you want to continue?"))
					Expect(testUI.Out).To(Say("Rolling back to revision 1 for app some-app in org some-org / space some-space as steve..."))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
					Expect(testUI.Err).To(Say("warning-3"))

					Expect(testUI.Out).To(Say("OK"))
				})
			})

			When("user says no to prompt", func() {
				BeforeEach(func() {
					_, err := input.Write([]byte("n\n"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("does not execute the command and outputs warnings", func() {
					Expect(fakeAppStager.StartAppCallCount()).To(Equal(0), "GetStartApp call count")

					Expect(testUI.Out).To(Say("Rolling '%s' back to revision '1' will create a new revision. The new revision will use the settings from revision '1'.", app))
					Expect(testUI.Out).To(Say("App '%s' has not been rolled back to revision '1'.", app))

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

					Expect(testUI.Out).To(Say("App '%s' has not been rolled back to revision '1'.", app))
				})
			})
		})
	})
})
