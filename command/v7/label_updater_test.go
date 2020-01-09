package v7_test

import (
	"errors"
	"regexp"
	"strings"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
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

var _ = Describe("LabelUpdater", func() {

	var (
		cmd             LabelUpdater
		fakeActor       *v7fakes.FakeSetLabelActor
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		testUI          *ui.UI
		targetResource  TargetResource
		labels          map[string]types.NullString
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeSetLabelActor)
		cmd = LabelUpdater{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}
	})

	Context("shared validations", func() {
		var resourceName string

		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeActor = new(v7fakes.FakeSetLabelActor)
			fakeConfig = new(commandfakes.FakeConfig)
			fakeSharedActor = new(commandfakes.FakeSharedActor)
			cmd = LabelUpdater{
				Actor:       fakeActor,
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
			}
		})

		When("fetching the current user's name fails", func() {
			BeforeEach(func() {
				targetResource = TargetResource{
					ResourceType: "anything",
					ResourceName: resourceName,
				}
				fakeConfig.CurrentUserNameReturns("some-user", errors.New("boom"))
			})

			It("returns an error", func() {
				err := cmd.Execute(targetResource, labels)
				Expect(err).To(MatchError("boom"))
			})
		})

		When("an unrecognized resource is specified", func() {
			BeforeEach(func() {
				resourceName = "some-unrecognized-resource"
				cmd = LabelUpdater{
					Actor:       fakeActor,
					UI:          testUI,
					Config:      fakeConfig,
					SharedActor: fakeSharedActor,
				}
				targetResource = TargetResource{
					ResourceType: "unrecognized-resource",
					ResourceName: resourceName,
				}
			})

			It("errors", func() {
				executeErr := cmd.Execute(targetResource, labels)

				Expect(executeErr).To(MatchError("Unsupported resource type of 'unrecognized-resource'"))
			})
		})

		DescribeTable(
			"Combination of --stack with resource type",
			func(resourceType string) {
				targetResource = TargetResource{
					ResourceType:   resourceType,
					BuildpackStack: "cflinuxfs3",
				}

				err := cmd.Execute(targetResource, nil)

				argumentCombinationError := translatableerror.ArgumentCombinationError{
					Args: []string{strings.ToLower(resourceType), "--stack, -s"},
				}
				Expect(err).To(MatchError(argumentCombinationError))
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
				targetResource = TargetResource{
					ResourceType: resourceType,
				}

				err := cmd.Execute(targetResource, nil)
				Expect(err).To(MatchError("Target not found"))
			},
			Entry("app", "app"),
			Entry("buildpack", "buildpack"),
			// domain - does not check target
			Entry("org", "org"),
			Entry("route", "route"),
			Entry("service-broker", "service-broker"),
			Entry("space", "space"),
			Entry("stack", "stack"),
		)
	})

	When("updating labels on apps", func() {
		var (
			appName     string
			executeErr  error
			expectedMap map[string]types.NullString
		)

		BeforeEach(func() {
			appName = "some-app"
			targetResource = TargetResource{
				ResourceType: "app",
				ResourceName: appName,
			}

			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org"})
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "fake-space", GUID: "some-space-guid"})
			fakeConfig.CurrentUserNameReturns("some-user", nil)

			expectedMap = map[string]types.NullString{
				"some-label":     types.NewNullString("some-value"),
				"some-other-key": types.NewNullString()}
			labels = expectedMap
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
		})

		It("checks that the user is logged in and targeted to an org and space", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkOrg).To(BeTrue())
			Expect(checkSpace).To(BeTrue())
		})

		When("updating the app labels succeeds", func() {
			BeforeEach(func() {
				fakeActor.UpdateApplicationLabelsByApplicationNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
					nil)

			})

			It("prints all warnings and does not return an error ", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

			It("passes the correct parameters into the actor", func() {
				Expect(fakeActor.UpdateApplicationLabelsByApplicationNameCallCount()).To(Equal(1))
				actualAppName, spaceGUID, labelsMap := fakeActor.UpdateApplicationLabelsByApplicationNameArgsForCall(0)
				Expect(actualAppName).To(Equal(appName))
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(labelsMap).To(Equal(expectedMap))
			})
		})
		// FIXME maybe checking it calls the right method is enough?

		When("the resource type argument is not lowercase", func() {
			BeforeEach(func() {
				targetResource.ResourceType = "aPp"
			})

			It("passes the correct parameters into the actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeActor.UpdateApplicationLabelsByApplicationNameCallCount()).To(Equal(1))
				actualAppName, spaceGUID, labelsMap := fakeActor.UpdateApplicationLabelsByApplicationNameArgsForCall(0)
				Expect(actualAppName).To(Equal(appName))
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(labelsMap).To(Equal(expectedMap))
			})

			It("displays a message in the right casing", func() {
				Expect(testUI.Out).To(Say("(.*) label\\(s\\) for app (.*)"))
				Expect(testUI.Out).To(Say("OK"))
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

		Context("Shows the right update message with org and space", func() {
			When("Unsetting labels", func() {
				BeforeEach(func() {
					cmd.Action = Unset
					//FIXME do we want to change the labels to all have nil values?
				})
				It("shows 'Removing' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for app %s in org fake-org / space fake-space as some-user...`), appName))
				})
			})
			When("Setting labels", func() {
				BeforeEach(func() {
					cmd.Action = Set
					//FIXME do we want to change the labels to all have not nil values?
				})

				It("shows 'Setting' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for app %s in org fake-org / space fake-space as some-user...`), appName))

				})
			})

		})
	})

	When("updating labels on buildpacks", func() {
		var resourceName string
		var expectedMap map[string]types.NullString
		var executeErr error

		BeforeEach(func() {
			resourceName = "buildpack-name"
			targetResource = TargetResource{
				ResourceType: "buildpack",
				ResourceName: resourceName,
			}

			fakeSharedActor.CheckTargetReturns(nil)
			fakeConfig.CurrentUserNameReturns("some-user", nil)

			expectedMap = map[string]types.NullString{
				"some-label":     types.NewNullString("some-value"),
				"some-other-key": types.NewNullString()}
			labels = expectedMap

			fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackReturns(
				v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
				nil,
			)

		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
		})

		It("checks that the user is logged in", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkOrg).To(BeFalse())
			Expect(checkSpace).To(BeFalse())
		})

		When("updating the buildpack labels succeeds", func() {
			When("the stack is specified", func() {
				BeforeEach(func() {
					targetResource.BuildpackStack = "globinski"
				})

				It("passes the right parameters", func() {
					Expect(fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackCallCount()).To(Equal(1))
					name, stack, labels := fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackArgsForCall(0)
					Expect(name).To(Equal(resourceName), "failed to pass buildpack name")
					Expect(stack).To(Equal("globinski"), "failed to pass stack name")
					Expect(labels).To(BeEquivalentTo(expectedMap))
				})

				It("prints all warnings and does not return an argument combination error ", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
				})
			})

			When("the stack is not specified", func() {
				It("passes the right parameters", func() {
					Expect(fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackCallCount()).To(Equal(1))
					name, stack, labels := fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackArgsForCall(0)
					Expect(name).To(Equal(resourceName), "failed to pass buildpack name")
					Expect(stack).To(Equal(""), "failed to pass stack name")
					Expect(labels).To(BeEquivalentTo(expectedMap))
				})
			})
		})

		When("the resource type is not lowercase", func() {
			BeforeEach(func() {
				targetResource = TargetResource{
					ResourceType: "bUiLdPaCk",
					ResourceName: resourceName,
				}
				expectedMap = map[string]types.NullString{
					"some-label":     types.NewNullString("some-value"),
					"some-other-key": types.NewNullString()}
			})

			When("updating the buildpack labels succeeds", func() {
				When("the stack is specified", func() {
					BeforeEach(func() {
						targetResource.BuildpackStack = "globinski"
					})

					It("passes the right parameters", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackCallCount()).To(Equal(1))
						name, stack, labels := fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackArgsForCall(0)
						Expect(name).To(Equal(resourceName), "failed to pass buildpack name")
						Expect(stack).To(Equal("globinski"), "failed to pass stack name")
						Expect(labels).To(BeEquivalentTo(expectedMap))
					})

					It("displays a message in the right casing", func() {
						Expect(testUI.Out).To(Say("(.*) label\\(s\\) for buildpack (.*)"))
						Expect(testUI.Out).To(Say("OK"))
					})

				})

			})
		})

		When("updating the buildpack labels fails", func() {
			BeforeEach(func() {
				fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("api call failed"))
			})

			It("returns the error and prints all warnings", func() {
				Expect(executeErr).To(MatchError("api call failed"))
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

		})

		Context("Shows the right update message with correct stack and action", func() {
			When("Unsetting labels", func() {
				BeforeEach(func() {
					cmd.Action = Unset
					//FIXME do we want to change the labels to all have nil values?
				})

				When("stack is passed", func() {
					BeforeEach(func() {
						targetResource.BuildpackStack = "globinski"
					})
					It("shows 'Removing' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for buildpack %s with stack %s as some-user...`), resourceName, targetResource.BuildpackStack))
					})
				})

				When("stack is not passed", func() {
					It("shows 'Removing' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for buildpack %s as some-user...`), resourceName))
					})
				})

			})
			When("Setting labels", func() {
				BeforeEach(func() {
					cmd.Action = Set
					//FIXME do we want to change the labels to all have not nil values?
				})

				When("stack is passed", func() {
					BeforeEach(func() {
						targetResource.BuildpackStack = "globinski"
					})
					It("shows 'Setting' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for buildpack %s with stack %s as some-user...`), resourceName, targetResource.BuildpackStack))
					})
				})

				When("stack is not passed", func() {
					It("shows 'Setting' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for buildpack %s as some-user...`), resourceName))
					})
				})
			})

		})
	})

	When("updating labels in domains", func() {
		var (
			domainName  string
			executeErr  error
			expectedMap map[string]types.NullString
		)

		BeforeEach(func() {
			domainName = "example.com"
			targetResource = TargetResource{
				ResourceType: "domain",
				ResourceName: domainName,
			}
			expectedMap = map[string]types.NullString{
				"some-label":     types.NewNullString("some-value"),
				"some-other-key": types.NewNullString()}
			labels = expectedMap

			fakeConfig.CurrentUserNameReturns("some-user", nil)
			fakeActor.UpdateDomainLabelsByDomainNameReturns(
				v7action.Warnings{"some-warning-1", "some-warning-2"},
				nil,
			)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
		})

		It("doesn't check that the user is logged in", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(0))
		})

		When("updating the app labels succeeds", func() {
			It("prints all warnings and does not return an error ", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

			It("passes the correct parameters into the actor", func() {
				Expect(fakeActor.UpdateDomainLabelsByDomainNameCallCount()).To(Equal(1))
				name, labels := fakeActor.UpdateDomainLabelsByDomainNameArgsForCall(0)
				Expect(name).To(Equal(domainName), "failed to pass domain name")
				Expect(labels).To(BeEquivalentTo(expectedMap))
			})
		})

		When("the resource type argument is not lowercase", func() {
			BeforeEach(func() {
				targetResource.ResourceType = "DoMaiN"
			})

			It("passes the correct parameters into the actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeActor.UpdateDomainLabelsByDomainNameCallCount()).To(Equal(1))
				name, labels := fakeActor.UpdateDomainLabelsByDomainNameArgsForCall(0)
				Expect(name).To(Equal(domainName), "failed to pass domain name")
				Expect(labels).To(BeEquivalentTo(expectedMap))
			})

			It("displays a message in the right casing", func() {
				Expect(testUI.Out).To(Say("(.*) label\\(s\\) for domain (.*)"))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("updating the domain labels fails", func() {
			BeforeEach(func() {
				fakeActor.UpdateDomainLabelsByDomainNameReturns(
					v7action.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("api call failed"))
			})

			It("returns the error and prints all warnings", func() {
				Expect(executeErr).To(MatchError("api call failed"))
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})
		})

		Context("Shows the right update message", func() {
			When("Unsetting labels", func() {
				BeforeEach(func() {
					cmd.Action = Unset
				})
				It("shows 'Removing' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for domain %s as some-user...`), domainName))
				})
			})
			When("Setting labels", func() {
				BeforeEach(func() {
					cmd.Action = Set
				})

				It("shows 'Setting' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for domain %s as some-user...`), domainName))
				})
			})
		})
	})

	When("updating labels on orgs", func() {
		var (
			executeErr  error
			orgName     = "some-org"
			expectedMap map[string]types.NullString
		)

		BeforeEach(func() {
			targetResource = TargetResource{
				ResourceType: "org",
				ResourceName: orgName,
			}
			fakeSharedActor.CheckTargetReturns(nil)
			fakeConfig.CurrentUserNameReturns("some-user", nil)

			expectedMap = map[string]types.NullString{
				"some-label":     types.NewNullString("some-value"),
				"some-other-key": types.NewNullString()}
			labels = expectedMap

			fakeActor.UpdateOrganizationLabelsByOrganizationNameReturns(
				v7action.Warnings{"some-warning-1", "some-warning-2"},
				nil)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
		})

		It("checks that the user is logged in", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkOrg).To(BeFalse())
			Expect(checkSpace).To(BeFalse())
		})

		When("updating the orgs labels succeeds", func() {
			It("does not return an error and prints all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

			It("passes the correct parameters into the actor", func() {
				Expect(fakeActor.UpdateOrganizationLabelsByOrganizationNameCallCount()).To(Equal(1))
				actualOrgName, labelsMap := fakeActor.UpdateOrganizationLabelsByOrganizationNameArgsForCall(0)
				Expect(actualOrgName).To(Equal(orgName))
				Expect(labelsMap).To(Equal(expectedMap))
			})
		})

		When("the resource type argument is not lowercase", func() {
			BeforeEach(func() {
				targetResource.ResourceType = "OrG"
			})

			It("passes the correct parameters into the actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.UpdateOrganizationLabelsByOrganizationNameCallCount()).To(Equal(1))
				actualOrgName, labelsMap := fakeActor.UpdateOrganizationLabelsByOrganizationNameArgsForCall(0)
				Expect(actualOrgName).To(Equal(orgName))
				Expect(labelsMap).To(Equal(expectedMap))
			})

			It("displays a message in the right casing", func() {
				Expect(testUI.Out).To(Say("(.*) label\\(s\\) for org (.*)"))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("updating the org labels fails", func() {
			BeforeEach(func() {
				fakeActor.UpdateOrganizationLabelsByOrganizationNameReturns(
					v7action.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("api call failed"))
			})

			It("returns the error and prints all warnings", func() {
				Expect(executeErr).To(MatchError("api call failed"))
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

		})

		Context("Shows the right update message", func() {
			When("Unsetting labels", func() {
				BeforeEach(func() {
					cmd.Action = Unset
				})
				It("shows 'Removing' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for org %s as some-user...`), orgName))
				})
			})
			When("Setting labels", func() {
				BeforeEach(func() {
					cmd.Action = Set
				})

				It("shows 'Setting' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for org %s as some-user...`), orgName))
				})
			})
		})
	})

	When("updating labels on routes", func() {
		var (
			resourceName string
			expectedMap  map[string]types.NullString
			executeErr   error
		)

		BeforeEach(func() {
			resourceName = "some-route.example.com"
			targetResource = TargetResource{
				ResourceType: "route",
				ResourceName: resourceName,
			}

			expectedMap = map[string]types.NullString{
				"some-label":     types.NewNullString("some-value"),
				"some-other-key": types.NewNullString()}
			labels = expectedMap

			fakeConfig.CurrentUserNameReturns("some-user", nil)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org"})
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "fake-space", GUID: "space-guid"})
			fakeActor.UpdateRouteLabelsReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
				nil)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
		})

		It("checks that the user is logged in", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkOrg).To(BeTrue())
			Expect(checkSpace).To(BeTrue())
		})

		When("updating the route labels succeeds", func() {
			It("doesn't error and prints all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

			It("passes the right parameters to the actor", func() {
				Expect(fakeActor.UpdateRouteLabelsCallCount()).To(Equal(1))
				name, spaceGUID, labels := fakeActor.UpdateRouteLabelsArgsForCall(0)
				Expect(name).To(Equal(resourceName), "failed to pass route name")
				Expect(labels).To(BeEquivalentTo(expectedMap))
				Expect(spaceGUID).To(Equal("space-guid"))
			})
		})

		When("the resource type argument is not lowercase", func() {
			BeforeEach(func() {
				targetResource.ResourceType = "rouTE"
			})

			It("passes the correct parameters into the actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeActor.UpdateRouteLabelsCallCount()).To(Equal(1))

				name, spaceGUID, labels := fakeActor.UpdateRouteLabelsArgsForCall(0)

				Expect(name).To(Equal(resourceName), "failed to pass route name")
				Expect(spaceGUID).To(Equal("space-guid"))
				Expect(labels).To(BeEquivalentTo(expectedMap))
			})

			It("displays a message in the right casing", func() {
				Expect(testUI.Out).To(Say("(.*) label\\(s\\) for route (.*)"))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("updating the route labels fails", func() {
			BeforeEach(func() {
				fakeActor.UpdateRouteLabelsReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("api call failed"))
			})

			It("returns the error and prints all warnings", func() {
				Expect(executeErr).To(MatchError("api call failed"))
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

		})

		Context("Shows the right update message", func() {
			When("Unsetting labels", func() {
				BeforeEach(func() {
					cmd.Action = Unset
				})
				It("shows 'Removing' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for route %s in org fake-org / space fake-space as some-user...`), resourceName))
				})
			})
			When("Setting labels", func() {
				BeforeEach(func() {
					cmd.Action = Set
				})

				It("shows 'Setting' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for route %s in org fake-org / space fake-space as some-user...`), resourceName))
				})
			})
		})
	})

	When("updating labels on service-broker", func() {
		expectedServiceBrokerName := "my-broker"
		var (
			expectedMap map[string]types.NullString
			executeErr  error
		)

		BeforeEach(func() {
			targetResource = TargetResource{
				ResourceType: "service-broker",
				ResourceName: expectedServiceBrokerName,
			}

			fakeConfig.CurrentUserNameReturns("some-user", nil)
			expectedMap = map[string]types.NullString{
				"some-label":     types.NewNullString("some-value"),
				"some-other-key": types.NewNullString(),
			}
			labels = expectedMap
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
		})

		It("checks that the user is logged in", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkOrg).To(BeFalse())
			Expect(checkSpace).To(BeFalse())
		})

		When("updating the service-broker labels succeeds", func() {
			BeforeEach(func() {
				fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
					nil)
			})

			It("prints all warnings and does not return an error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

			It("passes the correct parameters into the actor", func() {
				Expect(fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameCallCount()).To(Equal(1))
				serviceBrokerName, labelsMap := fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameArgsForCall(0)
				Expect(serviceBrokerName).To(Equal(expectedServiceBrokerName))
				Expect(labelsMap).To(Equal(expectedMap))
			})
		})

		When("the resource type argument is not lowercase", func() {
			BeforeEach(func() {
				fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
					nil)

				targetResource.ResourceType = "sErVice-BroKer"
			})

			It("passes the correct parameters into the actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameCallCount()).To(Equal(1))
				serviceBrokerName, labelsMap := fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameArgsForCall(0)
				Expect(serviceBrokerName).To(Equal(expectedServiceBrokerName))
				Expect(labelsMap).To(Equal(expectedMap))
			})

			It("displays a message in the right casing", func() {
				Expect(testUI.Out).To(Say("(.*) label\\(s\\) for service-broker (.*)"))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("updating the service-broker labels fails", func() {
			BeforeEach(func() {
				fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("api call failed"))
			})

			It("prints all warnings and returns the error", func() {
				Expect(executeErr).To(MatchError("api call failed"))
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})
		})

		Context("Shows the right update message", func() {
			When("Unsetting labels", func() {
				BeforeEach(func() {
					cmd.Action = Unset
				})
				It("shows 'Removing' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for service-broker %s as some-user...`), expectedServiceBrokerName))
					Expect(testUI.Out).To(Say("OK"))
				})
			})
			When("Setting labels", func() {
				BeforeEach(func() {
					cmd.Action = Set
				})

				It("shows 'Setting' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for service-broker %s as some-user...`), expectedServiceBrokerName))
					Expect(testUI.Out).To(Say("OK"))
				})
			})
		})
	})

	When("updating labels on spaces", func() {
		var (
			executeErr  error
			spaceName   string
			expectedMap map[string]types.NullString
		)

		BeforeEach(func() {
			spaceName = "spiff"
			targetResource = TargetResource{
				ResourceType: "space",
				ResourceName: spaceName,
			}
			expectedMap = map[string]types.NullString{
				"some-label":     types.NewNullString("some-value"),
				"some-other-key": types.NewNullString()}
			labels = expectedMap

			fakeConfig.CurrentUserNameReturns("some-user", nil)
			fakeConfig.TargetedOrganizationReturns(
				configv3.Organization{Name: "fake-org", GUID: "some-org-guid"})

			fakeActor.UpdateSpaceLabelsBySpaceNameReturns(
				v7action.Warnings{"some-warning-1", "some-warning-2"},
				nil)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
		})

		It("checks that the user is logged in and targeted to an org and space", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkOrg).To(BeTrue())
			Expect(checkSpace).To(BeFalse())
		})

		When("updating the space labels succeeds", func() {
			It("does not return an error and prints all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

			It("passes the correct parameters into the actor", func() {
				Expect(fakeActor.UpdateSpaceLabelsBySpaceNameCallCount()).To(Equal(1))
				actualSpaceName, orgGUID, labelsMap := fakeActor.UpdateSpaceLabelsBySpaceNameArgsForCall(0)
				Expect(actualSpaceName).To(Equal(spaceName))
				Expect(orgGUID).To(Equal("some-org-guid"))
				Expect(labelsMap).To(Equal(expectedMap))
			})

		})

		When("the resource type argument is not lowercase", func() {
			BeforeEach(func() {
				targetResource.ResourceType = "SpAcE"
			})

			It("passes the right parameters to the actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.UpdateSpaceLabelsBySpaceNameCallCount()).To(Equal(1))
				actualSpaceName, orgGUID, labelsMap := fakeActor.UpdateSpaceLabelsBySpaceNameArgsForCall(0)
				Expect(actualSpaceName).To(Equal(spaceName))
				Expect(orgGUID).To(Equal("some-org-guid"))
				Expect(labelsMap).To(Equal(expectedMap))
			})

			It("displays a message in the right casing", func() {
				Expect(testUI.Out).To(Say("(.*) label\\(s\\) for space (.*)"))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("updating the space labels fails", func() {
			BeforeEach(func() {
				fakeActor.UpdateSpaceLabelsBySpaceNameReturns(
					v7action.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("api call failed"),
				)
			})

			It("returns the error and prints all warnings", func() {
				Expect(executeErr).To(MatchError("api call failed"))
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})
		})

		Context("Shows the right update message", func() {
			When("Unsetting labels", func() {
				BeforeEach(func() {
					cmd.Action = Unset
				})
				It("shows 'Removing' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for space %s in org fake-org as some-user...`), spaceName))
					Expect(testUI.Out).To(Say("OK"))
				})
			})

			When("Setting labels", func() {
				BeforeEach(func() {
					cmd.Action = Set
				})

				It("shows 'Setting' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for space %s in org fake-org as some-user...`), spaceName))
					Expect(testUI.Out).To(Say("OK"))
				})
			})
		})
	})

	When("updating labels on stacks", func() {
		var (
			executeErr  error
			expectedMap map[string]types.NullString
		)

		stackName := "some-stack"
		BeforeEach(func() {
			targetResource = TargetResource{
				ResourceType: "stack",
				ResourceName: stackName,
			}
			fakeConfig.CurrentUserNameReturns("some-user", nil)
			fakeSharedActor.CheckTargetReturns(nil)
			expectedMap = map[string]types.NullString{
				"some-label":     types.NewNullString("some-value"),
				"some-other-key": types.NewNullString(),
			}
			labels = expectedMap

		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
		})

		It("checks that the user is logged in but not necessarily targeted to an org", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkOrg).To(BeFalse())
			Expect(checkSpace).To(BeFalse())
		})

		When("updating the stack labels succeeds", func() {
			BeforeEach(func() {
				fakeActor.UpdateStackLabelsByStackNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
					nil)
			})

			It("does not return an error and prints all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))

			})

			It("passes the correct parameters into the actor", func() {

				Expect(fakeActor.UpdateStackLabelsByStackNameCallCount()).To(Equal(1))
				actualStackName, labelsMap := fakeActor.UpdateStackLabelsByStackNameArgsForCall(0)
				Expect(actualStackName).To(Equal(stackName))
				Expect(labelsMap).To(Equal(expectedMap))
			})

		})

		When("the resource type argument is not lowercase", func() {
			BeforeEach(func() {
				targetResource.ResourceType = "sTaCk"
			})
			It("passes the correct parameters into the actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.UpdateStackLabelsByStackNameCallCount()).To(Equal(1))
				actualStackName, labelsMap := fakeActor.UpdateStackLabelsByStackNameArgsForCall(0)
				Expect(actualStackName).To(Equal(stackName))
				Expect(labelsMap).To(Equal(expectedMap))
			})

			It("displays a message in the right casing", func() {
				Expect(testUI.Out).To(Say("(.*) label\\(s\\) for stack (.*)"))
				Expect(testUI.Out).To(Say("OK"))
			})

		})

		When("updating the stack labels fails", func() {
			BeforeEach(func() {
				fakeActor.UpdateStackLabelsByStackNameReturns(
					v7action.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("api call failed"),
				)
			})

			It("returns the error and prints all warnings", func() {
				Expect(executeErr).To(MatchError("api call failed"))
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})
		})

		Context("Shows the right update message", func() {
			When("Unsetting labels", func() {
				BeforeEach(func() {
					cmd.Action = Unset
				})
				It("shows 'Removing' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for stack %s as some-user...`), stackName))
					Expect(testUI.Out).To(Say("OK"))
				})
			})

			When("Setting labels", func() {
				BeforeEach(func() {
					cmd.Action = Set
				})

				It("shows 'Setting' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for stack %s as some-user...`), stackName))
					Expect(testUI.Out).To(Say("OK"))
				})
			})
		})

	})

})
