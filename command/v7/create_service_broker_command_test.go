package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-service-broker Command", func() {
	var (
		cmd             *v7.CreateServiceBrokerCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		input           *Buffer
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		fakeActor.CreateServiceBrokerReturns(v7action.Warnings{"some default warning"}, nil)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = &v7.CreateServiceBrokerCommand{
			BaseCommand: command.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		setPositionalFlags(cmd, "service-broker-name", "username", "password", "https://example.org/super-broker")
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("an error occurred"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("an error occurred"))
		})
	})

	When("fetching the current user fails", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("an error occurred"))
		})

		It("return an error", func() {
			Expect(executeErr).To(MatchError("an error occurred"))
		})
	})

	When("fetching the current user succeeds", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
		})

		It("checks that there is a valid target", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})

		It("displays a message with the username", func() {
			Expect(testUI.Out).To(Say(`Creating service broker %s as %s\.\.\.`, "service-broker-name", "steve"))
		})

		It("passes the data to the actor layer", func() {
			Expect(fakeActor.CreateServiceBrokerCallCount()).To(Equal(1))

			model := fakeActor.CreateServiceBrokerArgsForCall(0)

			Expect(model.Name).To(Equal("service-broker-name"))
			Expect(model.Username).To(Equal("username"))
			Expect(model.Password).To(Equal("password"))
			Expect(model.URL).To(Equal("https://example.org/super-broker"))
			Expect(model.SpaceGUID).To(Equal(""))
		})

		It("displays the warnings", func() {
			Expect(testUI.Err).To(Say("some default warning"))
		})

		It("displays OK", func() {
			Expect(testUI.Out).To(Say("OK"))
		})

		When("the actor returns an error", func() {
			BeforeEach(func() {
				fakeActor.CreateServiceBrokerReturns(v7action.Warnings{"service-broker-warnings"}, errors.New("fake create-service-broker error"))
			})

			It("prints the error and warnings", func() {
				Expect(testUI.Out).NotTo(Say("OK"))
				Expect(executeErr).To(MatchError("fake create-service-broker error"))
				Expect(testUI.Err).To(Say("service-broker-warnings"))
			})
		})

		When("creating a space scoped broker", func() {
			BeforeEach(func() {
				cmd.SpaceScoped = true
				fakeConfig.TargetedSpaceReturns(configv3.Space{
					Name: "fake-space-name",
					GUID: "fake-space-guid",
				})
				fakeConfig.TargetedOrganizationNameReturns("fake-org-name")
			})

			It("checks that a space is targeted", func() {
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeTrue())
				Expect(checkTargetedSpace).To(BeTrue())
			})

			It("displays the space name in the message", func() {
				Expect(testUI.Out).To(Say(`Creating service broker %s in org %s / space %s as %s\.\.\.`, "service-broker-name", "fake-org-name", "fake-space-name", "steve"))
			})

			It("looks up the space guid and passes it to the actor", func() {
				Expect(fakeActor.CreateServiceBrokerCallCount()).To(Equal(1))

				model := fakeActor.CreateServiceBrokerArgsForCall(0)
				Expect(model.SpaceGUID).To(Equal("fake-space-guid"))
			})
		})
	})
})
