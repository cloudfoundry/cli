package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/commands/commandsfakes"
	. "code.cloudfoundry.org/cli/commands/v2"
	"code.cloudfoundry.org/cli/commands/v2/common"
	"code.cloudfoundry.org/cli/commands/v2/v2fakes"
	"code.cloudfoundry.org/cli/utils/configv3"
	"code.cloudfoundry.org/cli/utils/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Unbind Service Command", func() {
	var (
		cmd        UnbindServiceCommand
		fakeUI     *ui.UI
		fakeActor  *v2fakes.FakeUnbindServiceActor
		fakeConfig *commandsfakes.FakeConfig
		executeErr error
	)

	BeforeEach(func() {
		out := NewBuffer()
		fakeUI = ui.NewTestUI(nil, out, out)
		fakeActor = new(v2fakes.FakeUnbindServiceActor)
		fakeConfig = new(commandsfakes.FakeConfig)
		fakeConfig.ExperimentalReturns(true)

		cmd = UnbindServiceCommand{
			UI:     fakeUI,
			Actor:  fakeActor,
			Config: fakeConfig,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("Displays the experimental warning message", func() {
		Expect(fakeUI.Out).To(Say(ExperimentalWarning))
	})

	Context("when checking that the api endpoint is set, the user is logged in, and an org and space are targeted", func() {
		BeforeEach(func() {
			fakeConfig.BinaryNameReturns("faceman")
		})

		It("returns an error if the check fails", func() {
			Expect(executeErr).To(MatchError(common.NotLoggedInError{
				BinaryName: "faceman",
			}))
		})
	})

	Context("when the api endpoint is set, the user is logged in, and an org and space are targeted", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("some-url")
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space",
			})

			cmd.RequiredArgs.AppName = "some-app"
			cmd.RequiredArgs.ServiceInstanceName = "some-service"
		})

		Context("when getting the logged in user results in an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("got bananapants??")
				fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
			})

			It("returns the same error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		Context("when getting the logged in user does not result in an error", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{
					Name: "some-user",
				}, nil)
			})

			It("displays flavor text", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeUI.Out).To(Say("Unbinding app some-app from service some-service in org some-org / space some-space as some-user..."))
			})

			Context("when unbinding the service instance results in an error not related to service binding", func() {
				BeforeEach(func() {
					fakeActor.UnbindServiceBySpaceReturns(nil, v2actions.ApplicationNotFoundError{
						Name: "some-app",
					})
				})

				It("should return the error", func() {
					Expect(executeErr).To(MatchError(common.ApplicationNotFoundError{
						Name: "some-app",
					}))
				})
			})

			Context("when the service binding does not exist", func() {
				BeforeEach(func() {
					fakeActor.UnbindServiceBySpaceReturns([]string{"foo", "bar"}, v2actions.ServiceBindingNotFoundError{})
				})

				It("displays warnings and 'OK'", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(fakeUI.Err).To(Say("foo"))
					Expect(fakeUI.Err).To(Say("bar"))
					Expect(fakeUI.Err).To(Say("Binding between some-service and some-app did not exist"))
					Expect(fakeUI.Out).To(Say("OK"))
				})
			})

			Context("when the service binding exists", func() {
				It("displays OK", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeUI.Out).To(Say("OK"))
					Expect(fakeUI.Err).NotTo(Say("Binding between some-service and some-app did not exist"))

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
