package v7_test

import (
	"errors"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("revisions Command", func() {
	var (
		cmd             RevisionsCommand
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

		cmd = v7.RevisionsCommand{
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
		executeErr = cmd.Execute(nil)
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
				fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("some-error"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("some-error"))
			})
		})

		When("getting the current user succeeds", func() {
			BeforeEach(func() {
				fakeActor.GetCurrentUserReturns(configv3.User{Name: "banana"}, nil)
			})

			When("when revisions are available", func() {
				BeforeEach(func() {
					revisions := []resources.Revision{
						{
							Version:     3,
							GUID:        "A68F13F7-7E5E-4411-88E8-1FAC54F73F50",
							Description: "On a different note",
							CreatedAt:   "2020-03-10T17:11:58Z",
							Deployable:  true,
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
					fakeActor.GetRevisionsByApplicationNameAndSpaceReturns(revisions, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)

					fakeApp := resources.Application{
						GUID: "fake-guid",
					}
					fakeActor.GetApplicationByNameAndSpaceReturns(fakeApp, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)

					deployedRevisions := []resources.Revision{
						{
							Version:     3,
							GUID:        "A68F13F7-7E5E-4411-88E8-1FAC54F73F50",
							Description: "On a different note",
							CreatedAt:   "2020-03-10T17:11:58Z",
							Deployable:  true,
						},
					}
					fakeActor.GetApplicationRevisionsDeployedReturns(deployedRevisions, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
				})

				It("displays the revisions", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeActor.GetApplicationRevisionsDeployedCallCount()).To(Equal(1))
					appGUID := fakeActor.GetApplicationRevisionsDeployedArgsForCall(0)
					Expect(appGUID).To(Equal("fake-guid"))

					Expect(testUI.Out).To(Say(`Getting revisions for app some-app in org some-org / space some-space as banana\.\.\.`))
					Expect(testUI.Out).To(Say("revision      description           deployable   revision guid                          created at"))
					Expect(testUI.Out).To(Say("3\\(deployed\\)   On a different note   true         A68F13F7-7E5E-4411-88E8-1FAC54F73F50   2020-03-10T17:11:58Z"))
					Expect(testUI.Out).To(Say("2             Something else        true         A89F8259-D32B-491A-ABD6-F100AC42D74C   2020-03-08T12:43:30Z"))
					Expect(testUI.Out).To(Say("1             Something             false        17E0E587-0E53-4A6E-B6AE-82073159F910   2020-03-04T13:23:32Z"))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))

					Expect(fakeActor.GetRevisionsByApplicationNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetRevisionsByApplicationNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})

				It("does not display an informative message", func() {
					Expect(testUI.Out).NotTo(Say("Info: this app is in the middle of a rolling deployment. More than one revision is deployed."))
				})

				When("there is more than one revision deployed", func() {
					BeforeEach(func() {
						deployedRevisions := []resources.Revision{
							{
								Version:     2,
								GUID:        "A89F8259-D32B-491A-ABD6-F100AC42D74C",
								Description: "Something else",
								CreatedAt:   "2020-03-08T12:43:30Z",
								Deployable:  true,
							},
							{
								Version:     3,
								GUID:        "A68F13F7-7E5E-4411-88E8-1FAC54F73F50",
								Description: "On a different note",
								CreatedAt:   "2020-03-10T17:11:58Z",
								Deployable:  true,
							},
						}
						fakeActor.GetApplicationRevisionsDeployedReturns(deployedRevisions, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
					})

					It("marks both as deployed", func() {
						Expect(testUI.Out).To(Say("3\\(deployed\\)   On a different note   true         A68F13F7-7E5E-4411-88E8-1FAC54F73F50   2020-03-10T17:11:58Z"))
						Expect(testUI.Out).To(Say("2\\(deployed\\)   Something else        true         A89F8259-D32B-491A-ABD6-F100AC42D74C   2020-03-08T12:43:30Z"))
					})
					It("displays an informative message", func() {
						Expect(testUI.Out).To(Say("Info: this app is in the middle of a rolling deployment. More than one revision is deployed."))
					})
				})

				When("the revisions feature is disabled on the app", func() {
					BeforeEach(func() {
						revisionsFeature := resources.ApplicationFeature{
							Name:    "revisions",
							Enabled: false,
						}
						fakeActor.GetAppFeatureReturns(revisionsFeature, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
					})

					It("displays the revisions with a warning", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`Getting revisions for app some-app in org some-org / space some-space as banana\.\.\.`))
						Expect(testUI.Err).To(Say(`Warning: Revisions for app 'some-app' are disabled. Updates to the app will not create new revisions.`))

						Expect(testUI.Out).To(Say("revision      description           deployable   revision guid                          created at"))
						Expect(testUI.Out).To(Say("3\\(deployed\\)   On a different note   true         A68F13F7-7E5E-4411-88E8-1FAC54F73F50   2020-03-10T17:11:58Z"))
						Expect(testUI.Out).To(Say("2             Something else        true         A89F8259-D32B-491A-ABD6-F100AC42D74C   2020-03-08T12:43:30Z"))
						Expect(testUI.Out).To(Say("1             Something             false        17E0E587-0E53-4A6E-B6AE-82073159F910   2020-03-04T13:23:32Z"))

						Expect(testUI.Err).To(Say("get-warning-1"))
						Expect(testUI.Err).To(Say("get-warning-2"))

						Expect(fakeActor.GetRevisionsByApplicationNameAndSpaceCallCount()).To(Equal(1))
						appName, spaceGUID := fakeActor.GetRevisionsByApplicationNameAndSpaceArgsForCall(0)
						Expect(appName).To(Equal("some-app"))
						Expect(spaceGUID).To(Equal("some-space-guid"))
					})

					When("the app is in the STOPPED state", func() {
						BeforeEach(func() {
							fakeApp := resources.Application{
								GUID:  "fake-guid",
								Name:  "app-name",
								State: constant.ApplicationStopped,
							}
							fakeActor.GetApplicationByNameAndSpaceReturns(fakeApp, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
						})

						It("displays the revisions with an info message about being unable to determine the deployed revision", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say(`Getting revisions for app some-app in org some-org / space some-space as banana\.\.\.`))
							Expect(testUI.Out).To(Say(`Info: this app is in a stopped state. It is not possible to determine which revision is currently deployed.`))
						})

					})
				})

				When("Application Revisions deployed call fails", func() {
					var expectedErr error
					BeforeEach(func() {
						expectedErr = errors.New("some-error")
						fakeActor.GetApplicationRevisionsDeployedReturns(
							[]resources.Revision{},
							v7action.Warnings{"get-warning-1", "get-warning-2"},
							expectedErr,
						)
					})

					It("returns the error", func() {
						Expect(executeErr).To(Equal(expectedErr))
						Expect(testUI.Out).To(Say(`Getting revisions for app some-app in org some-org / space some-space as banana\.\.\.`))

						Expect(testUI.Err).To(Say("get-warning-1"))
						Expect(testUI.Err).To(Say("get-warning-2"))

					})
				})
			})

			When("there are no revisions available", func() {
				BeforeEach(func() {
					fakeActor.GetRevisionsByApplicationNameAndSpaceReturns(
						[]resources.Revision{},
						v7action.Warnings{"get-warning-1", "get-warning-2"},
						nil,
					)
				})

				It("returns 'no revisions found'", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(strings.TrimSpace(string(out.Contents()))).To(Equal(strings.TrimSpace(`
Getting revisions for app some-app in org some-org / space some-space as banana...

No revisions found
`)))
					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))
				})
			})

			When("revisions variables returns an unknown error", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeActor.GetRevisionsByApplicationNameAndSpaceReturns([]resources.Revision{}, v7action.Warnings{"get-warning-1", "get-warning-2"}, expectedErr)
				})

				It("returns the error", func() {
					Expect(executeErr).To(Equal(expectedErr))
					Expect(testUI.Out).To(Say(`Getting revisions for app some-app in org some-org / space some-space as banana\.\.\.`))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))
				})
			})
		})
	})
})
