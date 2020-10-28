package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
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
		Expect(cmd.Execute(nil)).To(Succeed())
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
				var revision resources.Revision
				BeforeEach(func() {
					fakeApp := resources.Application{
						GUID: "fake-guid",
						Name: "some-app",
					}
					fakeActor.GetApplicationByNameAndSpaceReturns(fakeApp, nil, nil)

					revision = resources.Revision{
						Version:     3,
						GUID:        "A68F13F7-7E5E-4411-88E8-1FAC54F73F50",
						Description: "On a different note",
						CreatedAt:   "2020-03-10T17:11:58Z",
						Deployable:  true,
						Droplet: resources.Droplet{
							GUID: "droplet-guid",
						},
						Links: resources.APILinks{
							"environment_variables": resources.APILink{
								HREF: "revision-environment-variables-link-3",
							},
						},
					}
					fakeActor.GetRevisionByApplicationAndVersionReturns(revision, nil, nil)
					fakeActor.GetApplicationByNameAndSpaceReturns(resources.Application{GUID: "app-guid"}, nil, nil)
					fakeActor.GetApplicationRevisionsDeployedReturns([]resources.Revision{revision}, nil, nil)

					environmentVariableGroup := v7action.EnvironmentVariableGroup{
						"foo": *types.NewFilteredString("bar3"),
					}
					fakeActor.GetEnvironmentVariableGroupByRevisionReturns(
						environmentVariableGroup,
						true,
						nil,
						nil,
					)

					cmd.Version = flag.Revision{NullInt: types.NullInt{Value: 3, IsSet: true}}
				})

				It("gets the app guid", func() {
					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})

				It("retrieves the requested revision for the app", func() {
					Expect(fakeActor.GetRevisionByApplicationAndVersionCallCount()).To(Equal(1))
					appGUID, version := fakeActor.GetRevisionByApplicationAndVersionArgsForCall(0)
					Expect(appGUID).To(Equal("app-guid"))
					Expect(version).To(Equal(3))
				})

				It("retrieves the deployed revisions", func() {
					Expect(fakeActor.GetApplicationRevisionsDeployedCallCount()).To(Equal(1))
					Expect(fakeActor.GetApplicationRevisionsDeployedArgsForCall(0)).To(Equal("app-guid"))
				})

				It("retrieves the environment variables for the revision", func() {
					Expect(fakeActor.GetEnvironmentVariableGroupByRevisionCallCount()).To(Equal(1))
					Expect(fakeActor.GetEnvironmentVariableGroupByRevisionArgsForCall(0)).To(Equal(
						revision,
					))
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

					Expect(testUI.Out).To(Say(`application environment variables:`))
					Expect(testUI.Out).To(Say(`foo:   bar3`))

					// Expect(testUI.Out).To(Say(`labels:`))
					// Expect(testUI.Out).To(Say(`label: foo3`))

					// Expect(testUI.Out).To(Say(`annotations:`))
					// Expect(testUI.Out).To(Say(`annotation: foo3`))
				})

				When("there is not a environment_variables link provided", func() {
					BeforeEach(func() {
						revision = resources.Revision{
							Version:     3,
							GUID:        "A68F13F7-7E5E-4411-88E8-1FAC54F73F50",
							Description: "No env var link",
							CreatedAt:   "2020-03-10T17:11:58Z",
							Deployable:  true,
							Droplet: resources.Droplet{
								GUID: "droplet-guid",
							},
							Links: resources.APILinks{},
						}
						fakeActor.GetRevisionByApplicationAndVersionReturns(revision, nil, nil)
						fakeActor.GetApplicationRevisionsDeployedReturns([]resources.Revision{revision}, nil, nil)
						fakeActor.GetEnvironmentVariableGroupByRevisionReturns(nil, false, v7action.Warnings{"warn-env-var"}, nil)
					})

					It("warns the user it will not display env vars ", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Err).To(Say("warn-env-var"))
						Expect(testUI.Out).ToNot(Say("application environment variables:"))
					})
				})

				When("revision is not deployed", func() {
					BeforeEach(func() {
						revisionDeployed := resources.Revision{
							Version:     12345,
							GUID:        "Fake-guid",
							Description: "derployed and definitely not your revision",
							CreatedAt:   "2020-03-10T17:11:58Z",
							Deployable:  true,
							Droplet: resources.Droplet{
								GUID: "droplet-guid",
							},
						}
						fakeActor.GetApplicationRevisionsDeployedReturns([]resources.Revision{revisionDeployed}, nil, nil)
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
