package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("unbind-service Command", func() {
	var (
		cmd             UnbindServiceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeUnbindServiceActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeUnbindServiceActor)

		cmd = UnbindServiceCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.AppName = "some-app"
		cmd.RequiredArgs.ServiceInstanceName = "some-service"

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
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Unbinding app some-app from service some-service in org some-org / space some-space as some-user..."))
				})

				Context("when unbinding the service instance results in an error not related to service binding", func() {
					BeforeEach(func() {
						fakeActor.UnbindServiceBySpaceReturns(nil, actionerror.ApplicationNotFoundError{Name: "some-app"})
					})

					It("should return the error", func() {
						Expect(executeErr).To(MatchError(actionerror.ApplicationNotFoundError{
							Name: "some-app",
						}))
					})
				})

				Context("when the service binding does not exist", func() {
					BeforeEach(func() {
						fakeActor.UnbindServiceBySpaceReturns(
							[]string{"foo", "bar"},
							actionerror.ServiceBindingNotFoundError{})
					})

					It("displays warnings and 'OK'", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(testUI.Err).To(Say("foo"))
						Expect(testUI.Err).To(Say("bar"))
						Expect(testUI.Err).To(Say("Binding between some-service and some-app did not exist"))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				Context("when the service binding exists", func() {
					It("displays OK", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Err).NotTo(Say("Binding between some-service and some-app did not exist"))

						Expect(fakeActor.UnbindServiceBySpaceCallCount()).To(Equal(1))
						appName, serviceInstanceName, spaceGUID := fakeActor.UnbindServiceBySpaceArgsForCall(0)
						Expect(appName).To(Equal("some-app"))
						Expect(serviceInstanceName).To(Equal("some-service"))
						Expect(spaceGUID).To(Equal("some-space-guid"))
					})
				})
			})
		})
	})
})
