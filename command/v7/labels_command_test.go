package v7_test

import (
	"errors"
	"regexp"

	"strings"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
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
		fakeLabelsActor *v7fakes.FakeLabelsActor
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		testUI          *ui.UI

		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeLabelsActor = new(v7fakes.FakeLabelsActor)
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		cmd = LabelsCommand{
			Actor:       fakeLabelsActor,
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	verifyStackArgNotAllowed := func() {
		BeforeEach(func() {
			cmd.BuildpackStack = "cflinuxfs3"
		})

		It("displays an argument combination error", func() {
			argumentCombinationError := translatableerror.ArgumentCombinationError{
				Args: []string{strings.ToLower(cmd.RequiredArgs.ResourceType), "--stack, -s"},
			}

			Expect(executeErr).To(MatchError(argumentCombinationError))
		})
	}

	Describe("listing labels", func() {
		Describe("for apps", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserNameReturns("some-user", nil)
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org"})
				fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "fake-space", GUID: "some-space-guid"})
				cmd.RequiredArgs = flag.LabelsArgs{
					ResourceType: "app",
					ResourceName: "dora",
				}
				fakeLabelsActor.GetApplicationLabelsReturns(
					map[string]types.NullString{
						"some-other-label": types.NewNullString("some-other-value"),
						"some-label":       types.NewNullString("some-value"),
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
				Expect(fakeLabelsActor.GetApplicationLabelsCallCount()).To(Equal(1))
				appName, spaceGUID := fakeLabelsActor.GetApplicationLabelsArgsForCall(0)
				Expect(appName).To(Equal("dora"))
				Expect(spaceGUID).To(Equal("some-space-guid"))
			})

			It("displays the labels that are associated with the application, alphabetically", func() {
				Expect(testUI.Out).To(Say(`key\s+value`))
				Expect(testUI.Out).To(Say(`some-label\s+some-value`))
				Expect(testUI.Out).To(Say(`some-other-label\s+some-other-value`))
			})

			When("CAPI returns warnings", func() {
				BeforeEach(func() {
					fakeLabelsActor.GetApplicationLabelsReturns(
						map[string]types.NullString{
							"some-other-label": types.NewNullString("some-other-value"),
							"some-label":       types.NewNullString("some-value"),
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
					fakeLabelsActor.GetApplicationLabelsReturns(
						map[string]types.NullString{},
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

			When("when the --stack flag is specified", func() {
				verifyStackArgNotAllowed()
			})
		})

		Describe("for orgs", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserNameReturns("some-user", nil)
				cmd.RequiredArgs = flag.LabelsArgs{
					ResourceType: "org",
					ResourceName: "fake-org",
				}
				fakeLabelsActor.GetOrganizationLabelsReturns(
					map[string]types.NullString{
						"some-other-label": types.NewNullString("some-other-value"),
						"some-label":       types.NewNullString("some-value"),
					},
					v7action.Warnings{},
					nil)
			})

			It("doesn't error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
			})

			It("checks that the user is logged in", func() {
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkOrg).To(BeFalse())
				Expect(checkSpace).To(BeFalse())
			})

			It("displays a message that it is retrieving the labels", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Getting labels for org fake-org as some-user...`)))
			})

			It("retrieves the labels associated with the organization", func() {
				Expect(fakeLabelsActor.GetOrganizationLabelsCallCount()).To(Equal(1))
			})

			It("displays the labels that are associated with the organization, alphabetically", func() {
				Expect(testUI.Out).To(Say(`key\s+value`))
				Expect(testUI.Out).To(Say(`some-label\s+some-value`))
				Expect(testUI.Out).To(Say(`some-other-label\s+some-other-value`))
			})

			When("CAPI returns warnings", func() {
				BeforeEach(func() {
					fakeLabelsActor.GetOrganizationLabelsReturns(
						map[string]types.NullString{
							"some-other-label": types.NewNullString("some-other-value"),
							"some-label":       types.NewNullString("some-value"),
						},
						v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
						nil)
				})

				It("prints all warnings", func() {
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
				})
			})

			When("there is an error retrieving the organization", func() {
				BeforeEach(func() {
					fakeLabelsActor.GetOrganizationLabelsReturns(
						map[string]types.NullString{},
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

			When("the resource type argument is not lowercase", func() {
				BeforeEach(func() {
					cmd.RequiredArgs = flag.LabelsArgs{
						ResourceType: "OrG",
						ResourceName: "fake-org",
					}
				})

				It("retrieves the labels associated with the org", func() {
					Expect(fakeLabelsActor.GetOrganizationLabelsCallCount()).To(Equal(1))
					orgName := fakeLabelsActor.GetOrganizationLabelsArgsForCall(0)
					Expect(orgName).To(Equal("fake-org"))
				})
			})

			When("when the --stack flag is specified", func() {
				verifyStackArgNotAllowed()
			})
		})

		Describe("for spaces", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserNameReturns("some-user", nil)
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org", GUID: "some-org-guid"})
				cmd.RequiredArgs = flag.LabelsArgs{
					ResourceType: "space",
					ResourceName: "fake-space",
				}
				fakeLabelsActor.GetSpaceLabelsReturns(
					map[string]types.NullString{
						"some-other-label": types.NewNullString("some-other-value"),
						"some-label":       types.NewNullString("some-value"),
					},
					v7action.Warnings{},
					nil)
			})

			It("doesn't error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
			})

			It("checks that the user is logged in and targetted to an org", func() {
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkOrg).To(BeTrue())
				Expect(checkSpace).To(BeFalse())
			})

			It("displays a message that it is retrieving the labels", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Getting labels for space fake-space in org fake-org as some-user...`)))
			})

			It("retrieves the labels associated with the space", func() {
				Expect(fakeLabelsActor.GetSpaceLabelsCallCount()).To(Equal(1))
			})

			It("displays the labels that are associated with the space, alphabetically", func() {
				Expect(testUI.Out).To(Say(`key\s+value`))
				Expect(testUI.Out).To(Say(`some-label\s+some-value`))
				Expect(testUI.Out).To(Say(`some-other-label\s+some-other-value`))
			})

			When("CAPI returns warnings", func() {
				BeforeEach(func() {
					fakeLabelsActor.GetSpaceLabelsReturns(
						map[string]types.NullString{
							"some-other-label": types.NewNullString("some-other-value"),
							"some-label":       types.NewNullString("some-value"),
						},
						v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
						nil)
				})

				It("prints all warnings", func() {
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
				})
			})

			When("there is an error retrieving the space", func() {
				BeforeEach(func() {
					fakeLabelsActor.GetSpaceLabelsReturns(
						map[string]types.NullString{},
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

			When("the resource type argument is not lowercase", func() {
				BeforeEach(func() {
					cmd.RequiredArgs = flag.LabelsArgs{
						ResourceType: "SpAcE",
						ResourceName: "fake-space",
					}
				})

				It("retrieves the labels associated with the space", func() {
					Expect(fakeLabelsActor.GetSpaceLabelsCallCount()).To(Equal(1))
					spaceName, orgGUID := fakeLabelsActor.GetSpaceLabelsArgsForCall(0)
					Expect(spaceName).To(Equal("fake-space"))
					Expect(orgGUID).To(Equal("some-org-guid"))
				})
			})

			When("when the --stack flag is specified", func() {
				verifyStackArgNotAllowed()
			})
		})

		Describe("for stacks", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserNameReturns("some-user", nil)
				cmd.RequiredArgs = flag.LabelsArgs{
					ResourceType: "stack",
					ResourceName: "fake-stack",
				}
				fakeLabelsActor.GetStackLabelsReturns(
					map[string]types.NullString{
						"some-other-label": types.NewNullString("some-other-value"),
						"some-label":       types.NewNullString("some-value"),
					},
					v7action.Warnings{},
					nil)
			})

			It("doesn't error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
			})

			It("checks that the user is logged in", func() {
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkOrg).To(BeFalse())
				Expect(checkSpace).To(BeFalse())
			})

			It("displays a message that it is retrieving the labels", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Getting labels for stack fake-stack as some-user...`)))
			})

			It("retrieves the labels associated with the stack", func() {
				Expect(fakeLabelsActor.GetStackLabelsCallCount()).To(Equal(1))
			})

			It("displays the labels that are associated with the stack, alphabetically", func() {
				Expect(testUI.Out).To(Say(`key\s+value`))
				Expect(testUI.Out).To(Say(`some-label\s+some-value`))
				Expect(testUI.Out).To(Say(`some-other-label\s+some-other-value`))
			})

			When("CAPI returns warnings", func() {
				BeforeEach(func() {
					fakeLabelsActor.GetStackLabelsReturns(
						map[string]types.NullString{
							"some-other-label": types.NewNullString("some-other-value"),
							"some-label":       types.NewNullString("some-value"),
						},
						v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
						nil)
				})

				It("prints all warnings", func() {
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
				})
			})

			When("there is an error retrieving the stack", func() {
				BeforeEach(func() {
					fakeLabelsActor.GetStackLabelsReturns(
						map[string]types.NullString{},
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

			When("checking targeted stack and space fails", func() {
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

		Describe("for buildpacks", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserNameReturns("some-user", nil)
				cmd.RequiredArgs = flag.LabelsArgs{
					ResourceType: "buildpack",
					ResourceName: "my-buildpack",
				}
				fakeLabelsActor.GetBuildpackLabelsReturns(
					map[string]types.NullString{
						"some-other-label": types.NewNullString("some-other-value"),
						"some-label":       types.NewNullString("some-value"),
					},
					v7action.Warnings{},
					nil)
			})

			It("doesn't error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
			})

			It("checks that the user is logged in", func() {
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkOrg).To(BeFalse())
				Expect(checkSpace).To(BeFalse())
			})

			Describe("the getting-labels message", func() {
				When("the buildpack stack is not specified", func() {
					BeforeEach(func() {
						cmd.BuildpackStack = ""
					})

					It("displays a message that it is retrieving the labels", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Getting labels for buildpack my-buildpack as some-user...`)))
					})
				})

				When("the buildpack stack is specified", func() {
					BeforeEach(func() {
						cmd.BuildpackStack = "omelette"
					})

					It("displays a message that it is retrieving the labels", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Getting labels for buildpack my-buildpack with stack omelette as some-user...`)))
					})
				})
			})

			It("retrieves the labels associated with the buildpack", func() {
				Expect(fakeLabelsActor.GetBuildpackLabelsCallCount()).To(Equal(1))
			})

			It("displays the labels that are associated with the buildpack, alphabetically", func() {
				Expect(testUI.Out).To(Say(`key\s+value`))
				Expect(testUI.Out).To(Say(`some-label\s+some-value`))
				Expect(testUI.Out).To(Say(`some-other-label\s+some-other-value`))
			})

			When("CAPI returns warnings", func() {
				BeforeEach(func() {
					fakeLabelsActor.GetBuildpackLabelsReturns(
						map[string]types.NullString{
							"some-other-label": types.NewNullString("some-other-value"),
							"some-label":       types.NewNullString("some-value"),
						},
						v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
						nil)
				})

				It("prints all warnings", func() {
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
				})
			})

			When("there is an error retrieving the buildpack", func() {
				BeforeEach(func() {
					fakeLabelsActor.GetBuildpackLabelsReturns(
						map[string]types.NullString{},
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

	Describe("mixed case resource names", func() {
		When("showing labels", func() {
			It("succeeds", func() {
				names := []string{"ApP", "BuIlDpAcK", "sPaCe", "StAcK", "oRg"}
				for _, name := range names {
					resourceName := "oshkosh"
					cmd.RequiredArgs = flag.LabelsArgs{
						ResourceType: name,
						ResourceName: resourceName,
					}
					executeErr := cmd.Execute(nil)
					Expect(executeErr).ToNot(HaveOccurred())
				}
			})
		})
	})

	Describe("disallowed --stack option", func() {
		When("specifying --stack", func() {
			It("complains", func() {
				names := []string{"app", "space", "stack", "org"}
				for _, name := range names {
					cmd.RequiredArgs = flag.LabelsArgs{
						ResourceType: name,
						ResourceName: "oshkosh",
					}
					cmd.BuildpackStack = "cflinuxfs3"
					executeErr := cmd.Execute(nil)
					argumentCombinationError := translatableerror.ArgumentCombinationError{
						Args: []string{strings.ToLower(cmd.RequiredArgs.ResourceType), "--stack, -s"},
					}
					Expect(executeErr).To(MatchError(argumentCombinationError), "'%s' resource did complain", name)
				}
			})

			It("retrieves the labels associated with the buildpack", func() {
				cmd.RequiredArgs = flag.LabelsArgs{
					ResourceType: "buildpack",
					ResourceName: "oshkosh",
				}
				cmd.BuildpackStack = "cflinuxfs3"
				executeErr := cmd.Execute(nil)
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeLabelsActor.GetBuildpackLabelsCallCount()).To(Equal(1))
				buildpackName, stackName := fakeLabelsActor.GetBuildpackLabelsArgsForCall(0)
				Expect(buildpackName).To(Equal("oshkosh"))
				Expect(stackName).To(Equal("cflinuxfs3"))
			})
		})
	})
})
