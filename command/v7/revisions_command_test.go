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

var _ = Describe("revisions Command", func() {
	var (
		cmd             RevisionsCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeRevisionsActor
		binaryName      string
		executeErr      error
		appName         string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeRevisionsActor)

		cmd = RevisionsCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		appName = "some-app"

		cmd.RequiredArgs.AppName = appName
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is logged in, an org is targeted and a space is targeted", func() {
		BeforeEach(func() {
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
		})

		When("getting the current user returns an error", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("some-error"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("some-error"))
			})
		})

		When("getting the current user succeeds", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
			})

			When("getting the revisions", func() {
				BeforeEach(func() {
					revisions := v7action.Revisions{
						{
							Version:     1,
							GUID:        "17E0E587-0E53-4A6E-B6AE-82073159F910",
							Description: "Something",
							CreatedAt:   "2020-03-04T13:23:32Z",
						},
						{
							Version:     2,
							GUID:        "A89F8259-D32B-491A-ABD6-F100AC42D74C",
							Description: "Something else",
							CreatedAt:   "2020-03-08T12:43:30Z",
						},
						{
							Version:     3,
							GUID:        "A68F13F7-7E5E-4411-88E8-1FAC54F73F50",
							Description: "On a different note",
							CreatedAt:   "2020-03-10T17:11:58Z",
						},
					}
					fakeActor.GetRevisionsByApplicationNameAndSpaceReturns(revisions, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
				})

				It("displays the revisions", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say(`Getting revisions for app some-app in org some-org / space some-space as banana\.\.\.`))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say("version   guid                                   description           created at"))
					Expect(testUI.Out).To(Say("1         17E0E587-0E53-4A6E-B6AE-82073159F910   Something             2020-03-04T13:23:32Z"))
					Expect(testUI.Out).To(Say("2         A89F8259-D32B-491A-ABD6-F100AC42D74C   Something else        2020-03-08T12:43:30Z"))
					Expect(testUI.Out).To(Say("3         A68F13F7-7E5E-4411-88E8-1FAC54F73F50   On a different note   2020-03-10T17:11:58Z"))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))

					Expect(fakeActor.GetRevisionsByApplicationNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetRevisionsByApplicationNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})

			})

			When("revisions variables returns an unknown error", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeActor.GetRevisionsByApplicationNameAndSpaceReturns(v7action.Revisions{}, v7action.Warnings{"get-warning-1", "get-warning-2"}, expectedErr)
				})

				It("returns the error", func() {
					Expect(executeErr).To(Equal(expectedErr))
					Expect(testUI.Out).To(Say(`Getting revisions for app some-app in org some-org / space some-space as banana\.\.\.`))
					Expect(testUI.Out).To(Say("OK"))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))
				})
			})
		})
	})
})
