package v7_test

import (
	"errors"
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-service-broker Command", func() {
	const (
		binaryName        = "cf-command"
		user              = "steve"
		serviceBrokerName = "fake-service-broker-name"
		username          = "fake-username"
		password          = "fake-password"
		url               = "fake-url"
	)

	var (
		cmd             *v7.CreateServiceBrokerCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		input           *Buffer
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		fakeActor.CreateServiceBrokerReturns(v7action.Warnings{"some default warning"}, nil)

		fakeConfig.BinaryNameReturns(binaryName)

		cmd = &v7.CreateServiceBrokerCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}
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
			fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("an error occurred"))
			setPositionalFlags(cmd, serviceBrokerName, username, password, url)
		})

		It("return an error", func() {
			Expect(executeErr).To(MatchError("an error occurred"))
		})
	})

	When("fetching the current user succeeds", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{Name: user}, nil)
			setPositionalFlags(cmd, serviceBrokerName, username, password, url)
		})

		It("checks that there is a valid target", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})

		It("displays a message with the username", func() {
			Expect(testUI.Out).To(Say(`Creating service broker %s as %s\.\.\.`, serviceBrokerName, user))
		})

		It("passes the data to the actor layer", func() {
			Expect(fakeActor.CreateServiceBrokerCallCount()).To(Equal(1))

			model := fakeActor.CreateServiceBrokerArgsForCall(0)

			Expect(model.Name).To(Equal(serviceBrokerName))
			Expect(model.Username).To(Equal(username))
			Expect(model.Password).To(Equal(password))
			Expect(model.URL).To(Equal(url))
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
			const (
				orgName   = "fake-org-name"
				spaceName = "fake-space-name"
				spaceGUID = "fake-space-guid"
			)

			BeforeEach(func() {
				cmd.SpaceScoped = true
				fakeConfig.TargetedSpaceReturns(configv3.Space{
					Name: spaceName,
					GUID: spaceGUID,
				})
				fakeConfig.TargetedOrganizationNameReturns(orgName)
			})

			It("checks that a space is targeted", func() {
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeTrue())
				Expect(checkTargetedSpace).To(BeTrue())
			})

			It("displays the space name in the message", func() {
				Expect(testUI.Out).To(Say(`Creating service broker %s in org %s / space %s as %s\.\.\.`, serviceBrokerName, orgName, spaceName, user))
			})

			It("looks up the space guid and passes it to the actor", func() {
				Expect(fakeActor.CreateServiceBrokerCallCount()).To(Equal(1))

				model := fakeActor.CreateServiceBrokerArgsForCall(0)
				Expect(model.SpaceGUID).To(Equal(spaceGUID))
			})
		})
	})

	When("password is provided as environment variable", func() {
		const (
			varName     = "CF_BROKER_PASSWORD"
			varPassword = "var-password"
		)

		BeforeEach(func() {
			setPositionalFlags(cmd, serviceBrokerName, username, url, "")
			os.Setenv(varName, varPassword)
		})

		AfterEach(func() {
			os.Unsetenv(varName)
		})

		It("passes the data to the actor layer", func() {
			Expect(fakeActor.CreateServiceBrokerCallCount()).To(Equal(1))

			model := fakeActor.CreateServiceBrokerArgsForCall(0)

			Expect(model.Name).To(Equal(serviceBrokerName))
			Expect(model.Username).To(Equal(username))
			Expect(model.Password).To(Equal(varPassword))
			Expect(model.URL).To(Equal(url))
			Expect(model.SpaceGUID).To(Equal(""))
		})
	})

	When("password is provided via prompt", func() {
		const promptPassword = "prompt-password"

		BeforeEach(func() {
			setPositionalFlags(cmd, serviceBrokerName, username, url, "")

			_, err := input.Write([]byte(fmt.Sprintf("%s\n", promptPassword)))
			Expect(err).NotTo(HaveOccurred())
		})

		It("prompts the user for credentials", func() {
			Expect(testUI.Out).To(Say("Service Broker Password: "))
		})

		It("does not echo the credentials", func() {
			Expect(testUI.Out).NotTo(Say(promptPassword))
			Expect(testUI.Err).NotTo(Say(promptPassword))
		})

		It("passes the data to the actor layer", func() {
			Expect(fakeActor.CreateServiceBrokerCallCount()).To(Equal(1))

			model := fakeActor.CreateServiceBrokerArgsForCall(0)

			Expect(model.Name).To(Equal(serviceBrokerName))
			Expect(model.Username).To(Equal(username))
			Expect(model.Password).To(Equal(promptPassword))
			Expect(model.URL).To(Equal(url))
			Expect(model.SpaceGUID).To(Equal(""))
		})
	})

	When("the --update-if-exists flag is used", func() {
		BeforeEach(func() {
			setPositionalFlags(cmd, serviceBrokerName, username, password, url)
			setFlag(cmd, "--update-if-exists")
		})

		Context("and the broker does not exist", func() {
			BeforeEach(func() {
				fakeActor.GetServiceBrokerByNameReturns(resources.ServiceBroker{}, v7action.Warnings{}, actionerror.ServiceBrokerNotFoundError{})
			})

			It("checks to see whether the broker exists", func() {
				Expect(fakeActor.GetServiceBrokerByNameCallCount()).To(Equal(1))
				Expect(fakeActor.GetServiceBrokerByNameArgsForCall(0)).To(Equal(serviceBrokerName))
			})

			It("creates a new service broker", func() {
				Expect(fakeActor.CreateServiceBrokerCallCount()).To(Equal(1))

				model := fakeActor.CreateServiceBrokerArgsForCall(0)

				Expect(model.Name).To(Equal(serviceBrokerName))
				Expect(model.Username).To(Equal(username))
				Expect(model.Password).To(Equal(password))
				Expect(model.URL).To(Equal(url))
				Expect(model.SpaceGUID).To(Equal(""))
			})
		})

		Context("and the broker already exists", func() {
			const brokerGUID = "fake-broker-guid"

			BeforeEach(func() {
				fakeActor.GetServiceBrokerByNameReturns(resources.ServiceBroker{GUID: brokerGUID}, v7action.Warnings{}, nil)
			})

			It("checks to see whether the broker exists", func() {
				Expect(fakeActor.GetServiceBrokerByNameCallCount()).To(Equal(1))
				Expect(fakeActor.GetServiceBrokerByNameArgsForCall(0)).To(Equal(serviceBrokerName))
			})

			It("updates an existing service broker", func() {
				Expect(fakeActor.UpdateServiceBrokerCallCount()).To(Equal(1))

				guid, model := fakeActor.UpdateServiceBrokerArgsForCall(0)

				Expect(guid).To(Equal(brokerGUID))
				Expect(model.Username).To(Equal(username))
				Expect(model.Password).To(Equal(password))
				Expect(model.URL).To(Equal(url))
				Expect(model.SpaceGUID).To(Equal(""))
			})
		})
	})
})
