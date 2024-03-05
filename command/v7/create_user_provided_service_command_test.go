package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/types"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-user-provided-service Command", func() {
	var (
		input           *Buffer
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeActor       *v7fakes.FakeActor
		fakeSharedActor *commandfakes.FakeSharedActor
		cmd             CreateUserProvidedServiceCommand
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(v7fakes.FakeActor)
		fakeSharedActor = new(commandfakes.FakeSharedActor)

		cmd = CreateUserProvidedServiceCommand{
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

			fakeActor.CreateUserProvidedServiceInstanceReturns(v7action.Warnings{"be warned", "take care"}, nil)
		})

		It("succeeds", func() {
			Expect(executeErr).NotTo(HaveOccurred())
		})

		It("prints a message and warnings", func() {
			Expect(testUI.Out).To(SatisfyAll(
				Say("Creating user provided service %s in org %s / space %s as %s...", fakeServiceInstanceName, fakeOrgName, fakeSpaceName, fakeUserName),
				Say("OK"),
			))

			Expect(testUI.Err).To(SatisfyAll(
				Say("be warned"),
				Say("take care"),
			))
		})

		It("calls the actor with the service instance name and space GUID", func() {
			Expect(fakeActor.CreateUserProvidedServiceInstanceCallCount()).To(Equal(1))
			serviceInstance := fakeActor.CreateUserProvidedServiceInstanceArgsForCall(0)
			Expect(serviceInstance).To(Equal(resources.ServiceInstance{
				Name:      fakeServiceInstanceName,
				SpaceGUID: fakeSpaceGUID,
			}))
		})

		When("all optional arguments are provided", func() {
			BeforeEach(func() {
				setFlag(&cmd, "-t", flag.Tags{IsSet: true, Value: []string{"list", "of", "tags"}})
				setFlag(&cmd, "-l", flag.OptionalString{IsSet: true, Value: "https://fake-sylogg.com"})
				setFlag(&cmd, "-r", flag.OptionalString{IsSet: true, Value: "https://fake-route.com"})
				setFlag(&cmd, "-p", flag.CredentialsOrJSON{
					OptionalObject: types.OptionalObject{
						IsSet: true,
						Value: map[string]interface{}{
							"foo": "bar",
							"baz": 42,
						},
					},
				})
			})

			It("calls the actor with all the flag values", func() {
				Expect(fakeActor.CreateUserProvidedServiceInstanceCallCount()).To(Equal(1))
				serviceInstance := fakeActor.CreateUserProvidedServiceInstanceArgsForCall(0)
				Expect(serviceInstance).To(Equal(resources.ServiceInstance{
					Name:            fakeServiceInstanceName,
					SpaceGUID:       fakeSpaceGUID,
					Tags:            types.NewOptionalStringSlice("list", "of", "tags"),
					SyslogDrainURL:  types.NewOptionalString("https://fake-sylogg.com"),
					RouteServiceURL: types.NewOptionalString("https://fake-route.com"),
					Credentials: types.NewOptionalObject(map[string]interface{}{
						"foo": "bar",
						"baz": 42,
					}),
				}))
			})
		})

		When("setting credentials interactively", func() {
			BeforeEach(func() {
				cmd.Credentials.IsSet = true
				cmd.Credentials.UserPromptCredentials = []string{"pass phrase", "cred"}

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

			It("succeeds", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(fakeActor.CreateUserProvidedServiceInstanceCallCount()).To(Equal(1))
				serviceInstance := fakeActor.CreateUserProvidedServiceInstanceArgsForCall(0)
				Expect(serviceInstance).To(Equal(resources.ServiceInstance{
					Name:      fakeServiceInstanceName,
					SpaceGUID: fakeSpaceGUID,
					Credentials: types.NewOptionalObject(map[string]interface{}{
						"pass phrase": "very secret passphrase",
						"cred":        "secret cred",
					}),
				}))
			})
		})

		When("getting the user fails", func() {
			BeforeEach(func() {
				fakeActor.GetCurrentUserReturns(configv3.User{Name: fakeUserName}, errors.New("boom"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("boom"))
			})
		})

		When("the actor returns an error", func() {
			BeforeEach(func() {
				fakeActor.CreateUserProvidedServiceInstanceReturns(v7action.Warnings{"be warned", "take care"}, errors.New("bang"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("bang"))
			})

			It("prints a message and warnings", func() {
				Expect(testUI.Out).To(Say("Creating user provided service %s in org %s / space %s as %s...", fakeServiceInstanceName, fakeOrgName, fakeSpaceName, fakeUserName))
				Expect(testUI.Out).NotTo(Say("OK"))

				Expect(testUI.Err).To(SatisfyAll(
					Say("be warned"),
					Say("take care"),
				))
			})
		})
	})
})
