package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/util/configv3"

	"code.cloudfoundry.org/cli/command/translatableerror"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("rename-service command test", func() {
	const (
		currentName = "current-name"
		newName     = "new-name"
		spaceName   = "fake-space-name"
		spaceGUID   = "fake-space-guid"
		orgName     = "fake-org-name"
		username    = "fake-username"
	)

	var (
		cmd             RenameServiceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
	)

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = RenameServiceCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		setPositionalFlags(&cmd, currentName, newName)

		fakeConfig.TargetedSpaceReturns(configv3.Space{GUID: spaceGUID, Name: spaceName})
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: orgName})
		fakeActor.GetCurrentUserReturns(configv3.User{Name: username}, nil)

		fakeActor.RenameServiceInstanceReturns(v7action.Warnings{"rename instance warning"}, nil)
	})

	It("checks the user is logged in, and targeting an org and space", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		orgChecked, spaceChecked := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(orgChecked).To(BeTrue())
		Expect(spaceChecked).To(BeTrue())
	})

	It("prints messages", func() {
		Expect(testUI.Out).To(SatisfyAll(
			Say(`Renaming service %s to %s in org %s / space %s as %s...\n`, currentName, newName, orgName, spaceName, username),
			Say(`OK`),
		))
	})

	It("delegates the rename to the actor", func() {
		Expect(fakeActor.RenameServiceInstanceCallCount()).To(Equal(1))
		actualCurrentName, actualSpaceGUID, actualNewName := fakeActor.RenameServiceInstanceArgsForCall(0)
		Expect(actualCurrentName).To(Equal(currentName))
		Expect(actualSpaceGUID).To(Equal(spaceGUID))
		Expect(actualNewName).To(Equal(newName))
	})

	It("prints warnings", func() {
		Expect(testUI.Err).To(Say("rename instance warning"))
	})

	It("does not return an error", func() {
		Expect(executeErr).NotTo(HaveOccurred())
	})

	When("checking the target returns an error", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("explode"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("explode"))
		})
	})

	When("getting the user returns an error", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("bang"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("bang"))
		})
	})

	When("the service instance cannot be found", func() {
		BeforeEach(func() {
			fakeActor.RenameServiceInstanceReturns(
				v7action.Warnings{"rename instance warning"},
				actionerror.ServiceInstanceNotFoundError{Name: currentName},
			)
		})

		It("prints a tip, warnings and returns an error", func() {
			Expect(testUI.Err).To(Say("rename instance warning"))
			Expect(testUI.Out).To(Say(`TIP: Use 'cf services' to view all services in this org and space\.`))
			Expect(executeErr).To(MatchError(translatableerror.ServiceInstanceNotFoundError{Name: currentName}))
		})
	})

	When("renaming the service instance fails", func() {
		BeforeEach(func() {
			fakeActor.RenameServiceInstanceReturns(
				v7action.Warnings{"rename instance warning"},
				errors.New("unknown bad thing"),
			)
		})

		It("prints warnings and returns an error", func() {
			Expect(testUI.Err).To(Say("rename instance warning"))
			Expect(testUI.Out).NotTo(Say("TIP"))
			Expect(executeErr).To(MatchError("unknown bad thing"))
		})
	})
})
