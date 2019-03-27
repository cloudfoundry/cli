package v7_test

import (
	"errors"
	"regexp"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("set-label command", func() {
	var (
		cmd             SetLabelCommand
		fakeActor       *v7fakes.FakeSetLabelActor
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		testUI          *ui.UI

		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeActor = new(v7fakes.FakeSetLabelActor)
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		cmd = SetLabelCommand{
			Actor:       fakeActor,
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
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

	When("checking target succeeds", func() {
		var appName string

		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(nil)
		})

		When("fetching current user's name succeeds", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserNameReturns("some-user", nil)
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org"})
				fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "fake-space", GUID: "some-space-guid"})

				u, err := uuid.NewV4()
				Expect(err).NotTo(HaveOccurred())
				appName = u.String()
			})

			When("all the provided labels are valid", func() {
				BeforeEach(func() {
					cmd.RequiredArgs = flag.SetLabelArgs{
						ResourceType: "app",
						ResourceName: appName,
						Labels:       []string{"FOO=BAR", "ENV=FAKE"},
					}
					fakeActor.UpdateApplicationLabelsByApplicationNameReturns(
						v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
						nil,
					)
				})

				When("updating the application labels succeeds", func() {
					It("sets the provided labels on the app", func() {
						name, spaceGUID, labels := fakeActor.UpdateApplicationLabelsByApplicationNameArgsForCall(0)
						Expect(name).To(Equal(appName), "failed to pass app name")
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(labels).To(BeEquivalentTo(map[string]string{
							"FOO": "BAR",
							"ENV": "FAKE",
						}))
					})

					It("displays a message", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))

						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for app %s in org fake-org / space fake-space as some-user...`), appName))
						Expect(testUI.Out).To(Say("OK"))
					})

					It("prints all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})
				})
			})

			When("updating the application labels fail", func() {
				BeforeEach(func() {
					cmd.RequiredArgs = flag.SetLabelArgs{
						ResourceType: "app",
						ResourceName: appName,
						Labels:       []string{"FOO=BAR", "ENV=FAKE"},
					}
					fakeActor.UpdateApplicationLabelsByApplicationNameReturns(
						v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
						errors.New("some-updating-error"),
					)
				})
				It("displays warnings and an error message", func() {
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
					Expect(executeErr).To(HaveOccurred())
					Expect(executeErr).To(MatchError("some-updating-error"))
				})
			})

			When("some provided labels do not have a value part", func() {
				BeforeEach(func() {
					cmd.RequiredArgs = flag.SetLabelArgs{
						ResourceType: "app",
						ResourceName: appName,
						Labels:       []string{"FOO=BAR", "MISSING_EQUALS", "ENV=FAKE"},
					}
				})

				It("complains about the missing equal sign", func() {
					Expect(executeErr).To(MatchError("Metadata error: no value provided for label 'MISSING_EQUALS'"))
					Expect(executeErr).To(HaveOccurred())
				})
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

	When("checking targeted org and space fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("nope"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("nope"))
		})
	})
})
