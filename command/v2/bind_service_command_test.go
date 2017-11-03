package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("bind-service Command", func() {
	var (
		cmd             BindServiceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeBindServiceActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeBindServiceActor)

		cmd = BindServiceCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.AppName = "some-app"
		cmd.RequiredArgs.ServiceInstanceName = "some-service"
		cmd.ParametersAsJSON = map[string]interface{}{
			"some-parameter": "some-value",
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns("faceman")
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when a cloud controller API endpoint is set", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("some-url")
		})

		Context("when checking target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeTrue())
				Expect(checkTargetedSpace).To(BeTrue())
			})
		})

		Context("when the user is logged in, and an org and space are targeted", func() {
			BeforeEach(func() {
				fakeConfig.HasTargetedOrganizationReturns(true)
				fakeConfig.HasTargetedSpaceReturns(true)
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{
					GUID: "some-org-guid",
					Name: "some-org",
				})
				fakeConfig.TargetedSpaceReturns(configv3.Space{
					GUID: "some-space-guid",
					Name: "some-space",
				})
			})

			Context("when getting the current user returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("got bananapants??")
					fakeConfig.CurrentUserReturns(
						configv3.User{},
						expectedErr)
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError(expectedErr))
				})
			})

			Context("when getting the current user does not return an error", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserReturns(
						configv3.User{Name: "some-user"},
						nil)
				})

				It("displays flavor text", func() {
					Expect(testUI.Out).To(Say("Binding service some-service to app some-app in org some-org / space some-space as some-user..."))

					Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))
				})

				Context("when the service was already bound", func() {
					BeforeEach(func() {
						fakeActor.BindServiceBySpaceReturns(
							[]string{"foo", "bar"},
							ccerror.ServiceBindingTakenError{})
					})

					It("displays warnings and 'OK'", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(testUI.Err).To(Say("foo"))
						Expect(testUI.Err).To(Say("bar"))
						Expect(testUI.Out).To(Say("App some-app is already bound to some-service."))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				Context("when binding the service instance results in an error other than ServiceBindingTakenError", func() {
					BeforeEach(func() {
						fakeActor.BindServiceBySpaceReturns(
							nil,
							actionerror.ApplicationNotFoundError{Name: "some-app"})
					})

					It("should return the error", func() {
						Expect(executeErr).To(MatchError(actionerror.ApplicationNotFoundError{
							Name: "some-app",
						}))
					})
				})

				Context("when the service binding is successful", func() {
					BeforeEach(func() {
						fakeActor.BindServiceBySpaceReturns(
							v2action.Warnings{"some-warning", "another-warning"},
							nil,
						)
					})

					It("displays OK and the TIP", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).To(Say("TIP: Use 'faceman restage some-app' to ensure your env variable changes take effect"))
						Expect(testUI.Err).To(Say("some-warning"))
						Expect(testUI.Err).To(Say("another-warning"))

						Expect(fakeActor.BindServiceBySpaceCallCount()).To(Equal(1))
						appName, serviceInstanceName, spaceGUID, parameters := fakeActor.BindServiceBySpaceArgsForCall(0)
						Expect(appName).To(Equal("some-app"))
						Expect(serviceInstanceName).To(Equal("some-service"))
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(parameters).To(Equal(map[string]interface{}{"some-parameter": "some-value"}))
					})
				})
			})
		})
	})
})
