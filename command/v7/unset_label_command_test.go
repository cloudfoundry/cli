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

var _ = Describe("unset-label command", func() {
	var (
		cmd             UnsetLabelCommand
		fakeConfig      *commandfakes.FakeConfig
		testUI          *ui.UI
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeSetLabelActor
		executeErr      error
	)
	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeSetLabelActor)
		cmd = UnsetLabelCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}
	})

	Describe("shared validations", func() {
		When("getting the user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserNameReturns("", errors.New("could not get user"))
				cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}
			})

			It("returns the error", func() {
				err := cmd.Execute(nil)
				Expect(err).To(MatchError("could not get user"))
			})
		})

		DescribeTable(
			"when checking the target fails",
			func(resourceType string) {
				fakeSharedActor.CheckTargetReturns(errors.New("Target not found"))
				cmd.RequiredArgs = flag.UnsetLabelArgs{
					ResourceType: resourceType,
				}

				err := cmd.Execute(nil)
				Expect(err).To(MatchError("Target not found"))
			},
			Entry("app", "app"),
			Entry("buildpack", "buildpack"),
			// domain - does not check target
			Entry("org", "org"),
			// route - does not check target
			Entry("service-broker", "service-broker"),
			Entry("space", "space"),
			Entry("stack", "stack"),
		)
	})

	When("unsetting labels on apps", func() {
		var appName string

		BeforeEach(func() {
			cmd.RequiredArgs = flag.UnsetLabelArgs{
				ResourceType: "app",
			}

			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org"})
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "fake-space", GUID: "some-space-guid"})
			appName = "some-app"
			cmd.RequiredArgs.ResourceName = appName

			fakeConfig.CurrentUserNameReturns("some-user", nil)
			cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		It("checks that the user is logged in and targeted to an org and space", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkOrg).To(BeTrue())
			Expect(checkSpace).To(BeTrue())
		})

		It("informs the user that labels are being removed", func() {
			Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for app %s in org fake-org / space fake-space as some-user...`), appName))
		})

		When("updating the app labels succeeds", func() {
			var expectedMap map[string]types.NullString

			BeforeEach(func() {
				fakeActor.UpdateApplicationLabelsByApplicationNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
					nil)
				expectedMap = map[string]types.NullString{
					"some-label":     types.NewNullString(),
					"some-other-key": types.NewNullString()}
			})

			It("does not return an error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
			})

			It("prints all warnings", func() {
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

			When("the resource type argument is not lowercase", func() {
				BeforeEach(func() {
					cmd.RequiredArgs.ResourceType = "aPp"
				})

				It("passes the correct parameters into the actor", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeActor.UpdateApplicationLabelsByApplicationNameCallCount()).To(Equal(1))
					actualAppName, spaceGUID, labelsMap := fakeActor.UpdateApplicationLabelsByApplicationNameArgsForCall(0)
					Expect(actualAppName).To(Equal(appName))
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(labelsMap).To(Equal(expectedMap))
				})
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
	})

	When("unsetting labels on buildpacks", func() {
		var resourceName string

		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeConfig = new(commandfakes.FakeConfig)
			fakeSharedActor = new(commandfakes.FakeSharedActor)
			fakeActor = new(v7fakes.FakeSetLabelActor)
			resourceName = "some-buildpack"
			cmd = UnsetLabelCommand{
				Actor:       fakeActor,
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
			}
			cmd.RequiredArgs = flag.UnsetLabelArgs{
				ResourceType: "buildpack",
				ResourceName: resourceName,
			}
			cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}

			fakeSharedActor.CheckTargetReturns(nil)
			fakeConfig.CurrentUserNameReturns("some-user", nil)

		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
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

		When("all the provided labels are valid", func() {
			BeforeEach(func() {
				cmd.RequiredArgs = flag.UnsetLabelArgs{
					ResourceType: "buildpack",
					ResourceName: "some-buildpack",
					LabelKeys:    []string{"FOO", "ENV"},
				}

				fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackReturns(
					v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
					nil,
				)
			})

			When("updating the buildpack labels succeeds", func() {
				When("the stack is specified", func() {
					BeforeEach(func() {
						cmd.BuildpackStack = "globinski"
					})

					It("unsets the provided labels on the buildpack", func() {
						Expect(fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackCallCount()).To(Equal(1))
						name, stack, labels := fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackArgsForCall(0)
						Expect(name).To(Equal(resourceName), "failed to pass buildpack name")
						Expect(stack).To(Equal("globinski"), "failed to pass stack name")
						Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
							"FOO": types.NewNullString(),
							"ENV": types.NewNullString(),
						}))
					})

					It("displays a message", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))

						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for buildpack %s with stack %s as some-user...`), resourceName, cmd.BuildpackStack))
						Expect(testUI.Out).To(Say("OK"))
					})

					It("prints all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})
				})

				When("the stack is not specified", func() {
					It("unsets the provided labels on the buildpack", func() {
						Expect(fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackCallCount()).To(Equal(1))
						name, stack, labels := fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackArgsForCall(0)
						Expect(name).To(Equal(resourceName), "failed to pass buildpack name")
						Expect(stack).To(Equal(""), "failed to pass stack name")
						Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
							"FOO": types.NewNullString(),
							"ENV": types.NewNullString(),
						}))
					})

					It("displays a message", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))

						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for buildpack %s as some-user...`), resourceName))
						Expect(testUI.Out).To(Say("OK"))
					})

					It("prints all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})
				})
			})
		})

		When("the resource type is not lowercase", func() {
			BeforeEach(func() {
				cmd.RequiredArgs = flag.UnsetLabelArgs{
					ResourceType: "bUiLdPaCk",
					ResourceName: "some-buildpack",
					LabelKeys:    []string{"FOO", "ENV"},
				}

				fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackReturns(
					v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
					nil,
				)
			})

			When("updating the buildpack labels succeeds", func() {
				When("the stack is specified", func() {
					BeforeEach(func() {
						cmd.BuildpackStack = "globinski"
					})

					It("does not display an argument combination error", func() {
						Expect(executeErr).ToNot(HaveOccurred())
					})

					It("unsets the provided labels on the buildpack", func() {
						Expect(fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackCallCount()).To(Equal(1))
						name, stack, labels := fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackArgsForCall(0)
						Expect(name).To(Equal(resourceName), "failed to pass buildpack name")
						Expect(stack).To(Equal("globinski"), "failed to pass stack name")
						Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
							"FOO": types.NewNullString(),
							"ENV": types.NewNullString(),
						}))
					})

					It("displays a message", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))

						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for buildpack %s with stack globinski as some-user...`), resourceName))
						Expect(testUI.Out).To(Say("OK"))
					})

					It("prints all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})
				})

				When("the stack is not specified", func() {
					It("unsets the provided labels on the buildpack", func() {
						Expect(fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackCallCount()).To(Equal(1))
						name, stack, labels := fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackArgsForCall(0)
						Expect(name).To(Equal(resourceName), "failed to pass buildpack name")
						Expect(stack).To(Equal(""), "failed to pass stack name")
						Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
							"FOO": types.NewNullString(),
							"ENV": types.NewNullString(),
						}))
					})

					It("displays a message", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))

						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for buildpack %s as some-user...`), resourceName))
						Expect(testUI.Out).To(Say("OK"))
					})

					It("prints all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})
				})
			})
		})
	})

	When("unsetting labels on domains", func() {
		var resourceName string
		BeforeEach(func() {
			resourceName = "example.com"
			cmd.RequiredArgs = flag.UnsetLabelArgs{
				ResourceType: "domain",
				ResourceName: resourceName,
			}
			cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}

			fakeConfig.CurrentUserNameReturns("some-user", nil)
			fakeActor.UpdateDomainLabelsByDomainNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
				nil)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		It("doesn't error", func() {
			Expect(executeErr).ToNot(HaveOccurred())
		})

		It("informs the user that labels are being removed", func() {
			Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for domain %s as some-user...`), resourceName))
		})

		It("removes the provided labels from the domain", func() {
			Expect(fakeActor.UpdateDomainLabelsByDomainNameCallCount()).To(Equal(1))
			name, labels := fakeActor.UpdateDomainLabelsByDomainNameArgsForCall(0)
			Expect(name).To(Equal(resourceName), "failed to pass domain name")
			Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
				"some-label":     types.NewNullString(),
				"some-other-key": types.NewNullString(),
			}))
		})

		It("prints all warnings", func() {
			Expect(testUI.Err).To(Say("some-warning-1"))
			Expect(testUI.Err).To(Say("some-warning-2"))
		})

		When("the resource type argument is not lowercase", func() {
			BeforeEach(func() {
				cmd.RequiredArgs.ResourceType = "DoMaiN"
			})

			It("passes the correct parameters into the actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeActor.UpdateDomainLabelsByDomainNameCallCount()).To(Equal(1))

				name, labels := fakeActor.UpdateDomainLabelsByDomainNameArgsForCall(0)
				Expect(name).To(Equal(resourceName), "failed to pass domain name")
				Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
					"some-label":     types.NewNullString(),
					"some-other-key": types.NewNullString(),
				}))
			})
		})

		When("updating the domain labels fails", func() {
			BeforeEach(func() {
				fakeActor.UpdateDomainLabelsByDomainNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
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
	})

	When("unsetting labels on orgs", func() {
		BeforeEach(func() {
			cmd.RequiredArgs = flag.UnsetLabelArgs{
				ResourceType: "org",
			}
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		When("checking target succeeds", func() {
			var orgName = "some-org"

			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(nil)
				cmd.RequiredArgs.ResourceName = orgName

			})

			When("fetching current user's name succeeds", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("some-user", nil)
					cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}
				})

				It("informs the user that labels are being removed", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for org %s as some-user...`), orgName))
				})

				When("updating the org labels succeeds", func() {
					var expectedMap map[string]types.NullString

					BeforeEach(func() {
						fakeActor.UpdateOrganizationLabelsByOrganizationNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
							nil)
						expectedMap = map[string]types.NullString{
							"some-label":     types.NewNullString(),
							"some-other-key": types.NewNullString()}
					})

					It("does not return an error", func() {
						Expect(executeErr).ToNot(HaveOccurred())
					})

					It("prints all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})

					It("passes the correct parameters into the actor", func() {
						Expect(fakeActor.UpdateOrganizationLabelsByOrganizationNameCallCount()).To(Equal(1))
						actualOrgName, labelsMap := fakeActor.UpdateOrganizationLabelsByOrganizationNameArgsForCall(0)
						Expect(actualOrgName).To(Equal(orgName))
						Expect(labelsMap).To(Equal(expectedMap))
					})

					When("the resource type argument is not lowercase", func() {
						BeforeEach(func() {
							cmd.RequiredArgs.ResourceType = "OrG"
						})

						It("retrieves the labels associated with the org", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(fakeActor.UpdateOrganizationLabelsByOrganizationNameCallCount()).To(Equal(1))
							actualOrgName, labelsMap := fakeActor.UpdateOrganizationLabelsByOrganizationNameArgsForCall(0)
							Expect(actualOrgName).To(Equal(orgName))
							Expect(labelsMap).To(Equal(expectedMap))
						})
					})
				})
			})
		})
	})

	When("unsetting labels on routes", func() {
		var resourceName string
		BeforeEach(func() {
			resourceName = "some-route.example.com"
			cmd.RequiredArgs = flag.UnsetLabelArgs{
				ResourceType: "route",
				ResourceName: resourceName,
			}
			cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}

			fakeConfig.CurrentUserNameReturns("some-user", nil)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org"})
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "fake-space", GUID: "space-guid"})
			fakeActor.UpdateRouteLabelsReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
				nil)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		It("doesn't error", func() {
			Expect(executeErr).ToNot(HaveOccurred())
		})

		It("informs the user that labels are being removed", func() {
			Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for route %s in org fake-org / space fake-space as some-user...`), resourceName))
		})

		It("removes the provided labels from the route", func() {
			Expect(fakeActor.UpdateRouteLabelsCallCount()).To(Equal(1))
			name, spaceGUID, labels := fakeActor.UpdateRouteLabelsArgsForCall(0)
			Expect(name).To(Equal(resourceName), "failed to pass route name")
			Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
				"some-label":     types.NewNullString(),
				"some-other-key": types.NewNullString(),
			}))
			Expect(spaceGUID).To(Equal("space-guid"))
		})

		It("prints all warnings", func() {
			Expect(testUI.Err).To(Say("some-warning-1"))
			Expect(testUI.Err).To(Say("some-warning-2"))
		})

		When("the resource type argument is not lowercase", func() {
			BeforeEach(func() {
				cmd.RequiredArgs.ResourceType = "rouTE"
			})

			It("passes the correct parameters into the actor", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeActor.UpdateRouteLabelsCallCount()).To(Equal(1))

				name, spaceGUID, labels := fakeActor.UpdateRouteLabelsArgsForCall(0)

				Expect(name).To(Equal(resourceName), "failed to pass route name")
				Expect(spaceGUID).To(Equal("space-guid"))
				Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
					"some-label":     types.NewNullString(),
					"some-other-key": types.NewNullString(),
				}))
			})
		})

		When("updating the route labels fails", func() {
			BeforeEach(func() {
				fakeActor.UpdateRouteLabelsReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
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
	})

	When("unsetting labels on service-broker", func() {
		expectedServiceBrokerName := "my-broker"
		BeforeEach(func() {
			cmd.RequiredArgs = flag.UnsetLabelArgs{
				ResourceType: "service-broker",
			}

			cmd.RequiredArgs.ResourceName = expectedServiceBrokerName

			fakeConfig.CurrentUserNameReturns("some-user", nil)
			cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		It("checks that the user is logged in", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkOrg).To(BeFalse())
			Expect(checkSpace).To(BeFalse())
		})

		It("informs the user that labels are being removed", func() {
			Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for service-broker %s as some-user...`), expectedServiceBrokerName))
		})

		When("updating the service-broker labels succeeds", func() {
			var expectedMap map[string]types.NullString

			BeforeEach(func() {
				fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
					nil)
				expectedMap = map[string]types.NullString{
					"some-label":     types.NewNullString(),
					"some-other-key": types.NewNullString()}
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

			When("the resource type argument is not lowercase", func() {
				BeforeEach(func() {
					cmd.RequiredArgs.ResourceType = "sErVice-BroKer"
				})

				It("passes the correct parameters into the actor", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameCallCount()).To(Equal(1))
					serviceBrokerName, labelsMap := fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameArgsForCall(0)
					Expect(serviceBrokerName).To(Equal(expectedServiceBrokerName))
					Expect(labelsMap).To(Equal(expectedMap))
				})
			})
		})

		When("updating the service-broker labels fails", func() {
			BeforeEach(func() {
				fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
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
	})

	When("unsetting labels on spaces", func() {
		BeforeEach(func() {
			cmd.RequiredArgs = flag.UnsetLabelArgs{
				ResourceType: "space",
			}
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		It("doesn't error", func() {
			Expect(executeErr).ToNot(HaveOccurred())
		})

		It("checks that the user is logged in and targeted to an org and space", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkOrg).To(BeTrue())
			Expect(checkSpace).To(BeFalse())
		})

		When("checking the target succeeds", func() {
			var (
				spaceName   string
				expectedMap map[string]types.NullString
			)

			BeforeEach(func() {
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org", GUID: "some-org-guid"})
				spaceName = "spiff"
				cmd.RequiredArgs.ResourceName = spaceName
			})

			When("getting the current user succeeds", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("some-user", nil)
					cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}
				})

				It("informs the user that labels are being removed", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for space %s in org fake-org as some-user...`), spaceName))
				})

				When("updating the space labels succeeds", func() {
					BeforeEach(func() {
						fakeActor.UpdateSpaceLabelsBySpaceNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
							nil)
						expectedMap = map[string]types.NullString{
							"some-label":     types.NewNullString(),
							"some-other-key": types.NewNullString()}
					})

					It("does not return an error", func() {
						Expect(executeErr).ToNot(HaveOccurred())
					})

					It("prints all warnings", func() {
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

					When("the resource type argument is not lowercase", func() {
						BeforeEach(func() {
							cmd.RequiredArgs.ResourceType = "SpAcE"
						})

						It("retrieves the labels associated with the space", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(fakeActor.UpdateSpaceLabelsBySpaceNameCallCount()).To(Equal(1))
							actualSpaceName, orgGUID, labelsMap := fakeActor.UpdateSpaceLabelsBySpaceNameArgsForCall(0)
							Expect(actualSpaceName).To(Equal(spaceName))
							Expect(orgGUID).To(Equal("some-org-guid"))
							Expect(labelsMap).To(Equal(expectedMap))
						})
					})
				})

				When("updating the space labels fails", func() {
					BeforeEach(func() {
						fakeActor.UpdateSpaceLabelsBySpaceNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
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
			})
		})
	})

	When("unsetting labels on stacks", func() {
		BeforeEach(func() {
			cmd.RequiredArgs = flag.UnsetLabelArgs{
				ResourceType: "stack",
			}
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		When("checking target succeeds", func() {
			var (
				stackName   = "some-stack"
				expectedMap map[string]types.NullString
			)

			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(nil)
				cmd.RequiredArgs.ResourceName = stackName
				expectedMap = map[string]types.NullString{
					"some-label":     types.NewNullString(),
					"some-other-key": types.NewNullString(),
				}
			})

			When("fetching current user's name succeeds", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("some-user", nil)
					cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}
				})

				It("informs the user that labels are being removed", func() {
					Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Removing label(s) for stack %s as some-user...`), stackName))
				})

				When("updating the stack labels succeeds", func() {
					BeforeEach(func() {
						fakeActor.UpdateStackLabelsByStackNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
							nil)
					})

					It("does not return an error", func() {
						Expect(executeErr).ToNot(HaveOccurred())
					})

					It("prints all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})

					It("passes the correct parameters into the actor", func() {

						Expect(fakeActor.UpdateStackLabelsByStackNameCallCount()).To(Equal(1))
						actualStackName, labelsMap := fakeActor.UpdateStackLabelsByStackNameArgsForCall(0)
						Expect(actualStackName).To(Equal(stackName))
						Expect(labelsMap).To(Equal(expectedMap))
					})

					When("the resource type argument is not lowercase", func() {
						BeforeEach(func() {
							cmd.RequiredArgs.ResourceType = "sTaCk"
						})
						It("passes the correct parameters into the actor", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(fakeActor.UpdateStackLabelsByStackNameCallCount()).To(Equal(1))
							actualStackName, labelsMap := fakeActor.UpdateStackLabelsByStackNameArgsForCall(0)
							Expect(actualStackName).To(Equal(stackName))
							Expect(labelsMap).To(Equal(expectedMap))
						})
					})
				})
			})
		})
	})

	When("the resource type is invalid", func() {
		It("fails", func() {
			cmd.RequiredArgs = flag.UnsetLabelArgs{
				ResourceType: "invalid",
			}

			err := cmd.Execute(nil)
			Expect(err).To(MatchError("Unsupported resource type of 'invalid'"))
		})
	})

	Describe("disallowed --stack option", func() {
		When("specifying --stack", func() {
			It("complains", func() {
				names := []string{"app", "space", "stack", "org"}
				for _, name := range names {
					cmd.RequiredArgs = flag.UnsetLabelArgs{
						ResourceType: name,
						ResourceName: "oshkosh",
						LabelKeys:    []string{"FOO", "ENV"},
					}
					cmd.BuildpackStack = "cflinuxfs3"
					executeErr := cmd.Execute(nil)
					argumentCombinationError := translatableerror.ArgumentCombinationError{
						Args: []string{strings.ToLower(cmd.RequiredArgs.ResourceType), "--stack, -s"},
					}
					Expect(executeErr).To(MatchError(argumentCombinationError))
				}
			})
		})
	})
})
