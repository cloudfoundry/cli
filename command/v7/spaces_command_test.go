package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("spaces Command", func() {
	var (
		cmd             SpacesCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = SpacesCommand{
			BaseCommand: command.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("an error is encountered checking if the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrgArg, checkTargetedSpaceArg := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrgArg).To(BeTrue())
			Expect(checkTargetedSpaceArg).To(BeFalse())
		})
	})

	When("the user is logged in and an org is targeted", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
		})

		When("getting the current user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("get-user-error"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("get-user-error"))
			})
		})

		When("getting the current user succeeds", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
			})

			When("there are no spaces", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationSpacesWithLabelSelectorReturns(
						[]v7action.Space{},
						v7action.Warnings{"get-spaces-warning"},
						nil,
					)
				})

				It("displays that there are no spaces", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say(`Getting spaces in org some-org as some-user\.\.\.`))
					Expect(testUI.Out).To(Say(""))
					Expect(testUI.Out).To(Say(`No spaces found\.`))

					Expect(testUI.Err).To(Say("get-spaces-warning"))

					Expect(fakeActor.GetOrganizationSpacesWithLabelSelectorCallCount()).To(Equal(1))

					orgGUID, labelSelector := fakeActor.GetOrganizationSpacesWithLabelSelectorArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(labelSelector).To(Equal(""))
				})
			})

			When("there are multiple spaces", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationSpacesWithLabelSelectorReturns(
						[]v7action.Space{
							{Name: "space-1"},
							{Name: "space-2"},
						},
						v7action.Warnings{"get-spaces-warning"},
						nil,
					)
				})

				It("displays all the spaces in the org", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say(`Getting spaces in org some-org as some-user\.\.\.`))
					Expect(testUI.Out).To(Say(""))
					Expect(testUI.Out).To(Say("name"))
					Expect(testUI.Out).To(Say("space-1"))
					Expect(testUI.Out).To(Say("space-2"))

					Expect(testUI.Err).To(Say("get-spaces-warning"))

					Expect(fakeActor.GetOrganizationSpacesWithLabelSelectorCallCount()).To(Equal(1))
					orgGUID, labelSelector := fakeActor.GetOrganizationSpacesWithLabelSelectorArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(labelSelector).To(Equal(""))
				})

				When("a label selector is provided to filter the spaces", func() {
					BeforeEach(func() {
						cmd.Labels = "some-label-selector"
					})
					It("passes the label selector to the actor", func() {
						Expect(fakeActor.GetOrganizationSpacesWithLabelSelectorCallCount()).To(Equal(1))
						orgGUID, labelSelector := fakeActor.GetOrganizationSpacesWithLabelSelectorArgsForCall(0)
						Expect(orgGUID).To(Equal("some-org-guid"))
						Expect(labelSelector).To(Equal("some-label-selector"))
					})
				})
			})

			When("a translatable error is encountered getting spaces", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationSpacesWithLabelSelectorReturns(
						nil,
						v7action.Warnings{"get-spaces-warning"},
						actionerror.OrganizationNotFoundError{Name: "not-found-org"},
					)
				})

				It("returns a translatable error", func() {
					Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "not-found-org"}))

					Expect(testUI.Out).To(Say(`Getting spaces in org some-org as some-user\.\.\.`))
					Expect(testUI.Out).To(Say(""))

					Expect(testUI.Err).To(Say("get-spaces-warning"))

					Expect(fakeActor.GetOrganizationSpacesWithLabelSelectorCallCount()).To(Equal(1))
					orgGUID, labelSelector := fakeActor.GetOrganizationSpacesWithLabelSelectorArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(labelSelector).To(Equal(""))
				})
			})
		})
	})
})
