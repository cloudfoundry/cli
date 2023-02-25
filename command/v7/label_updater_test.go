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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

func labelSubcommands(subcommandsToRemove ...string) []TableEntry {
	all := []string{
		"app",
		"buildpack",
		"domain",
		"org",
		"route",
		"space",
		"stack",
		"service-broker",
		"service-instance",
		"service-offering",
		"service-plan",
	}
	var entries []TableEntry
	for _, labelSubcommand := range all {
		remove := false
		for _, subCommand := range subcommandsToRemove {
			if labelSubcommand == subCommand {
				remove = true
				break
			}
		}
		if !remove {
			entries = append(entries, Entry(labelSubcommand, labelSubcommand))
		}
	}
	return entries
}

var _ = Describe("LabelUpdater", func() {
	var (
		cmd             LabelUpdater
		fakeActor       *v7fakes.FakeActor
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
		fakeActor = new(v7fakes.FakeActor)
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
			fakeActor = new(v7fakes.FakeActor)
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
				fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, errors.New("boom"))
			})

			It("returns an error", func() {
				err := cmd.Execute(targetResource, labels)
				Expect(err).To(MatchError("boom"))
			})
		})

		When("an unrecognized resource type is specified", func() {
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
			"Failure when --stack is combined with anything other than 'buildpack'",
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
			labelSubcommands("buildpack")...,
		)

		DescribeTable(
			"Failure when --broker is combined with anything other than 'service-offering' or 'service-plan'",
			func(resourceType string) {
				targetResource = TargetResource{
					ResourceType:  resourceType,
					ServiceBroker: "my-broker",
				}

				err := cmd.Execute(targetResource, nil)

				argumentCombinationError := translatableerror.ArgumentCombinationError{
					Args: []string{strings.ToLower(resourceType), "--broker, -b"},
				}
				Expect(err).To(MatchError(argumentCombinationError))
			},
			labelSubcommands("service-offering", "service-plan")...,
		)

		DescribeTable(
			"Failure when --offering is combined with anything other than 'service-plan'",
			func(resourceType string) {
				targetResource = TargetResource{
					ResourceType:    resourceType,
					ServiceOffering: "my-service-offering",
				}

				err := cmd.Execute(targetResource, nil)

				argumentCombinationError := translatableerror.ArgumentCombinationError{
					Args: []string{strings.ToLower(resourceType), "--offering, -o"},
				}
				Expect(err).To(MatchError(argumentCombinationError))
			},
			labelSubcommands("service-plan")...,
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
			labelSubcommands()...,
		)

		DescribeTable(
			"checking that the user is logged in",
			func(resourceType string) {
				targetResource = TargetResource{
					ResourceType: resourceType,
				}
				err := cmd.Execute(targetResource, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)

				switch resourceType {
				case "app", "route", "service-instance":
					Expect(checkOrg).To(BeTrue())
					Expect(checkSpace).To(BeTrue())
				case "space":
					Expect(checkOrg).To(BeTrue())
					Expect(checkSpace).To(BeFalse())
				default:
					Expect(checkOrg).To(BeFalse())
					Expect(checkSpace).To(BeFalse())
				}
			},
			labelSubcommands()...,
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
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)

			expectedMap = map[string]types.NullString{
				"some-label":     types.NewNullString("some-value"),
				"some-other-key": types.NewNullString(),
			}
			labels = expectedMap
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
		})

		When("updating the app labels succeeds", func() {
			BeforeEach(func() {
				fakeActor.UpdateApplicationLabelsByApplicationNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"}, nil)
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

		When("the resource type argument is not lowercase", func() {
			BeforeEach(func() {
				targetResource.ResourceType = "aPp"
			})

			It("calls the right actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.UpdateApplicationLabelsByApplicationNameCallCount()).To(Equal(1))
			})

			It("displays a message in the right case", func() {
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

		Context("shows the right update message with org and space", func() {
			When("Unsetting labels", func() {
				BeforeEach(func() {
					cmd.Action = Unset
					// FIXME do we want to change the labels to all have nil values?
				})
				It("shows 'Removing' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for app %s in org fake-org / space fake-space as some-user...`), appName))
				})
			})

			When("Setting labels", func() {
				BeforeEach(func() {
					cmd.Action = Set
					// FIXME do we want to change the labels to all have not nil values?
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
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)

			expectedMap = map[string]types.NullString{
				"some-label":     types.NewNullString("some-value"),
				"some-other-key": types.NewNullString(),
			}
			labels = expectedMap

			fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackReturns(
				v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
				nil,
			)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
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

				It("prints all warnings and does not error ", func() {
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

				It("prints all warnings and does not error ", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
				})
			})
		})

		When("the resource type is not lowercase", func() {
			BeforeEach(func() {
				targetResource = TargetResource{
					ResourceType:   "bUiLdPaCk",
					ResourceName:   resourceName,
					BuildpackStack: "globinski",
				}
				expectedMap = map[string]types.NullString{
					"some-label":     types.NewNullString("some-value"),
					"some-other-key": types.NewNullString(),
				}
			})

			It("calls the right actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackCallCount()).To(Equal(1))
			})

			It("displays a message in the right case", func() {
				Expect(testUI.Out).To(Say("(.*) label\\(s\\) for buildpack (.*)"))
				Expect(testUI.Out).To(Say("OK"))
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

		Context("shows the right update message with correct stack and action", func() {
			When("Unsetting labels", func() {
				BeforeEach(func() {
					cmd.Action = Unset
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
					// FIXME do we want to change the labels to all have not nil values?
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
				"some-other-key": types.NewNullString(),
			}
			labels = expectedMap

			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			fakeActor.UpdateDomainLabelsByDomainNameReturns(
				v7action.Warnings{"some-warning-1", "some-warning-2"},
				nil,
			)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
		})

		When("updating the labels succeeds", func() {
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

			It("calls the right actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.UpdateDomainLabelsByDomainNameCallCount()).To(Equal(1))
			})

			It("displays a message in the right case", func() {
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

		Context("shows the right update message", func() {
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
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)

			expectedMap = map[string]types.NullString{
				"some-label":     types.NewNullString("some-value"),
				"some-other-key": types.NewNullString(),
			}
			labels = expectedMap

			fakeActor.UpdateOrganizationLabelsByOrganizationNameReturns(
				v7action.Warnings{"some-warning-1", "some-warning-2"},
				nil)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
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

			It("calls the right actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.UpdateOrganizationLabelsByOrganizationNameCallCount()).To(Equal(1))
			})

			It("displays a message in the right case", func() {
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

		Context("shows the right update message", func() {
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
				"some-other-key": types.NewNullString(),
			}
			labels = expectedMap

			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org"})
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "fake-space", GUID: "space-guid"})
			fakeActor.UpdateRouteLabelsReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
				nil)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
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

			It("calls the right actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.UpdateRouteLabelsCallCount()).To(Equal(1))
			})

			It("displays a message in the right case", func() {
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

		Context("shows the right update message", func() {
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
		const expectedServiceBrokerName = "my-broker"
		var (
			expectedMap map[string]types.NullString
			executeErr  error
		)

		BeforeEach(func() {
			targetResource = TargetResource{
				ResourceType: "service-broker",
				ResourceName: expectedServiceBrokerName,
			}

			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			expectedMap = map[string]types.NullString{
				"some-label":     types.NewNullString("some-value"),
				"some-other-key": types.NewNullString(),
			}
			labels = expectedMap
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
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

			It("calls the right actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameCallCount()).To(Equal(1))
			})

			It("displays a message in the right case", func() {
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

		Context("shows the right update message", func() {
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

	When("updating labels on service-instance", func() {
		const serviceInstanceName = "some-service-instance"

		var (
			executeErr  error
			expectedMap map[string]types.NullString
		)

		BeforeEach(func() {
			targetResource = TargetResource{
				ResourceType: "service-instance",
				ResourceName: serviceInstanceName,
			}

			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org"})
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "fake-space", GUID: "some-space-guid"})
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)

			expectedMap = map[string]types.NullString{
				"some-label":     types.NewNullString("some-value"),
				"some-other-key": types.NewNullString(),
			}
			labels = expectedMap
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
		})

		When("updating the service instance labels succeeds", func() {
			BeforeEach(func() {
				fakeActor.UpdateServiceInstanceLabelsReturns(v7action.Warnings{"some-warning-1", "some-warning-2"}, nil)
			})

			It("prints all warnings and does not return an error ", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

			It("passes the correct parameters into the actor", func() {
				Expect(fakeActor.UpdateServiceInstanceLabelsCallCount()).To(Equal(1))
				actualServiceInstance, spaceGUID, labelsMap := fakeActor.UpdateServiceInstanceLabelsArgsForCall(0)
				Expect(actualServiceInstance).To(Equal(serviceInstanceName))
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(labelsMap).To(Equal(expectedMap))
			})
		})

		When("the resource type argument is not lowercase", func() {
			BeforeEach(func() {
				targetResource.ResourceType = "serViCE-iNSTance"
			})

			It("calls the right actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.UpdateServiceInstanceLabelsCallCount()).To(Equal(1))
			})

			It("displays a message in the right case", func() {
				Expect(testUI.Out).To(Say("(.*) label\\(s\\) for service-instance (.*)"))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("updating the labels fails", func() {
			BeforeEach(func() {
				fakeActor.UpdateServiceInstanceLabelsReturns(
					v7action.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("api call failed"),
				)
			})

			It("prints all warnings", func() {
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("api call failed"))
			})
		})

		Context("shows the right update message with org and space", func() {
			When("Unsetting labels", func() {
				BeforeEach(func() {
					cmd.Action = Unset
				})

				It("shows 'Removing' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for service-instance %s in org fake-org / space fake-space as some-user...`), serviceInstanceName))
				})
			})

			When("Setting labels", func() {
				BeforeEach(func() {
					cmd.Action = Set
				})

				It("shows 'Setting' as action", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for service-instance %s in org fake-org / space fake-space as some-user...`), serviceInstanceName))
				})
			})
		})
	})

	When("updating labels on service-offering", func() {
		var executeErr error

		const serviceBrokerName = "brokerName"
		const serviceOfferingName = "serviceOfferingName"

		BeforeEach(func() {
			targetResource = TargetResource{
				ResourceType: "service-offering",
				ResourceName: serviceOfferingName,
			}
			labels = map[string]types.NullString{
				"some-label":     types.NewNullString("some-value"),
				"some-other-key": types.NewNullString(),
			}

			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			fakeActor.UpdateServiceOfferingLabelsReturns(
				v7action.Warnings{"some-warning-1", "some-warning-2"},
				nil,
			)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
		})

		When("updating the labels succeeds", func() {
			It("does not return an error and prints all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

			It("passes the correct parameters into the actor", func() {
				Expect(fakeActor.UpdateServiceOfferingLabelsCallCount()).To(Equal(1))
				gotServiceOfferingName, gotBrokerName, gotLabelsMap := fakeActor.UpdateServiceOfferingLabelsArgsForCall(0)
				Expect(gotServiceOfferingName).To(Equal(serviceOfferingName))
				Expect(gotBrokerName).To(BeEmpty())
				Expect(gotLabelsMap).To(Equal(labels))
			})

			When("a service broker name is specified", func() {
				BeforeEach(func() {
					targetResource.ServiceBroker = serviceBrokerName
				})

				It("passes the broker name", func() {
					Expect(fakeActor.UpdateServiceOfferingLabelsCallCount()).To(Equal(1))
					_, gotBrokerName, _ := fakeActor.UpdateServiceOfferingLabelsArgsForCall(0)
					Expect(gotBrokerName).To(Equal(serviceBrokerName))
				})
			})
		})

		When("the resource type argument is not lowercase", func() {
			BeforeEach(func() {
				targetResource.ResourceType = "Service-OffErinG"
			})

			It("calls the right actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.UpdateServiceOfferingLabelsCallCount()).To(Equal(1))
			})

			It("displays a message in the right case", func() {
				Expect(testUI.Out).To(Say("(.*) label\\(s\\) for service-offering (.*)"))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("updating the labels fails", func() {
			BeforeEach(func() {
				fakeActor.UpdateServiceOfferingLabelsReturns(
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

		Context("shows the right update message", func() {
			When("the broker name is not specified", func() {
				When("Unsetting labels", func() {
					BeforeEach(func() {
						cmd.Action = Unset
					})

					It("shows 'Removing' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for service-offering %s as some-user...`), serviceOfferingName))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				When("Setting labels", func() {
					BeforeEach(func() {
						cmd.Action = Set
					})

					It("shows 'Setting' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for service-offering %s as some-user...`), serviceOfferingName))
						Expect(testUI.Out).To(Say("OK"))
					})
				})
			})

			When("the broker name is specified", func() {
				BeforeEach(func() {
					targetResource.ServiceBroker = serviceBrokerName
				})

				When("Unsetting labels", func() {
					BeforeEach(func() {
						cmd.Action = Unset
					})

					It("shows 'Removing' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for service-offering %s from service broker %s as some-user...`), serviceOfferingName, serviceBrokerName))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				When("Setting labels", func() {
					BeforeEach(func() {
						cmd.Action = Set
					})

					It("shows 'Setting' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for service-offering %s from service broker %s as some-user...`), serviceOfferingName, serviceBrokerName))
						Expect(testUI.Out).To(Say("OK"))
					})
				})
			})
		})
	})

	When("updating labels on service-plan", func() {
		var executeErr error

		const serviceBrokerName = "brokerName"
		const serviceOfferingName = "serviceOfferingName"
		const servicePlanName = "servicePlanName"

		BeforeEach(func() {
			targetResource = TargetResource{
				ResourceType: "service-plan",
				ResourceName: servicePlanName,
			}
			labels = map[string]types.NullString{
				"some-label":     types.NewNullString("some-value"),
				"some-other-key": types.NewNullString(),
			}

			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			fakeActor.UpdateServicePlanLabelsReturns(
				v7action.Warnings{"some-warning-1", "some-warning-2"},
				nil,
			)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
		})

		When("updating the labels succeeds", func() {
			It("does not return an error and prints all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

			It("passes the correct parameters into the actor", func() {
				Expect(fakeActor.UpdateServicePlanLabelsCallCount()).To(Equal(1))
				gotServicePlanName, gotServiceOfferingName, gotBrokerName, gotLabelsMap := fakeActor.UpdateServicePlanLabelsArgsForCall(0)
				Expect(gotServicePlanName).To(Equal(servicePlanName))
				Expect(gotServiceOfferingName).To(BeEmpty())
				Expect(gotBrokerName).To(BeEmpty())
				Expect(gotLabelsMap).To(Equal(labels))
			})

			When("a service broker name is specified", func() {
				BeforeEach(func() {
					targetResource.ServiceBroker = serviceBrokerName
				})

				It("passes the broker name", func() {
					Expect(fakeActor.UpdateServicePlanLabelsCallCount()).To(Equal(1))
					_, _, gotBrokerName, _ := fakeActor.UpdateServicePlanLabelsArgsForCall(0)
					Expect(gotBrokerName).To(Equal(serviceBrokerName))
				})
			})

			When("a service offering name is specified", func() {
				BeforeEach(func() {
					targetResource.ServiceOffering = serviceOfferingName
				})

				It("passes the broker name", func() {
					Expect(fakeActor.UpdateServicePlanLabelsCallCount()).To(Equal(1))
					_, gotOfferingName, _, _ := fakeActor.UpdateServicePlanLabelsArgsForCall(0)
					Expect(gotOfferingName).To(Equal(serviceOfferingName))
				})
			})
		})

		When("the resource type argument is not lowercase", func() {
			BeforeEach(func() {
				targetResource.ResourceType = "Service-PlAN"
			})

			It("calls the right actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.UpdateServicePlanLabelsCallCount()).To(Equal(1))
			})

			It("displays a message in the right case", func() {
				Expect(testUI.Out).To(Say("(.*) label\\(s\\) for service-plan (.*)"))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("updating the labels fails", func() {
			BeforeEach(func() {
				fakeActor.UpdateServicePlanLabelsReturns(
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

		Context("shows the right update message", func() {
			When("no extra flags are specified", func() {
				When("Unsetting labels", func() {
					BeforeEach(func() {
						cmd.Action = Unset
					})

					It("shows 'Removing' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for service-plan %s as some-user...`), servicePlanName))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				When("Setting labels", func() {
					BeforeEach(func() {
						cmd.Action = Set
					})

					It("shows 'Setting' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for service-plan %s as some-user...`), servicePlanName))
						Expect(testUI.Out).To(Say("OK"))
					})
				})
			})

			When("the broker name is specified", func() {
				BeforeEach(func() {
					targetResource.ServiceBroker = serviceBrokerName
				})

				When("Unsetting labels", func() {
					BeforeEach(func() {
						cmd.Action = Unset
					})

					It("shows 'Removing' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for service-plan %s from service broker %s as some-user...`), servicePlanName, serviceBrokerName))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				When("Setting labels", func() {
					BeforeEach(func() {
						cmd.Action = Set
					})

					It("shows 'Setting' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for service-plan %s from service broker %s as some-user...`), servicePlanName, serviceBrokerName))
						Expect(testUI.Out).To(Say("OK"))
					})
				})
			})

			When("the offering name is specified", func() {
				BeforeEach(func() {
					targetResource.ServiceOffering = serviceOfferingName
				})

				When("Unsetting labels", func() {
					BeforeEach(func() {
						cmd.Action = Unset
					})

					It("shows 'Removing' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for service-plan %s from service offering %s as some-user...`), servicePlanName, serviceOfferingName))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				When("Setting labels", func() {
					BeforeEach(func() {
						cmd.Action = Set
					})

					It("shows 'Setting' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for service-plan %s from service offering %s as some-user...`), servicePlanName, serviceOfferingName))
						Expect(testUI.Out).To(Say("OK"))
					})
				})
			})

			When("both the offering name and the broker name are specified", func() {
				BeforeEach(func() {
					targetResource.ServiceBroker = serviceBrokerName
					targetResource.ServiceOffering = serviceOfferingName
				})

				When("Unsetting labels", func() {
					BeforeEach(func() {
						cmd.Action = Unset
					})

					It("shows 'Removing' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for service-plan %s from service offering %s / service broker %s as some-user...`), servicePlanName, serviceOfferingName, serviceBrokerName))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				When("Setting labels", func() {
					BeforeEach(func() {
						cmd.Action = Set
					})

					It("shows 'Setting' as action", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for service-plan %s from service offering %s / service broker %s as some-user...`), servicePlanName, serviceOfferingName, serviceBrokerName))
						Expect(testUI.Out).To(Say("OK"))
					})
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
				"some-other-key": types.NewNullString(),
			}
			labels = expectedMap

			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			fakeConfig.TargetedOrganizationReturns(
				configv3.Organization{Name: "fake-org", GUID: "some-org-guid"})

			fakeActor.UpdateSpaceLabelsBySpaceNameReturns(
				v7action.Warnings{"some-warning-1", "some-warning-2"},
				nil)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(targetResource, labels)
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

			It("calls the right actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.UpdateSpaceLabelsBySpaceNameCallCount()).To(Equal(1))
			})

			It("displays a message in the right case", func() {
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

		Context("shows the right update message", func() {
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
		const stackName = "some-stack"

		var (
			executeErr  error
			expectedMap map[string]types.NullString
		)

		BeforeEach(func() {
			targetResource = TargetResource{
				ResourceType: "stack",
				ResourceName: stackName,
			}
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
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

			It("calls the right actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.UpdateStackLabelsByStackNameCallCount()).To(Equal(1))
			})

			It("displays a message in the right case", func() {
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

		Context("shows the right update message", func() {
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
