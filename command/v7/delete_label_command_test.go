package v7_test

import (
	"errors"
	"regexp"

	"code.cloudfoundry.org/cli/command/flag"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("delete-label command", func() {
	var (
		cmd             DeleteLabelCommand
		fakeConfig      *commandfakes.FakeConfig
		testUI          *ui.UI
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeDeleteLabelActor
		executeErr      error
	)

	When("deleting labels on apps", func() {
		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeConfig = new(commandfakes.FakeConfig)
			fakeSharedActor = new(commandfakes.FakeSharedActor)
			fakeActor = new(v7fakes.FakeDeleteLabelActor)
			cmd = DeleteLabelCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			}
			cmd.RequiredArgs = flag.DeleteLabelArgs{
				ResourceType: "app",
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

		When("checking the target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(errors.New("Target not found"))
			})

			It("we expect an error to be returned", func() {
				Expect(executeErr).To(MatchError("Target not found"))
			})
		})

		When("checking the target succeeds", func() {
			var appName string

			BeforeEach(func() {
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org"})
				fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "fake-space", GUID: "some-space-guid"})
				appName = "some-app"
				cmd.RequiredArgs.ResourceName = appName
			})

			When("getting the current user succeeds", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
					cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}
				})

				It("informs the user that labels are being deleted", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Deleting label(s) for app %s in org fake-org / space fake-space as some-user...`), appName))
				})

				When("updating the app labels succeeds", func() {
					BeforeEach(func() {
						fakeActor.UpdateApplicationLabelsByApplicationNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
							nil)
					})

					It("does not return an error", func() {
						Expect(executeErr).ToNot(HaveOccurred())
					})

					It("prints all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})

					It("passes the correct parameters into the actor", func() {
						expectedMap := map[string]types.NullString{
							"some-label":     types.NewNullString(),
							"some-other-key": types.NewNullString()}

						Expect(fakeActor.UpdateApplicationLabelsByApplicationNameCallCount()).To(Equal(1))
						actualAppName, spaceGUID, labelsMap := fakeActor.UpdateApplicationLabelsByApplicationNameArgsForCall(0)
						Expect(actualAppName).To(Equal(appName))
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(labelsMap).To(Equal(expectedMap))
					})
				})

				When("updating the app labels fails", func() {
					BeforeEach(func() {
						fakeActor.UpdateApplicationLabelsByApplicationNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
							errors.New("api call failed"))
					})

					It("prints all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError("api call failed"))
					})
				})
			})
			When("getting the user fails", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("could not get user"))
					cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("could not get user"))
				})
			})
		})
	})

	When("deleting labels on orgs", func() {
		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeConfig = new(commandfakes.FakeConfig)
			fakeSharedActor = new(commandfakes.FakeSharedActor)
			fakeActor = new(v7fakes.FakeDeleteLabelActor)
			cmd = DeleteLabelCommand{
				Actor:       fakeActor,
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
			}
			cmd.RequiredArgs = flag.DeleteLabelArgs{
				ResourceType: "org",
			}
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		When("checking target succeeds", func() {
			var orgName = "some-org"

			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(nil)
				cmd.RequiredArgs.ResourceName = orgName

			})

			When("fetching current user's name succeeds", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
					cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}
				})

				It("informs the user that labels are being deleted", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Deleting label(s) for org %s as some-user...`), orgName))
				})

				When("updating the org labels succeeds", func() {
					BeforeEach(func() {
						fakeActor.UpdateOrganizationLabelsByOrganizationNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
							nil)
					})

					It("does not return an error", func() {
						Expect(executeErr).ToNot(HaveOccurred())
					})

					It("prints all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})

					It("passes the correct parameters into the actor", func() {
						expectedMaps := map[string]types.NullString{
							"some-label":     types.NewNullString(),
							"some-other-key": types.NewNullString()}

						Expect(fakeActor.UpdateOrganizationLabelsByOrganizationNameCallCount()).To(Equal(1))
						actualOrgName, labelsMap := fakeActor.UpdateOrganizationLabelsByOrganizationNameArgsForCall(0)
						Expect(actualOrgName).To(Equal(orgName))
						Expect(labelsMap).To(Equal(expectedMaps))
					})
				})

			})

			When("fetching the current user's name fails", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("could not get user"))
					cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("could not get user"))
				})
			})
		})
	})

	When("deleting labels on spaces", func() {
		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeConfig = new(commandfakes.FakeConfig)
			fakeSharedActor = new(commandfakes.FakeSharedActor)
			fakeActor = new(v7fakes.FakeDeleteLabelActor)
			cmd = DeleteLabelCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			}
			cmd.RequiredArgs = flag.DeleteLabelArgs{
				ResourceType: "space",
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
			Expect(checkSpace).To(BeFalse())
		})

		When("checking the target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(errors.New("Target not found"))
			})

			It("we expect an error to be returned", func() {
				Expect(executeErr).To(MatchError("Target not found"))
			})
		})

		When("checking the target succeeds", func() {
			var spaceName string

			BeforeEach(func() {
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org", GUID: "some-org-guid"})
				spaceName = "spiff"
				cmd.RequiredArgs.ResourceName = spaceName
			})

			When("getting the current user succeeds", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
					cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}
				})

				It("informs the user that labels are being deleted", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Deleting label(s) for space %s in org fake-org as some-user...`), spaceName))
				})

				When("updating the space labels succeeds", func() {
					BeforeEach(func() {
						fakeActor.UpdateSpaceLabelsBySpaceNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
							nil)
					})

					It("does not return an error", func() {
						Expect(executeErr).ToNot(HaveOccurred())
					})

					It("prints all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})

					It("passes the correct parameters into the actor", func() {
						expectedMap := map[string]types.NullString{
							"some-label":     types.NewNullString(),
							"some-other-key": types.NewNullString()}

						Expect(fakeActor.UpdateSpaceLabelsBySpaceNameCallCount()).To(Equal(1))
						actualSpaceName, orgGUID, labelsMap := fakeActor.UpdateSpaceLabelsBySpaceNameArgsForCall(0)
						Expect(actualSpaceName).To(Equal(spaceName))
						Expect(orgGUID).To(Equal("some-org-guid"))
						Expect(labelsMap).To(Equal(expectedMap))
					})
				})

				When("updating the space labels fails", func() {
					BeforeEach(func() {
						fakeActor.UpdateSpaceLabelsBySpaceNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
							errors.New("api call failed"))
					})

					It("prints all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError("api call failed"))
					})
				})
			})
			When("getting the user fails", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("could not get user"))
					cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("could not get user"))
				})
			})
		})
	})
})
