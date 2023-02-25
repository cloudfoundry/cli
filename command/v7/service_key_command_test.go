package v7_test

import (
	"errors"

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

var _ = Describe("service-key Command", func() {
	var (
		cmd             v7.ServiceKeyCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		executeErr      error
		fakeActor       *v7fakes.FakeActor
	)

	const (
		fakeServiceInstanceName = "fake-service-instance-name"
		fakeServiceKeyName      = "fake-service-key-name"
		fakeSpaceGUID           = "fake-space-guid"
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = v7.ServiceKeyCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		fakeConfig.TargetedSpaceReturns(configv3.Space{GUID: fakeSpaceGUID})

		setPositionalFlags(&cmd, fakeServiceInstanceName, fakeServiceKeyName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("checks the user is logged in, and targeting an org and space", func() {
		Expect(executeErr).NotTo(HaveOccurred())

		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		actualOrg, actualSpace := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(actualOrg).To(BeTrue())
		Expect(actualSpace).To(BeTrue())
	})

	When("getting details", func() {
		const fakeUserName = "fake-user-name"

		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{Name: fakeUserName}, nil)

			fakeActor.GetServiceKeyDetailsByServiceInstanceAndNameReturns(
				resources.ServiceCredentialBindingDetails{
					Credentials: map[string]interface{}{"foo": "bar", "pass": "<3test"},
				},
				v7action.Warnings{"a warning"},
				nil,
			)
		})

		It("delegates to the actor", func() {
			Expect(fakeActor.GetServiceKeyDetailsByServiceInstanceAndNameCallCount()).To(Equal(1))
			actualServiceInstanceName, actualKeyName, actualSpaceGUID := fakeActor.GetServiceKeyDetailsByServiceInstanceAndNameArgsForCall(0)
			Expect(actualServiceInstanceName).To(Equal(fakeServiceInstanceName))
			Expect(actualKeyName).To(Equal(fakeServiceKeyName))
			Expect(actualSpaceGUID).To(Equal(fakeSpaceGUID))
		})

		It("prints an intro, details, and warnings", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Err).To(Say("a warning"))
			Expect(testUI.Out).To(SatisfyAll(
				Say(`Getting key %s for service instance %s as %s\.\.\.\n`, fakeServiceKeyName, fakeServiceInstanceName, fakeUserName),
				Say(`\n`),
				Say(`\{\n`),
				Say(`  "foo": "bar",\n`),
				Say(`  "pass": "<3test"\n`),
				Say(`\}\n`),
			))
		})

		When("getting the username returns an error", func() {
			BeforeEach(func() {
				fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("bad thing"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("bad thing"))
			})
		})

		When("actor returns another error", func() {
			BeforeEach(func() {
				fakeActor.GetServiceKeyDetailsByServiceInstanceAndNameReturns(
					resources.ServiceCredentialBindingDetails{},
					v7action.Warnings{"a warning"},
					errors.New("bang"),
				)
			})

			It("prints warnings and returns an error", func() {
				Expect(testUI.Err).To(Say("a warning"))
				Expect(executeErr).To(MatchError("bang"))
			})
		})
	})

	When("getting GUID", func() {
		const fakeGUID = "fake-key-guid"

		BeforeEach(func() {
			fakeActor.GetServiceKeyByServiceInstanceAndNameReturns(
				resources.ServiceCredentialBinding{GUID: fakeGUID},
				v7action.Warnings{"a warning"},
				nil,
			)

			setFlag(&cmd, "--guid")
		})

		It("delegates to the actor", func() {
			Expect(fakeActor.GetServiceKeyByServiceInstanceAndNameCallCount()).To(Equal(1))
			actualServiceInstanceName, actualKeyName, actualSpaceGUID := fakeActor.GetServiceKeyByServiceInstanceAndNameArgsForCall(0)
			Expect(actualServiceInstanceName).To(Equal(fakeServiceInstanceName))
			Expect(actualKeyName).To(Equal(fakeServiceKeyName))
			Expect(actualSpaceGUID).To(Equal(fakeSpaceGUID))
		})

		It("prints the GUID and nothing else", func() {
			Expect(testUI.Out).To(Say(fakeGUID))
			Expect(testUI.Err).NotTo(Say("a warning"))
		})

		When("actor returns an error", func() {
			BeforeEach(func() {
				fakeActor.GetServiceKeyByServiceInstanceAndNameReturns(
					resources.ServiceCredentialBinding{},
					v7action.Warnings{"a warning"},
					errors.New("bang"),
				)
			})

			It("prints warnings and returns an error", func() {
				Expect(testUI.Err).To(Say("a warning"))
				Expect(executeErr).To(MatchError("bang"))
			})
		})
	})
})
