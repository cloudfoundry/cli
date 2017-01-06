package v2_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("target Command", func() {
	var (
		cmd        v2.TargetCommand
		testUI     *ui.UI
		fakeActor  *v2fakes.FakeTargetActor
		fakeConfig *commandfakes.FakeConfig
		input      *Buffer
		executeErr error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeActor = new(v2fakes.FakeTargetActor)
		fakeConfig = new(commandfakes.FakeConfig)

		cmd = v2.TargetCommand{
			UI:     testUI,
			Actor:  fakeActor,
			Config: fakeConfig,
		}
		fakeConfig.BinaryNameReturns("faceman")
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking the minimum CLI version", func() {
		var (
			minCLIVersion string
			binaryVersion string
			apiVersion    string
		)

		BeforeEach(func() {
			apiVersion = "6.0.0"
			minCLIVersion = "1.0.0"
			fakeConfig.APIVersionReturns(apiVersion)
			fakeConfig.MinCLIVersionReturns(minCLIVersion)
		})

		Context("when the CLI version is less than the recommended minimum", func() {
			BeforeEach(func() {
				binaryVersion = "0.0.0"
				fakeConfig.BinaryVersionReturns(binaryVersion)
			})

			It("displays a recommendation to update the CLI version", func() {
				Expect(testUI.Out).To(Say(fmt.Sprintf("Cloud Foundry API version %s requires CLI version %s. You are currently on version %s. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads", apiVersion, minCLIVersion, binaryVersion)))
			})
		})

		Context("when an error is encountered while parsing the semver versions", func() {
			BeforeEach(func() {
				fakeConfig.BinaryVersionReturns("&#%")
			})

			It("does not recommend to update the CLI version", func() {
				Consistently(testUI.Out).ShouldNot(Say(fmt.Sprintf("Cloud Foundry API version %s requires CLI version %s.", apiVersion, minCLIVersion)))
			})
		})
	})

	Context("when the user is not logged in", func() {
		It("returns an error", func() {
			Expect(executeErr).To(MatchError(command.NotLoggedInError{
				BinaryName: "faceman",
			}))
		})
	})

	Context("when config.CurrentUser returns an error", func() {
		var someError error

		BeforeEach(func() {
			someError = errors.New("some-current-user-error")
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.CurrentUserReturns(configv3.User{}, someError)
		})

		It("returns a current user error", func() {
			expectedError := shared.CurrentUserError{
				Message: "some-current-user-error",
			}
			Expect(executeErr).To(MatchError(expectedError))
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("some-api-target")
			fakeConfig.APIVersionReturns("1.2.3")
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
			fakeConfig.CurrentUserReturns(configv3.User{
				Name: "some-user",
			}, nil)
		})

		Context("when no arguments are given", func() {
			Context("when no org or space is targeted", func() {
				BeforeEach(func() {
					fakeConfig.TargetedOrganizationReturns(configv3.Organization{})
					fakeConfig.TargetedSpaceReturns(configv3.Space{})
				})

				It("displays no org or space targeted", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("API endpoint:   some-api-target"))
					Expect(testUI.Out).To(Say("API version:    1.2.3"))
					Expect(testUI.Out).To(Say("User:           some-user"))
					Expect(testUI.Out).To(Say("No org or space targeted, use 'faceman target -o ORG -s SPACE'"))
				})
			})

			Context("when an org but no space is targeted", func() {
				BeforeEach(func() {
					fakeConfig.TargetedOrganizationReturns(configv3.Organization{
						GUID: "some-org-guid",
						Name: "some-org",
					})
					fakeConfig.TargetedSpaceReturns(configv3.Space{})
				})

				It("displays the org and no space targeted ", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("API endpoint:   some-api-target"))
					Expect(testUI.Out).To(Say("API version:    1.2.3"))
					Expect(testUI.Out).To(Say("User:           some-user"))
					Expect(testUI.Out).To(Say("Org:            some-org"))
					Expect(testUI.Out).To(Say("Space:          No space targeted, use 'faceman target -s SPACE'"))
				})
			})

			Context("when an org and space is targeted", func() {
				BeforeEach(func() {
					fakeConfig.TargetedOrganizationReturns(configv3.Organization{
						GUID: "some-org-guid",
						Name: "some-org",
					})
					fakeConfig.TargetedSpaceReturns(configv3.Space{
						GUID: "some-space-guid",
						Name: "some-space",
					})
				})

				It("displays the org and space targeted ", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("API endpoint:   some-api-target"))
					Expect(testUI.Out).To(Say("API version:    1.2.3"))
					Expect(testUI.Out).To(Say("User:           some-user"))
					Expect(testUI.Out).To(Say("Org:            some-org"))
					Expect(testUI.Out).To(Say("Space:          some-space"))
				})
			})
		})

		Context("when just a space argument is given", func() {
			BeforeEach(func() {
				cmd.Space = "some-space"
			})

			Context("when an org is already targeted", func() {
				BeforeEach(func() {
					fakeActor.GetSpaceByNameReturns(v2action.Space{
						GUID:     "some-space-guid",
						Name:     "some-space",
						AllowSSH: true,
					}, v2action.Warnings{}, nil)
					fakeConfig.TargetedOrganizationReturns(configv3.Organization{
						GUID: "some-org-guid",
					})
				})

				It("sets the space in the config", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(1))
					spaceGUID, spaceName, spaceAllowSSH := fakeConfig.SetSpaceInformationArgsForCall(0)
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(spaceName).To(Equal("some-space"))
					Expect(spaceAllowSSH).To(Equal(true))
				})
			})

			Context("when no org is targeted", func() {
				It("returns NoOrgTargeted error", func() {
					Expect(executeErr).To(MatchError(shared.NoOrgTargetedError{}))
					Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(0))
				})
			})
		})

		Context("when just an org argument is given", func() {
			BeforeEach(func() {
				cmd.Organization = "some-org"
			})

			Context("when getting the org returns an error", func() {
				var err error

				BeforeEach(func() {
					err = errors.New("get-org-error")
					fakeActor.GetOrganizationByNameReturns(
						v2action.Organization{},
						v2action.Warnings{
							"warning-1",
							"warning-2",
						},
						err)
				})

				It("displays all warnings and returns an org target error", func() {
					Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
					orgName := fakeActor.GetOrganizationByNameArgsForCall(0)
					Expect(orgName).To(Equal("some-org"))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))

					expectedError := shared.OrgTargetError{
						Message: "get-org-error",
					}
					Expect(executeErr).To(MatchError(expectedError))
				})
			})

			Context("when getting the org does not return an error", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationByNameReturns(
						v2action.Organization{
							GUID: "some-org-guid",
						},
						v2action.Warnings{
							"warning-1",
							"warning-2",
						},
						nil)
					fakeConfig.TargetedOrganizationReturns(configv3.Organization{
						GUID: "some-org-guid",
						Name: "some-org",
					})
				})

				Context("when GetOrganizationSpaces returns an error", func() {
					var err error
					BeforeEach(func() {
						err = errors.New("get-org-spaces-error")
						fakeActor.GetOrganizationSpacesReturns(
							[]v2action.Space{},
							v2action.Warnings{
								"warning-3",
							}, err)
					})

					It("displays all warnings and returns a get org spaces error", func() {
						Expect(fakeActor.GetOrganizationSpacesCallCount()).To(Equal(1))
						orgGUID := fakeActor.GetOrganizationSpacesArgsForCall(0)
						Expect(orgGUID).To(Equal("some-org-guid"))

						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))
						Expect(testUI.Err).To(Say("warning-3"))

						expectedError := shared.GetOrgSpacesError{
							Message: "get-org-spaces-error",
						}
						Expect(executeErr).To(MatchError(expectedError))
					})
				})

				Context("when there are 0 spaces in the org", func() {
					It("displays all warnings and sets the org in the config", func() {
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

				Context("when there is only 1 space in the org", func() {
					BeforeEach(func() {
						fakeActor.GetOrganizationSpacesReturns([]v2action.Space{{
							GUID:     "some-space-guid",
							Name:     "some-space",
							AllowSSH: true,
						}}, v2action.Warnings{
							"warning-3",
						}, nil)
					})

					It("displays all warnings and sets the org and space in the config", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))
						Expect(testUI.Err).To(Say("warning-3"))

						Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
						orgGUID, orgName := fakeConfig.SetOrganizationInformationArgsForCall(0)
						Expect(orgGUID).To(Equal("some-org-guid"))
						Expect(orgName).To(Equal("some-org"))

						Expect(fakeConfig.UnsetSpaceInformationCallCount()).To(Equal(1))
						Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(1))
						spaceGUID, spaceName, spaceAllowSSH := fakeConfig.SetSpaceInformationArgsForCall(0)
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(spaceName).To(Equal("some-space"))
						Expect(spaceAllowSSH).To(Equal(true))
					})
				})

				Context("when there are more than 1 spaces in the org", func() {
					BeforeEach(func() {
						fakeActor.GetOrganizationSpacesReturns([]v2action.Space{{
							GUID:     "some-space-guid",
							Name:     "some-space",
							AllowSSH: true,
						}, {
							GUID:     "another-space-space-guid",
							Name:     "another-space",
							AllowSSH: true,
						}}, v2action.Warnings{
							"warning-3",
						}, nil)
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
			})
		})

		Context("when org and space arguments are given and org is legit", func() {
			BeforeEach(func() {
				cmd.Space = "some-space"
				cmd.Organization = "some-org"

				fakeActor.GetOrganizationByNameReturns(
					v2action.Organization{
						GUID: "some-org-guid",
					}, nil, nil)
			})

			It("targets the org correctly", func() {
				Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
				orgGUID, orgName := fakeConfig.SetOrganizationInformationArgsForCall(0)
				Expect(orgGUID).To(Equal("some-org-guid"))
				Expect(orgName).To(Equal("some-org"))

				Expect(fakeConfig.UnsetSpaceInformationCallCount()).To(Equal(1))
			})

			Context("when getting the space returns an error", func() {
				BeforeEach(func() {
					fakeConfig.TargetedOrganizationReturns(configv3.Organization{
						GUID: "some-org-guid",
						Name: "some-org",
					})

					fakeActor.GetSpaceByNameReturns(v2action.Space{
						GUID: "some-space-guid",
					},
						v2action.Warnings{
							"warning-1",
							"warning-2",
						}, errors.New("get-space-by-name-error"))
				})

				It("displays all warnings and returns a space target error", func() {
					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))

					expectedError := shared.SpaceTargetError{
						Message:   "get-space-by-name-error",
						SpaceName: "some-space",
					}
					Expect(executeErr).To(MatchError(expectedError))
				})

				Context("when there is 1 space in the org", func() {
					BeforeEach(func() {
						fakeActor.GetOrganizationSpacesReturns([]v2action.Space{{
							GUID:     "some-space-guid",
							Name:     "some-space",
							AllowSSH: true,
						}}, nil, nil)
					})

					It("does not auto target the space since space name was provided as an argument", func() {
						// SetSpaceInformation does not get called from autoTargetSpace, and the call count is 0 because we intentionally exit early.
						Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(0))
					})
				})
			})

			Context("when getting the space does not return an error", func() {
				BeforeEach(func() {
					fakeActor.GetSpaceByNameReturns(v2action.Space{
						GUID:     "some-space-guid",
						Name:     "some-space",
						AllowSSH: true,
					},
						v2action.Warnings{
							"warning-1",
							"warning-2",
						}, nil)
					fakeConfig.TargetedOrganizationReturns(configv3.Organization{
						GUID: "some-org-guid",
					})
				})

				It("displays all warnings and sets the space in the config", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))

					Expect(fakeActor.GetSpaceByNameCallCount()).To(Equal(1))
					orgGUID, spaceName := fakeActor.GetSpaceByNameArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(spaceName).To(Equal("some-space"))

					Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(1))
					spaceGUID, spaceName, spaceAllowSSH := fakeConfig.SetSpaceInformationArgsForCall(0)
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(spaceName).To(Equal("some-space"))
					Expect(spaceAllowSSH).To(Equal(true))
				})
			})
		})
	})
})
