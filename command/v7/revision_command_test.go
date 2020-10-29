package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("revision Command", func() {
	var (
		cmd             v7.RevisionCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		appName         string

		out *Buffer
	)

	BeforeEach(func() {
		out = NewBuffer()
		testUI = ui.NewTestUI(nil, out, NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = v7.RevisionCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}
		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		appName = "some-app"

		cmd.RequiredArgs.AppName = appName
	})

	JustBeforeEach(func() {
		cmd.Execute(nil)
	})

	It("displays the experimental warning", func() {
		Expect(testUI.Err).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
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

			When("when the requested app and revision exist", func() {
				var revisions []resources.Revision
				BeforeEach(func() {
					fakeApp := resources.Application{
						GUID: "fake-guid",
						Name: "some-app",
					}
					fakeActor.GetApplicationByNameAndSpaceReturns(fakeApp, nil, nil)

					revisions = []resources.Revision{
						{
							Version:     3,
							GUID:        "A68F13F7-7E5E-4411-88E8-1FAC54F73F50",
							Description: "On a different note",
							CreatedAt:   "2020-03-10T17:11:58Z",
							Deployable:  true,
							Droplet: resources.Droplet{
								GUID: "droplet-guid",
							},
						},
						{
							Version:     2,
							GUID:        "A89F8259-D32B-491A-ABD6-F100AC42D74C",
							Description: "Something else",
							CreatedAt:   "2020-03-08T12:43:30Z",
							Deployable:  true,
						},
						{
							Version:     1,
							GUID:        "17E0E587-0E53-4A6E-B6AE-82073159F910",
							Description: "Something",
							CreatedAt:   "2020-03-04T13:23:32Z",
							Deployable:  false,
						},
					}
					fakeActor.GetRevisionsByApplicationNameAndSpaceReturns(revisions, nil, nil)
					fakeActor.GetApplicationByNameAndSpaceReturns(resources.Application{GUID: "app-guid"}, nil, nil)
					fakeActor.GetApplicationRevisionsDeployedReturns(revisions[0:1], nil, nil)

					cmd.Version = flag.Revision{NullInt: types.NullInt{Value: 3, IsSet: true}}
				})

				It("gets the app guid", func() {
					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})

				It("retrieves the deployed revisions", func() {
					Expect(fakeActor.GetApplicationRevisionsDeployedCallCount()).To(Equal(1))
					Expect(fakeActor.GetApplicationRevisionsDeployedArgsForCall(0)).To(Equal("app-guid"))
				})

				It("displays the revision", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say(`Showing revision 3 for app some-app in org some-org / space some-space as banana...`))
					Expect(testUI.Out).To(Say(`revision:        3`))
					Expect(testUI.Out).To(Say(`deployed:        true`))
					Expect(testUI.Out).To(Say(`description:     On a different note`))
					Expect(testUI.Out).To(Say(`deployable:      true`))
					Expect(testUI.Out).To(Say(`revision GUID:   A68F13F7-7E5E-4411-88E8-1FAC54F73F50`))
					Expect(testUI.Out).To(Say(`droplet GUID:    droplet-guid`))
					Expect(testUI.Out).To(Say(`created on:      2020-03-10T17:11:58Z`))

					// Expect(testUI.Out).To(Say(`labels:`))
					// Expect(testUI.Out).To(Say(`label: foo3`))

					// Expect(testUI.Out).To(Say(`annotations:`))
					// Expect(testUI.Out).To(Say(`annotation: foo3`))

					Expect(testUI.Out).To(Say(`application environment variables:`))
					Expect(testUI.Out).To(Say(`env: foo3`))

					Expect(fakeActor.GetRevisionsByApplicationNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetRevisionsByApplicationNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})

				When("revision is not deployed", func() {
					BeforeEach(func() {
						fakeActor.GetApplicationRevisionsDeployedReturns(revisions[1:2], nil, nil)
					})

					It("displays deployed field correctly", func() {
						Expect(testUI.Out).To(Say(`deployed:        false`))
					})
				})

				When("no revisions were deployed", func() {
					BeforeEach(func() {
						fakeActor.GetApplicationRevisionsDeployedReturns([]resources.Revision{}, nil, nil)
					})

					It("displays deployed field correctly", func() {
						Expect(testUI.Out).To(Say(`deployed:        false`))
					})
				})
			})
		})
	})
})
