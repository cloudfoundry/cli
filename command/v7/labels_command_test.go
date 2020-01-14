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
	. "github.com/onsi/ginkgo/extensions/table"
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

	Context("shared validations", func() {
		When("fetching the current user's name fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserNameReturns("some-user", errors.New("boom"))
				executeErr = cmd.Execute(nil)
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("boom"))
			})
		})

		When("an unrecognized resource is specified", func() {
			BeforeEach(func() {
				cmd.RequiredArgs.ResourceType = "unrecognized-resource"
				executeErr = cmd.Execute(nil)
			})

			It("errors", func() {
				Expect(executeErr).To(MatchError("Unsupported resource type of 'unrecognized-resource'"))
			})
		})

		DescribeTable(
			"Combination of --stack with resource type",
			func(resourceType string) {
				cmd.RequiredArgs.ResourceType = resourceType
				cmd.BuildpackStack = "cflinuxfs3"

				executeErr = cmd.Execute(nil)

				argumentCombinationError := translatableerror.ArgumentCombinationError{
					Args: []string{strings.ToLower(resourceType), "--stack, -s"},
				}
				Expect(executeErr).To(MatchError(argumentCombinationError))
			},
			Entry("app", "app"),
			Entry("domains", "domain"),
			Entry("orgs", "org"),
			Entry("routes", "route"),
			Entry("spaces", "space"),
			Entry("stacks", "stack"),
			Entry("service brokers", "service-broker"),
		)

		DescribeTable(
			"when checking the target fails",
			func(resourceType string) {
				fakeSharedActor.CheckTargetReturns(errors.New("Target not found"))
				cmd.RequiredArgs.ResourceType = resourceType
				err := cmd.Execute(nil)
				Expect(err).To(MatchError("Target not found"))
			},
			Entry("app", "app"),
			Entry("buildpack", "buildpack"),
			Entry("domain", "domain"),
			Entry("org", "org"),
			Entry("route", "route"),
			Entry("service-broker", "service-broker"),
			Entry("space", "space"),
			Entry("stack", "stack"),
		)

		type MethodCallCountType func() int
		When("the resource type is not lowercase", func() {
			It("calls the right method", func() {
				testBody := func(resourceType string, expectedMethodCallCount MethodCallCountType) {
					cmd.RequiredArgs.ResourceType = resourceType
					err := cmd.Execute(nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(expectedMethodCallCount()).To(Equal(1))
				}

				testCases :=
					map[string]MethodCallCountType{
						"aPp":            fakeLabelsActor.GetApplicationLabelsCallCount,
						"bUiLdPaCK":      fakeLabelsActor.GetBuildpackLabelsCallCount,
						"dOmAiN":         fakeLabelsActor.GetDomainLabelsCallCount,
						"oRg":            fakeLabelsActor.GetOrganizationLabelsCallCount,
						"rOuTe":          fakeLabelsActor.GetRouteLabelsCallCount,
						"sErViCe-BrOkEr": fakeLabelsActor.GetServiceBrokerLabelsCallCount,
						"sPaCe":          fakeLabelsActor.GetSpaceLabelsCallCount,
						"sTaCk":          fakeLabelsActor.GetStackLabelsCallCount,
					}

				for resourceType, callCountMethod := range testCases {
					testBody(resourceType, callCountMethod)
				}
			})
		})

	})

	Describe("listing labels", func() {

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

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
		})

		Describe("for domains", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserNameReturns("some-user", nil)
				cmd.RequiredArgs = flag.LabelsArgs{
					ResourceType: "domain",
					ResourceName: "example.com",
				}
				fakeLabelsActor.GetDomainLabelsReturns(
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
				Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Getting labels for domain example.com as some-user...`)))
			})

			It("retrieves the labels associated with the domain", func() {
				Expect(fakeLabelsActor.GetDomainLabelsCallCount()).To(Equal(1))
				domainName := fakeLabelsActor.GetDomainLabelsArgsForCall(0)
				Expect(domainName).To(Equal("example.com"))
			})

			It("displays the labels that are associated with the domain, alphabetically", func() {
				Expect(testUI.Out).To(Say(`key\s+value`))
				Expect(testUI.Out).To(Say(`some-label\s+some-value`))
				Expect(testUI.Out).To(Say(`some-other-label\s+some-other-value`))
			})

			When("CAPI returns warnings", func() {
				BeforeEach(func() {
					fakeLabelsActor.GetDomainLabelsReturns(
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

			When("there is an error retrieving the domain", func() {
				BeforeEach(func() {
					fakeLabelsActor.GetDomainLabelsReturns(
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

		})

		Describe("for routes", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserNameReturns("some-user", nil)
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org"})
				fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "fake-space", GUID: "some-space-guid"})
				cmd.RequiredArgs = flag.LabelsArgs{
					ResourceType: "route",
					ResourceName: "foo.example.com/the-path",
				}
				fakeLabelsActor.GetRouteLabelsReturns(
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
				Expect(checkOrg).To(BeTrue())
				Expect(checkSpace).To(BeTrue())
			})

			It("displays a message that it is retrieving the labels", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Getting labels for route foo.example.com/the-path in org fake-org / space fake-space as some-user...`)))
			})

			It("retrieves the labels associated with the route", func() {
				Expect(fakeLabelsActor.GetRouteLabelsCallCount()).To(Equal(1))
				routeName, spaceGUID := fakeLabelsActor.GetRouteLabelsArgsForCall(0)
				Expect(routeName).To(Equal("foo.example.com/the-path"))
				Expect(spaceGUID).To(Equal("some-space-guid"))
			})

			It("displays the labels that are associated with the route, alphabetically", func() {
				Expect(testUI.Out).To(Say(`key\s+value`))
				Expect(testUI.Out).To(Say(`some-label\s+some-value`))
				Expect(testUI.Out).To(Say(`some-other-label\s+some-other-value`))
			})

			When("CAPI returns warnings", func() {
				BeforeEach(func() {
					fakeLabelsActor.GetRouteLabelsReturns(
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

			When("there is an error retrieving the route", func() {
				BeforeEach(func() {
					fakeLabelsActor.GetRouteLabelsReturns(
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

			When("specifying --stack", func() {
				BeforeEach(func() {
					cmd.RequiredArgs = flag.LabelsArgs{
						ResourceType: "buildpack",
						ResourceName: "oshkosh",
					}
					cmd.BuildpackStack = "cflinuxfs3"
				})
				It("retrieves the labels when resource type is buildpack", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeLabelsActor.GetBuildpackLabelsCallCount()).To(Equal(1))
					buildpackName, stackName := fakeLabelsActor.GetBuildpackLabelsArgsForCall(0)
					Expect(buildpackName).To(Equal("oshkosh"))
					Expect(stackName).To(Equal("cflinuxfs3"))
				})
			})
		})

		Describe("for service-brokers", func() {
			When("There is an error fetching the labels", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("some-user", nil)

					cmd.RequiredArgs = flag.LabelsArgs{
						ResourceType: "service-broker",
						ResourceName: "existent-broker",
					}

					fakeLabelsActor.GetServiceBrokerLabelsReturns(
						map[string]types.NullString{},
						v7action.Warnings([]string{"a warning"}),
						errors.New("some random error"))
				})

				It("returns an error and prints all warnings", func() {
					Expect(executeErr).To(MatchError("some random error"))
					Expect(testUI.Err).To(Say("a warning"))
				})

				It("displays a message that it is retrieving the labels", func() {
					Expect(testUI.Out).To(Say("Getting labels for service-broker existent-broker as some-user..."))
				})
			})

			When("Service broker has labels", func() {
				var labels map[string]types.NullString
				BeforeEach(func() {
					labels = map[string]types.NullString{
						"some-other-label": types.NewNullString("some-other-value"),
						"some-label":       types.NewNullString("some-value"),
					}

					fakeConfig.CurrentUserNameReturns("some-user", nil)

					cmd.RequiredArgs = flag.LabelsArgs{
						ResourceType: "service-broker",
						ResourceName: "a-broker",
					}

					fakeLabelsActor.GetServiceBrokerLabelsReturns(
						labels,
						v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
						nil,
					)
				})

				It("displays a message that it is retrieving the labels", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Getting labels for service-broker a-broker as some-user...`)))
				})

				It("retrieves the labels associated with the broker, alphabetically", func() {
					Expect(testUI.Out).To(Say(`key\s+value`))
					Expect(testUI.Out).To(Say(`some-label\s+some-value`))
					Expect(testUI.Out).To(Say(`some-other-label\s+some-other-value`))
				})

				It("prints all the warnings", func() {
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))

				})

				It("checks that the user is logged in", func() {
					Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
					checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
					Expect(checkOrg).To(BeFalse())
					Expect(checkSpace).To(BeFalse())
				})

			})
		})
	})
})
