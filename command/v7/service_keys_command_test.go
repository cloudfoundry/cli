package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("service-keys Command", func() {
	var (
		cmd             v7.ServiceKeysCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		executeErr      error
		fakeActor       *v7fakes.FakeActor
	)

	const (
		fakeUserName            = "fake-user-name"
		fakeServiceInstanceName = "fake-service-instance-name"
		fakeSpaceGUID           = "fake-space-guid"
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = v7.ServiceKeysCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		fakeConfig.TargetedSpaceReturns(configv3.Space{GUID: fakeSpaceGUID})

		fakeConfig.CurrentUserReturns(configv3.User{Name: fakeUserName}, nil)

		fakeActor.GetServiceKeysByServiceInstanceReturns(
			[]string{"flopsy", "mopsy", "cottontail", "peter"},
			v7action.Warnings{"fake warning"},
			nil,
		)

		setPositionalFlags(&cmd, fakeServiceInstanceName)
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

	It("delegates to the actor", func() {
		Expect(fakeActor.GetServiceKeysByServiceInstanceCallCount()).To(Equal(1))
		actualServiceInstanceName, actualSpaceGUID := fakeActor.GetServiceKeysByServiceInstanceArgsForCall(0)
		Expect(actualServiceInstanceName).To(Equal(fakeServiceInstanceName))
		Expect(actualSpaceGUID).To(Equal(fakeSpaceGUID))
	})

	It("prints an intro, key names, and warnings", func() {
		Expect(executeErr).NotTo(HaveOccurred())
		Expect(testUI.Err).To(Say("fake warning"))
		Expect(testUI.Out).To(SatisfyAll(
			Say(`Getting keys for service instance %s as %s\.\.\.\n`, fakeServiceInstanceName, fakeUserName),
			Say(`\n`),
			Say(`name\n`),
			Say(`flopsy\n`),
			Say(`mopsy\n`),
			Say(`cottontail\n`),
			Say(`peter\n`),
		))
	})

	When("there are no keys", func() {
		BeforeEach(func() {
			fakeActor.GetServiceKeysByServiceInstanceReturns(
				nil,
				v7action.Warnings{"fake warning"},
				nil,
			)
		})

		It("prints an intro, message, and warnings", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Err).To(Say("fake warning"))
			Expect(testUI.Out).To(SatisfyAll(
				Say(`Getting keys for service instance %s as %s\.\.\.\n`, fakeServiceInstanceName, fakeUserName),
				Say(`\n`),
				Say(`No service keys for service instance %s\n`, fakeServiceInstanceName),
			))
		})
	})

	When("the service instance is user-provided", func() {
		BeforeEach(func() {
			fakeActor.GetServiceKeysByServiceInstanceReturns(
				nil,
				v7action.Warnings{"fake warning"},
				actionerror.ServiceInstanceTypeError{},
			)
		})

		It("returns a helpful error and prints warnings", func() {
			Expect(testUI.Err).To(Say("fake warning"))
			Expect(executeErr).To(MatchError(translatableerror.ServiceKeysNotSupportedWithUserProvidedServiceInstances{}))
		})
	})

	When("checking the target returns an error", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("explode"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("explode"))
		})
	})

	When("getting the username returns an error", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("bad thing"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("bad thing"))
		})
	})

	When("getting the keys returns an error", func() {
		BeforeEach(func() {
			fakeActor.GetServiceKeysByServiceInstanceReturns(
				nil,
				v7action.Warnings{"fake warning"},
				errors.New("boom"),
			)
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("boom"))
		})
	})
})
