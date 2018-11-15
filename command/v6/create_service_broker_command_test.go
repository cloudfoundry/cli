package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-service-broker Command", func() {
	var (
		cmd             CreateServiceBrokerCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeCreateServiceBrokerActor
		binaryName      string
		executeErr      error
		extraArgs       []string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeCreateServiceBrokerActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns("faceman")
		extraArgs = nil
		cmd = CreateServiceBrokerCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			RequiredArgs: flag.ServiceBrokerArgs{
				ServiceBroker: "cool-broker",
				Username:      "admin",
				Password:      "password",
				URL:           "https://broker.com",
			},
			SpaceScoped: false,
		}

	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(extraArgs)
	})

	When("the user provides arguments", func() {
		BeforeEach(func() {
			extraArgs = []string{"some-extra-arg"}
		})

		It("fails with a TooManyArgumentsError", func() {
			Expect(executeErr).To(MatchError(translatableerror.TooManyArgumentsError{
				ExtraArgument: "some-extra-arg",
			}))
		})
	})

	When("the user does not have an org and space targeted", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(nil)
			fakeConfig.CurrentUserReturns(
				configv3.User{Name: "some-user"},
				nil)
		})

		When("fetching the current user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("no user"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("no user"))
			})
		})

		Context("not space scoped", func() {
			When("creating a service broker is successful", func() {
				BeforeEach(func() {
					fakeActor.CreateServiceBrokerReturns(v2action.ServiceBroker{}, []string{"a-warning", "another-warning"}, nil)
				})

				It("displays a message indicating that it is creating the service broker", func() {
					Expect(fakeActor.CreateServiceBrokerCallCount()).To(Equal(1))
					serviceBroker, username, password, url, spaceGUID := fakeActor.CreateServiceBrokerArgsForCall(0)
					Expect(serviceBroker).To(Equal("cool-broker"))
					Expect(username).To(Equal("admin"))
					Expect(password).To(Equal("password"))
					Expect(url).To(Equal("https://broker.com"))
					Expect(spaceGUID).To(Equal(""))

					Expect(testUI.Out).To(Say("Creating service broker cool-broker as some-user..."))
					Expect(testUI.Out).To(Say("OK"))
					Expect(executeErr).NotTo(HaveOccurred())
				})

				It("displays all warnings", func() {
					Expect(testUI.Err).To(Say("a-warning"))
					Expect(testUI.Err).To(Say("another-warning"))
				})
			})

			When("creating a service broker is unsuccessful", func() {
				BeforeEach(func() {
					fakeActor.CreateServiceBrokerReturns(v2action.ServiceBroker{}, []string{"a-warning", "another-warning"}, errors.New("invalid-broker-name"))
				})

				It("prints the error and warnings returned by Cloud Controller", func() {
					Expect(fakeActor.CreateServiceBrokerCallCount()).To(Equal(1))

					Expect(testUI.Out).To(Say("Creating service broker cool-broker as some-user..."))
					Expect(testUI.Err).To(Say("a-warning"))
					Expect(testUI.Err).To(Say("another-warning"))
					Expect(executeErr).To(MatchError("invalid-broker-name"))
				})
			})
		})

		Context("space scoped", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(actionerror.NoSpaceTargetedError{BinaryName: binaryName})
				cmd.SpaceScoped = true
			})

			It("returns an error", func() {
				Expect(fakeActor.CreateServiceBrokerCallCount()).To(Equal(0))
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				Expect(executeErr).To(MatchError(actionerror.NoSpaceTargetedError{BinaryName: binaryName}))
			})
		})
	})

	When("the user is targeting an org and space", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(nil)
			fakeConfig.CurrentUserReturns(
				configv3.User{Name: "some-user"},
				nil)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space",
			})
			fakeActor.CreateServiceBrokerReturns(v2action.ServiceBroker{}, []string{"a-warning", "another-warning"}, nil)
		})

		Context("space scoped", func() {
			BeforeEach(func() {
				cmd.SpaceScoped = true
			})

			When("all arguments are passed and are valid", func() {
				It("registers a space scoped service broker", func() {
					Expect(fakeActor.CreateServiceBrokerCallCount()).To(Equal(1))
					serviceBroker, username, password, url, spaceGUID := fakeActor.CreateServiceBrokerArgsForCall(0)
					Expect(serviceBroker).To(Equal("cool-broker"))
					Expect(username).To(Equal("admin"))
					Expect(password).To(Equal("password"))
					Expect(url).To(Equal("https://broker.com"))
					Expect(spaceGUID).To(Equal("some-space-guid"))

					Expect(testUI.Out).To(Say("Creating service broker cool-broker in org some-org / space some-space as some-user..."))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Err).To(Say("a-warning"))
					Expect(testUI.Err).To(Say("another-warning"))
					Expect(executeErr).NotTo(HaveOccurred())
				})
			})
		})

		Context("not space scoped", func() {
			When("all arguments are passed and are valid", func() {
				It("registers a service broker", func() {
					Expect(fakeActor.CreateServiceBrokerCallCount()).To(Equal(1))
					Expect(testUI.Out).To(Say("Creating service broker cool-broker as some-user..."))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Err).To(Say("a-warning"))
					Expect(testUI.Err).To(Say("another-warning"))
					Expect(executeErr).NotTo(HaveOccurred())
				})
			})
		})
	})
})
