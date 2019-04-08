package v7_test

import (
	"errors"
	"regexp"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("labels command", func() {
	var (
		cmd             LabelsCommand
		fakeActor       *v7fakes.FakeAppActor
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		testUI          *ui.UI

		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeActor = new(v7fakes.FakeAppActor)
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		cmd = LabelsCommand{
			Actor:       fakeActor,
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Describe("listing labels", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserNameReturns("some-user", nil)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org"})
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "fake-space", GUID: "some-space-guid"})
			cmd.RequiredArgs = flag.LabelsArgs{
				ResourceType: "app",
				ResourceName: "dora",
			}
			fakeActor.GetApplicationByNameAndSpaceReturns(
				v7action.Application{
					Metadata: struct {
						Labels map[string]types.NullString `json:"labels,omitempty"`
					}{
						Labels: map[string]types.NullString{
							"some-other-label": types.NewNullString("some-other-value"),
							"some-label":       types.NewNullString("some-value"),
						},
					},
				},
				v7action.Warnings{},
				nil)
		})

		It("doesn't error", func() {
			Expect(executeErr).ToNot(HaveOccurred())
		})

		It("checks that the user is logged in and targeted to an org and space", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkOrg).To(BeTrue())
			Expect(checkSpace).To(BeTrue())
		})

		It("displays a message that it is retrieving the labels", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Getting labels for app dora in org fake-org / space fake-space as some-user...`)))
		})

		It("retrieves the labels associated with the application", func() {
			Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
			appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
			Expect(appName).To(Equal("dora"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
		})

		It("displays the labels that are associated with the application, alphabetically", func() {
			Expect(testUI.Out).To(Say(`Key\s+Value`))
			Expect(testUI.Out).To(Say(`some-label\s+some-value`))
			Expect(testUI.Out).To(Say(`some-other-label\s+some-other-value`))
		})

		When("CAPI returns warnings", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v7action.Application{
						Metadata: struct {
							Labels map[string]types.NullString `json:"labels,omitempty"`
						}{
							Labels: map[string]types.NullString{
								"some-other-label": types.NewNullString("some-other-value"),
								"some-label":       types.NewNullString("some-value"),
							},
						},
					},
					v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
					nil)
			})

			It("prints all warnings", func() {
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})
		})

		When("there is an error retrieving the application", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v7action.Application{},
					v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
					errors.New("boom"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("boom"))
			})

			It("still prints all warnings", func() {
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

			It("doesn't say ok", func() {
				Expect(testUI.Out).ToNot(Say("OK"))
			})
		})

		When("checking targeted org and space fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(errors.New("nope"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("nope"))
			})
		})

		When("fetching the current user's name fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserNameReturns("some-user", errors.New("boom"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("boom"))
			})
		})
	})
})
