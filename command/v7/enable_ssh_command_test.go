package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("enable-ssh Command", func() {
	var (
		cmd                EnableSSHCommand
		testUI             *ui.UI
		fakeConfig         *commandfakes.FakeConfig
		fakeSharedActor    *commandfakes.FakeSharedActor
		fakeUpdateSSHActor *v7fakes.FakeUpdateSSHActor

		binaryName  string
		currentUserName string
		executeErr  error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeUpdateSSHActor = new(v7fakes.FakeUpdateSSHActor)

		cmd = EnableSSHCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeUpdateSSHActor,
		}

		cmd.RequiredArgs.AppName = "some-app"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		currentUserName = "some-user"
		fakeConfig.CurrentUserNameReturns(currentUserName, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: "faceman"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is logged in", func() {
		When("no errors occur", func() {
			BeforeEach(func() {
				fakeUpdateSSHActor.GetApplicationByNameAndSpaceReturns(
					v7action.Application{Name: "some-app-name"},
					v7action.Warnings{"some-warning"},
					nil,
				)
				fakeUpdateSSHActor.UpdateSSHReturns(v7action.Warnings{"some-warnings"}, nil)
			})

			It("enables ssh on the app", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeUpdateSSHActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))

				appName, spaceGUID := fakeUpdateSSHActor.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(appName).To(Equal(cmd.RequiredArgs.AppName))
				Expect(spaceGUID).To(Equal(cmd.Config.TargetedSpace().GUID))

				Expect(fakeUpdateSSHActor.UpdateSSHCallCount()).To(Equal(1))
				app, enabled := fakeUpdateSSHActor.UpdateSSHArgsForCall(0)
				Expect(app.Name).To(Equal("some-app-name"))
				Expect(enabled).To(Equal(true))

				Expect(testUI.Err).To(Say("some-warnings"))
				Expect(testUI.Out).To(Say(`Enabling ssh support for app %s as %s\.\.\.`, appName, currentUserName))
				Expect(testUI.Out).To(Say("OK"))
			})

		})

		When("an error occurs", func() {
			When("GetApp action errors", func() {
				When("no user is found", func() {
					var returnedErr error

					BeforeEach(func() {
						returnedErr = actionerror.ApplicationNotFoundError{Name: "some-app"}
						fakeUpdateSSHActor.GetApplicationByNameAndSpaceReturns(
							v7action.Application{},
							nil,
							returnedErr)
					})

					It("returns the same error", func() {
						Expect(executeErr).To(HaveOccurred()) // TODO: verify that the integration level tests for FAILED
						Expect(testUI.Out).To(Say(`App 'some-app' not found.`))
					})
				})
			})

			When("Enable ssh action errors", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = errors.New("some-error")
					fakeUpdateSSHActor.GetApplicationByNameAndSpaceReturns(
						v7action.Application{Name: "some-app-name"},
						v7action.Warnings{"some-warning"},
						nil,
					)
					fakeUpdateSSHActor.UpdateSSHReturns(nil, returnedErr)
				})

				It("returns the same error", func() {
					Expect(executeErr).To(MatchError(returnedErr))
				})
			})
		})
	})
})
