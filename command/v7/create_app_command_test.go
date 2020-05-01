package v7_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-app Command", func() {
	var (
		cmd             v7.CreateAppCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		app             string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"

		cmd = v7.CreateAppCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequiredArgs: flag.AppName{AppName: app},
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("displays the experimental warning", func() {
		Expect(testUI.Err).NotTo(Say("This command is in EXPERIMENTAL stage and may change without notice"))
	})

	When("checking target fails", func() {
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

	When("the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
		})

		When("the create is successful", func() {
			BeforeEach(func() {
				fakeActor.CreateApplicationInSpaceReturns(resources.Application{}, v7action.Warnings{"I am a warning", "I am also a warning"}, nil)
			})

			It("displays the header and ok", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Creating app some-app in org some-org / space some-space as banana..."))
				Expect(testUI.Out).To(Say("OK"))

				Expect(testUI.Err).To(Say("I am a warning"))
				Expect(testUI.Err).To(Say("I am also a warning"))

				Expect(fakeActor.CreateApplicationInSpaceCallCount()).To(Equal(1))

				createApp, createSpaceGUID := fakeActor.CreateApplicationInSpaceArgsForCall(0)
				Expect(createApp).To(Equal(resources.Application{
					Name: app,
				}))
				Expect(createSpaceGUID).To(Equal("some-space-guid"))
			})

			When("app type is specified", func() {
				BeforeEach(func() {
					cmd.AppType = "docker"
				})

				It("creates an app with specified app type", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeActor.CreateApplicationInSpaceCallCount()).To(Equal(1))

					createApp, createSpaceGUID := fakeActor.CreateApplicationInSpaceArgsForCall(0)
					Expect(createApp).To(Equal(resources.Application{
						Name:          app,
						LifecycleType: constant.AppLifecycleTypeDocker,
					}))
					Expect(createSpaceGUID).To(Equal("some-space-guid"))
				})
			})
		})

		When("the create is unsuccessful", func() {
			Context("due to an unexpected error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("I am an error")
					fakeActor.CreateApplicationInSpaceReturns(resources.Application{}, v7action.Warnings{"I am a warning", "I am also a warning"}, expectedErr)
				})

				It("displays the header and error", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Out).To(Say("Creating app some-app in org some-org / space some-space as banana..."))

					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
				})
			})

			Context("due to NameNotUniqueInSpaceError{}", func() {
				BeforeEach(func() {
					fakeActor.CreateApplicationInSpaceReturns(
						resources.Application{},
						v7action.Warnings{"I am a warning", "I am also a warning"},
						ccerror.NameNotUniqueInSpaceError{
							UnprocessableEntityError: ccerror.UnprocessableEntityError{
								Message: fmt.Sprintf("Application '%s' already exists.", app),
							},
						})
				})

				It("displays the header and ok", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Creating app some-app in org some-org / space some-space as banana..."))
					Expect(testUI.Out).To(Say("Application '%s' already exists.", app))
					Expect(testUI.Out).To(Say("OK"))

					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
				})
			})
		})
	})
})
