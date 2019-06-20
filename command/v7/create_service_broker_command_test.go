package v7_test

import (
	"errors"

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

var _ = Describe("create-service-broker Command", func() {
	var (
		cmd             *v7.CreateServiceBrokerCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeCreateServiceBrokerActor
		input           *Buffer
		binaryName      string
		executeErr      error

		args = flag.ServiceBrokerArgs{
			ServiceBroker: "service-broker-name",
			Username:      "username",
			Password:      "password",
			URL:           "https://example.org/super-broker",
		}
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeCreateServiceBrokerActor)
		fakeActor.CreateServiceBrokerReturns(v7action.Warnings{"some default warning"}, nil)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = &v7.CreateServiceBrokerCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,

			RequiredArgs: args,
		}
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
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
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
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
		})

		It("displays a message with the username", func() {
			Expect(testUI.Out).To(Say("Creating service broker %s as %s...", args.ServiceBroker, "steve"))
		})

		It("calls the CreateServiceBrokerActor", func() {
			Expect(fakeActor.CreateServiceBrokerCallCount()).To(Equal(1))
			credentials := fakeActor.CreateServiceBrokerArgsForCall(0)
			Expect(credentials).To(Equal(v7action.ServiceBroker{
				Name: "service-broker-name",
				URL:  "https://example.org/super-broker",
				Credentials: v7action.ServiceBrokerCredentials{
					Type: constant.BasicCredentials,
					Data: v7action.ServiceBrokerCredentialsData{
						Username: "username",
						Password: "password",
					},
				},
			}))
		})

		It("displays the warnings", func() {
			Expect(testUI.Err).To(Say("some default warning"))
		})

		It("displays OK", func() {
			Expect(testUI.Out).To(Say("OK"))
		})

		When("calling the CreateServiceBrokerActor returns an error", func() {
			BeforeEach(func() {
				fakeActor.CreateServiceBrokerReturns(v7action.Warnings{"service-broker-warnings"}, errors.New("fake create-service-broker error"))
			})

			It("prints the error and warnings", func() {
				Expect(testUI.Out).NotTo(Say("OK"))
				Expect(executeErr).To(MatchError("fake create-service-broker error"))
				Expect(testUI.Err).To(Say("service-broker-warnings"))
			})
		})
	})
})
