package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("target Command", func() {
	var (
		cmd             v7.TargetCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		apiVersion      string
		minCLIVersion   string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = v7.TargetCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		apiVersion = "1.2.3"
		fakeActor.CloudControllerAPIVersionReturns(apiVersion)
		minCLIVersion = "1.0.0"
		fakeConfig.MinCLIVersionReturns(minCLIVersion)
		fakeConfig.BinaryVersionReturns("1.0.0")
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("a cloud controller API endpoint is set", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("some-api-target")
		})

		When("checking target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeFalse())
				Expect(checkTargetedSpace).To(BeFalse())

				Expect(fakeConfig.UnsetOrganizationAndSpaceInformationCallCount()).To(Equal(0))
				Expect(fakeConfig.UnsetSpaceInformationCallCount()).To(Equal(0))
			})
		})

		When("the user is logged in", func() {
			When("getting the current user returns an error", func() {
				var someErr error

				BeforeEach(func() {
					someErr = errors.New("some-current-user-error")
					fakeConfig.CurrentUserReturns(configv3.User{}, someErr)
				})

				It("returns the same error", func() {
					Expect(executeErr).To(MatchError(someErr))

					Expect(fakeConfig.UnsetOrganizationAndSpaceInformationCallCount()).To(Equal(0))
					Expect(fakeConfig.UnsetSpaceInformationCallCount()).To(Equal(0))
				})
			})

			When("getting the current user does not return an error", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserReturns(
						configv3.User{Name: "some-user"},
						nil)
				})

				When("no arguments are provided", func() {
					When("no org or space are targeted", func() {
						It("displays how to target an org and space", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("api endpoint:   some-api-target"))
							Expect(testUI.Out).To(Say("api version:    1.2.3"))
							Expect(testUI.Out).To(Say("user:           some-user"))
							Expect(testUI.Out).To(Say("No org or space targeted, use '%s target -o ORG -s SPACE'", binaryName))
						})
					})

					When("an org but no space is targeted", func() {
						BeforeEach(func() {
							fakeConfig.HasTargetedOrganizationReturns(true)
							fakeConfig.TargetedOrganizationReturns(configv3.Organization{
								GUID: "some-org-guid",
								Name: "some-org",
							})
						})

						It("displays the org and tip to target space", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("api endpoint:   some-api-target"))
							Expect(testUI.Out).To(Say("api version:    1.2.3"))
							Expect(testUI.Out).To(Say("user:           some-user"))
							Expect(testUI.Out).To(Say("org:            some-org"))
							Expect(testUI.Out).To(Say("No space targeted, use '%s target -s SPACE'", binaryName))
						})
					})

					When("an org and space are targeted", func() {
						BeforeEach(func() {
							fakeConfig.HasTargetedOrganizationReturns(true)
							fakeConfig.TargetedOrganizationReturns(configv3.Organization{
								GUID: "some-org-guid",
								Name: "some-org",
							})
							fakeConfig.HasTargetedSpaceReturns(true)
							fakeConfig.TargetedSpaceReturns(configv3.Space{
								GUID: "some-space-guid",
								Name: "some-space",
							})
						})

						It("displays the org and space targeted ", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("api endpoint:   some-api-target"))
							Expect(testUI.Out).To(Say("api version:    1.2.3"))
							Expect(testUI.Out).To(Say("user:           some-user"))
							Expect(testUI.Out).To(Say("org:            some-org"))
							Expect(testUI.Out).To(Say("space:          some-space"))
						})
					})
				})

				When("space is provided", func() {
					BeforeEach(func() {
						cmd.Space = "some-space"
					})

					When("an org is already targeted", func() {
						BeforeEach(func() {
							fakeConfig.HasTargetedOrganizationReturns(true)
							fakeConfig.TargetedOrganizationReturns(configv3.Organization{GUID: "some-org-guid"})
						})

						When("the space exists", func() {
							BeforeEach(func() {
								fakeActor.GetSpaceByNameAndOrganizationReturns(
									v7action.Space{
										GUID: "some-space-guid",
										Name: "some-space",
									},
									v7action.Warnings{},
									nil)
							})

							It("targets the space", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(fakeConfig.V7SetSpaceInformationCallCount()).To(Equal(1))
								spaceGUID, spaceName := fakeConfig.V7SetSpaceInformationArgsForCall(0)
								Expect(spaceGUID).To(Equal("some-space-guid"))
								Expect(spaceName).To(Equal("some-space"))
							})
						})

						When("the space does not exist", func() {
							BeforeEach(func() {
								fakeActor.GetSpaceByNameAndOrganizationReturns(
									v7action.Space{},
									v7action.Warnings{},
									actionerror.SpaceNotFoundError{Name: "some-space"})
							})

							It("returns a SpaceNotFoundError and clears existing space", func() {
								Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{Name: "some-space"}))

								Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(0))
								Expect(fakeConfig.UnsetOrganizationAndSpaceInformationCallCount()).To(Equal(0))
							})
						})
					})

					When("no org is targeted", func() {
						It("returns NoOrgTargeted error and clears existing space", func() {
							Expect(executeErr).To(MatchError(translatableerror.NoOrganizationTargetedError{BinaryName: "faceman"}))

							Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(0))
							Expect(fakeConfig.UnsetOrganizationAndSpaceInformationCallCount()).To(Equal(0))
						})
					})
				})

				When("org is provided", func() {
					BeforeEach(func() {
						cmd.Organization = "some-org"
					})

					When("the org does not exist", func() {
						BeforeEach(func() {
							fakeActor.GetOrganizationByNameReturns(
								v7action.Organization{},
								nil,
								actionerror.OrganizationNotFoundError{Name: "some-org"})
						})

						It("displays all warnings,returns an org target error, and clears existing targets", func() {
							Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org"}))

							Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(0))
							Expect(fakeConfig.UnsetOrganizationAndSpaceInformationCallCount()).To(Equal(1))
						})
					})

					When("the org exists", func() {
						BeforeEach(func() {
							fakeConfig.HasTargetedOrganizationReturns(true)
							fakeConfig.TargetedOrganizationReturns(configv3.Organization{
								GUID: "some-org-guid",
								Name: "some-org",
							})
							fakeActor.GetOrganizationByNameReturns(
								v7action.Organization{GUID: "some-org-guid"},
								v7action.Warnings{"warning-1", "warning-2"},
								nil)
						})

						When("getting the organization's spaces returns an error", func() {
							var err error

							BeforeEach(func() {
								err = errors.New("get-org-spaces-error")
								fakeActor.GetOrganizationSpacesReturns(
									[]v7action.Space{},
									v7action.Warnings{
										"warning-3",
									},
									err)
							})

							It("displays all warnings, returns a get org spaces error and clears existing targets", func() {
								Expect(executeErr).To(MatchError(err))

								Expect(fakeActor.GetOrganizationSpacesCallCount()).To(Equal(1))
								orgGUID := fakeActor.GetOrganizationSpacesArgsForCall(0)
								Expect(orgGUID).To(Equal("some-org-guid"))

								Expect(testUI.Err).To(Say("warning-1"))
								Expect(testUI.Err).To(Say("warning-2"))
								Expect(testUI.Err).To(Say("warning-3"))

								Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
								orgGUID, orgName := fakeConfig.SetOrganizationInformationArgsForCall(0)
								Expect(orgGUID).To(Equal("some-org-guid"))
								Expect(orgName).To(Equal("some-org"))
								Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(0))

								Expect(fakeConfig.UnsetOrganizationAndSpaceInformationCallCount()).To(Equal(1))
							})
						})

						When("there are no spaces in the targeted org", func() {
							It("displays all warnings", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Err).To(Say("warning-1"))
								Expect(testUI.Err).To(Say("warning-2"))
							})

							It("sets the org and unsets the space in the config", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
								orgGUID, orgName := fakeConfig.SetOrganizationInformationArgsForCall(0)
								Expect(orgGUID).To(Equal("some-org-guid"))
								Expect(orgName).To(Equal("some-org"))

								Expect(fakeConfig.UnsetSpaceInformationCallCount()).To(Equal(1))
								Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(0))
							})
						})

						When("there is only 1 space in the targeted org", func() {
							BeforeEach(func() {
								fakeActor.GetOrganizationSpacesReturns(
									[]v7action.Space{{
										GUID: "some-space-guid",
										Name: "some-space",
									}},
									v7action.Warnings{"warning-3"},
									nil,
								)
							})

							It("displays all warnings", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Err).To(Say("warning-1"))
								Expect(testUI.Err).To(Say("warning-2"))
								Expect(testUI.Err).To(Say("warning-3"))
							})

							It("targets the org and space", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
								orgGUID, orgName := fakeConfig.SetOrganizationInformationArgsForCall(0)
								Expect(orgGUID).To(Equal("some-org-guid"))
								Expect(orgName).To(Equal("some-org"))

								Expect(fakeConfig.UnsetSpaceInformationCallCount()).To(Equal(1))

								Expect(fakeConfig.V7SetSpaceInformationCallCount()).To(Equal(1))
								spaceGUID, spaceName := fakeConfig.V7SetSpaceInformationArgsForCall(0)
								Expect(spaceGUID).To(Equal("some-space-guid"))
								Expect(spaceName).To(Equal("some-space"))
							})
						})

						When("there are multiple spaces in the targeted org", func() {
							BeforeEach(func() {
								fakeActor.GetOrganizationSpacesReturns(
									[]v7action.Space{
										{
											GUID: "some-space-guid",
											Name: "some-space",
										},
										{
											GUID: "another-space-space-guid",
											Name: "another-space",
										},
									},
									v7action.Warnings{"warning-3"},
									nil,
								)
							})

							It("displays all warnings, sets the org, and clears the existing targetted space from the config", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Err).To(Say("warning-1"))
								Expect(testUI.Err).To(Say("warning-2"))

								Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
								orgGUID, orgName := fakeConfig.SetOrganizationInformationArgsForCall(0)
								Expect(orgGUID).To(Equal("some-org-guid"))
								Expect(orgName).To(Equal("some-org"))

								Expect(fakeConfig.UnsetSpaceInformationCallCount()).To(Equal(1))
								Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(0))
							})
						})

						When("getting the spaces in org returns an error", func() {
							var err error

							BeforeEach(func() {
								err = errors.New("get-org-spaces-error")
								fakeActor.GetOrganizationSpacesReturns(
									[]v7action.Space{},
									v7action.Warnings{
										"warning-3",
									},
									err)
							})

							It("displays all warnings, returns the error, and clears existing targets", func() {
								Expect(executeErr).To(MatchError(err))

								Expect(testUI.Err).To(Say("warning-1"))
								Expect(testUI.Err).To(Say("warning-2"))
								Expect(testUI.Err).To(Say("warning-3"))

								Expect(fakeConfig.UnsetOrganizationAndSpaceInformationCallCount()).To(Equal(1))
							})
						})
					})
				})

				When("org and space arguments are provided", func() {
					BeforeEach(func() {
						cmd.Space = "some-space"
						cmd.Organization = "some-org"
					})

					When("the org exists", func() {
						BeforeEach(func() {
							fakeActor.GetOrganizationByNameReturns(
								v7action.Organization{
									GUID: "some-org-guid",
									Name: "some-org",
								},
								v7action.Warnings{
									"warning-1",
								},
								nil)

							fakeConfig.HasTargetedOrganizationReturns(true)
							fakeConfig.TargetedOrganizationReturns(configv3.Organization{
								GUID: "some-org-guid",
								Name: "some-org",
							})

						})

						When("the space exists", func() {
							BeforeEach(func() {
								fakeActor.GetSpaceByNameAndOrganizationReturns(
									v7action.Space{
										GUID: "some-space-guid",
										Name: "some-space",
									},
									v7action.Warnings{
										"warning-2",
									},
									nil)
							})

							It("sets the target org and space", func() {
								Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
								orgGUID, orgName := fakeConfig.SetOrganizationInformationArgsForCall(0)
								Expect(orgGUID).To(Equal("some-org-guid"))
								Expect(orgName).To(Equal("some-org"))

								Expect(fakeConfig.V7SetSpaceInformationCallCount()).To(Equal(1))
								spaceGUID, spaceName := fakeConfig.V7SetSpaceInformationArgsForCall(0)
								Expect(spaceGUID).To(Equal("some-space-guid"))
								Expect(spaceName).To(Equal("some-space"))
							})

							It("displays all warnings", func() {
								Expect(testUI.Err).To(Say("warning-1"))
								Expect(testUI.Err).To(Say("warning-2"))
							})
						})

						When("the space does not exist", func() {
							BeforeEach(func() {
								fakeActor.GetSpaceByNameAndOrganizationReturns(
									v7action.Space{},
									nil,
									actionerror.SpaceNotFoundError{Name: "some-space"})
							})

							It("returns an error and clears existing targets", func() {
								Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{Name: "some-space"}))

								Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
								Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(0))

								Expect(fakeConfig.UnsetOrganizationAndSpaceInformationCallCount()).To(Equal(1))
							})
						})
					})

					When("the org does not exist", func() {
						BeforeEach(func() {
							fakeActor.GetOrganizationByNameReturns(
								v7action.Organization{},
								nil,
								actionerror.OrganizationNotFoundError{Name: "some-org"})
						})

						It("returns an error and clears existing targets", func() {
							Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org"}))

							Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(0))
							Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(0))

							Expect(fakeConfig.UnsetOrganizationAndSpaceInformationCallCount()).To(Equal(1))
						})
					})
				})
			})
		})
	})
})
