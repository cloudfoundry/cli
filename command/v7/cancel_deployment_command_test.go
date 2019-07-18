package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Cancel deployment command", func() {
	var (
		cmd             CancelDeploymentCommand
		testUI          *ui.UI
		input           *Buffer
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeCancelDeploymentActor
		binaryName      string
		appName         string
		spaceGUID       string
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeCancelDeploymentActor)

		binaryName = "clodFoundry"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = CancelDeploymentCommand{
			RequiredArgs: flag.AppName{AppName: appName},
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
		}

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
			GUID: "some-org-guid",
		})

		spaceGUID = "some-space-guid"
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: spaceGUID,
		})

		fakeConfig.CurrentUserReturns(configv3.User{Name: "timmyD"}, nil)
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

	When("the user is logged in", func() {
		It("delegates to actor.GetApplicationByNameAndSpace", func() {
			Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
			actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
			Expect(actualAppName).To(Equal(appName))
			Expect(actualSpaceGUID).To(Equal(spaceGUID))
		})

		When("getting the app fails", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v7action.Application{},
					v7action.Warnings{"get-app-warning"},
					errors.New("get-app-error"),
				)
			})

			It("returns the errors and outputs warnings", func() {
				Expect(executeErr).To(MatchError("get-app-error"))
				Expect(testUI.Err).To(Say("get-app-warning"))

				Expect(fakeActor.GetLatestActiveDeploymentForAppCallCount()).To(Equal(0))
				Expect(fakeActor.CancelDeploymentCallCount()).To(Equal(0))
			})
		})

		When("getting the app succeeds", func() {
			var appGUID string
			BeforeEach(func() {
				appGUID = "some-app-guid"
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v7action.Application{Name: appName, GUID: appGUID},
					v7action.Warnings{"get-app-warning"},
					nil,
				)
			})

			It("delegates to actor.GetLatestDeployment", func() {
				Expect(fakeActor.GetLatestActiveDeploymentForAppCallCount()).To(Equal(1))
				Expect(fakeActor.GetLatestActiveDeploymentForAppArgsForCall(0)).To(Equal(appGUID))
			})

			When("getting the latest deployment fails", func() {
				BeforeEach(func() {
					fakeActor.GetLatestActiveDeploymentForAppReturns(
						v7action.Deployment{},
						v7action.Warnings{"get-deployment-warning"},
						errors.New("get-deployment-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(MatchError("get-deployment-error"))
					Expect(testUI.Err).To(Say("get-app-warning"))
					Expect(testUI.Err).To(Say("get-deployment-warning"))

					Expect(fakeActor.CancelDeploymentCallCount()).To(Equal(0))
				})
			})

			When("getting the latest deployment succeeds", func() {
				var deploymentGUID string
				BeforeEach(func() {
					deploymentGUID = "some-deployment-guid"
					fakeActor.GetLatestActiveDeploymentForAppReturns(
						v7action.Deployment{GUID: deploymentGUID},
						v7action.Warnings{"get-deployment-warning"},
						nil,
					)
				})

				It("delegates to actor.CancelDeployment", func() {
					Expect(fakeActor.CancelDeploymentCallCount()).To(Equal(1))
					Expect(fakeActor.CancelDeploymentArgsForCall(0)).To(Equal(deploymentGUID))
				})

				When("cancelling the deployment fails", func() {
					BeforeEach(func() {
						fakeActor.CancelDeploymentReturns(
							v7action.Warnings{"cancel-deployment-warning"},
							errors.New("cancel-deployment-error"),
						)
					})

					It("returns all warnings and errors", func() {
						Expect(executeErr).To(MatchError("cancel-deployment-error"))
						Expect(testUI.Err).To(Say("get-app-warning"))
						Expect(testUI.Err).To(Say("get-deployment-warning"))
						Expect(testUI.Err).To(Say("cancel-deployment-warning"))
					})
				})

				When("cancelling the deployment succeeds", func() {
					BeforeEach(func() {
						fakeActor.CancelDeploymentReturns(
							v7action.Warnings{"cancel-deployment-warning"},
							nil,
						)
					})

					It("returns warnings and success", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Err).To(Say("get-app-warning"))
						Expect(testUI.Err).To(Say("get-deployment-warning"))
						Expect(testUI.Err).To(Say("cancel-deployment-warning"))
					})
				})
			})
		})
	})
})
