package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("update-user-provided-service Command", func() {
	var (
		input           *Buffer
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeActor       *v7fakes.FakeActor
		fakeSharedActor *commandfakes.FakeSharedActor
		cmd             UpdateUserProvidedServiceCommand
		executeErr      error
	)

	expectOKMessage := func(testUI *ui.UI, serviceName, orgName, spaceName, userName string) {
		Expect(testUI.Out).To(SatisfyAll(
			Say("Updating user provided service %s in org %s / space %s as %s...", serviceName, orgName, spaceName, userName),
			Say("OK"),
			Say("TIP: Use 'cf restage' for any bound apps to ensure your env variable changes take effect"),
		))
	}

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(v7fakes.FakeActor)
		fakeSharedActor = new(commandfakes.FakeSharedActor)

		cmd = UpdateUserProvidedServiceCommand{
			BaseCommand: BaseCommand{
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

	It("checks the user is logged in, and targeting an org and space", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		orgChecked, spaceChecked := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(orgChecked).To(BeTrue())
		Expect(spaceChecked).To(BeTrue())
	})

	When("checking the target returns an error", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("explode"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("explode"))
		})
	})

	When("the user is logged in, and is targeting an org and space", func() {
		const (
			fakeServiceInstanceName = "fake-service-instance-name"
			fakeOrgName             = "fake-org-name"
			fakeSpaceName           = "fake-space-name"
			fakeSpaceGUID           = "fake-space-guid"
			fakeUserName            = "fake-user-name"
		)

		BeforeEach(func() {
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: fakeSpaceName,
				GUID: fakeSpaceGUID,
			})

			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: fakeOrgName,
			})

			fakeActor.GetCurrentUserReturns(configv3.User{Name: fakeUserName}, nil)

			setPositionalFlags(&cmd, fakeServiceInstanceName)

			fakeActor.UpdateUserProvidedServiceInstanceReturns(v7action.Warnings{"something obstreperous"}, nil)
		})

		When("no flags were specified", func() {
			It("succeeds with a message", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(SatisfyAll(
					Say("Updating user provided service %s in org %s / space %s as %s...", fakeServiceInstanceName, fakeOrgName, fakeSpaceName, fakeUserName),
					Say("No flags specified. No changes were made"),
					Say("OK"),
				))
			})
		})

		expectUpdate := func(update resources.ServiceInstance) {
			It("succeeds with a message", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				expectOKMessage(testUI, fakeServiceInstanceName, fakeOrgName, fakeSpaceName, fakeUserName)

				Expect(testUI.Err).To(Say("something obstreperous"))

				Expect(fakeActor.UpdateUserProvidedServiceInstanceCallCount()).To(Equal(1))
				actualName, actualSpaceGUID, actualUpdates := fakeActor.UpdateUserProvidedServiceInstanceArgsForCall(0)
				Expect(actualName).To(Equal(fakeServiceInstanceName))
				Expect(actualSpaceGUID).To(Equal(fakeSpaceGUID))
				Expect(actualUpdates).To(Equal(update))
			})
		}

		When("updating syslog URL", func() {
			BeforeEach(func() {
				setFlag(&cmd, "-l", flag.OptionalString{IsSet: true, Value: "https://syslog.com"})
			})

			expectUpdate(resources.ServiceInstance{
				SyslogDrainURL: types.NewOptionalString("https://syslog.com"),
			})
		})

		When("updating route service URL", func() {
			BeforeEach(func() {
				setFlag(&cmd, "-r", flag.OptionalString{IsSet: true, Value: "https://route.com"})
			})

			expectUpdate(resources.ServiceInstance{
				RouteServiceURL: types.NewOptionalString("https://route.com"),
			})
		})

		When("updating tags", func() {
			BeforeEach(func() {
				setFlag(&cmd, "-t", flag.Tags{IsSet: true, Value: []string{"one", "two", "three"}})
			})

			expectUpdate(resources.ServiceInstance{
				Tags: types.NewOptionalStringSlice("one", "two", "three"),
			})
		})

		When("updating credentials", func() {
			BeforeEach(func() {
				setFlag(&cmd, "-p", flag.CredentialsOrJSON{
					OptionalObject: types.OptionalObject{
						IsSet: true,
						Value: map[string]interface{}{"foo": "bar", "baz": false},
					},
				})
			})

			expectUpdate(resources.ServiceInstance{
				Credentials: types.NewOptionalObject(map[string]interface{}{"foo": "bar", "baz": false}),
			})
		})

		When("updating credentials interactively", func() {
			BeforeEach(func() {
				setFlag(&cmd, "-p", flag.CredentialsOrJSON{
					UserPromptCredentials: []string{"pass phrase", "cred"},
					OptionalObject: types.OptionalObject{
						IsSet: true,
						Value: nil,
					},
				})

				_, err := input.Write([]byte("very secret passphrase\nsecret cred\n"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("prompts the user for credentials", func() {
				Expect(testUI.Out).To(Say("pass phrase: "))
				Expect(testUI.Out).To(Say("cred: "))
			})

			It("does not echo the credentials", func() {
				Expect(testUI.Out).NotTo(Say("secret"))
				Expect(testUI.Err).NotTo(Say("secret"))
			})

			expectUpdate(resources.ServiceInstance{
				Credentials: types.NewOptionalObject(map[string]interface{}{
					"pass phrase": "very secret passphrase",
					"cred":        "secret cred",
				}),
			})
		})

		When("updating everything", func() {
			BeforeEach(func() {
				setFlag(&cmd, "-l", flag.OptionalString{IsSet: true, Value: "https://syslog.com"})
				setFlag(&cmd, "-r", flag.OptionalString{IsSet: true, Value: "https://route.com"})
				setFlag(&cmd, "-t", flag.Tags{IsSet: true, Value: []string{"one", "two", "three"}})
				setFlag(&cmd, "-p", flag.CredentialsOrJSON{
					OptionalObject: types.OptionalObject{
						IsSet: true,
						Value: map[string]interface{}{"foo": "bar", "baz": false},
					},
				})
			})

			expectUpdate(resources.ServiceInstance{
				SyslogDrainURL:  types.NewOptionalString("https://syslog.com"),
				RouteServiceURL: types.NewOptionalString("https://route.com"),
				Tags:            types.NewOptionalStringSlice("one", "two", "three"),
				Credentials:     types.NewOptionalObject(map[string]interface{}{"foo": "bar", "baz": false}),
			})
		})

		When("the service instance is not found", func() {
			BeforeEach(func() {
				setFlag(&cmd, "-l", flag.OptionalString{IsSet: true, Value: "https://syslog.com"})

				fakeActor.UpdateUserProvidedServiceInstanceReturns(
					v7action.Warnings{"something obstreperous"},
					ccerror.ServiceInstanceNotFoundError{},
				)
			})

			It("should return the correct error", func() {
				Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotFoundError{
					Name: fakeServiceInstanceName,
				}))
			})
		})

		When("the update fails", func() {
			BeforeEach(func() {
				setFlag(&cmd, "-l", flag.OptionalString{IsSet: true, Value: "https://syslog.com"})

				fakeActor.UpdateUserProvidedServiceInstanceReturns(
					v7action.Warnings{"something obstreperous"},
					errors.New("bang"),
				)
			})

			It("should fail with warnings and not say OK", func() {
				Expect(executeErr).To(MatchError("bang"))

				Expect(testUI.Out).To(Say("Updating user provided service %s in org %s / space %s as %s...", fakeServiceInstanceName, fakeOrgName, fakeSpaceName, fakeUserName))
				Expect(testUI.Out).NotTo(Say("OK"))

				Expect(testUI.Err).To(Say("something obstreperous"))
			})
		})
	})
})
