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
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("set-label command", func() {
	var (
		cmd             SetLabelCommand
		resourceName    string
		fakeActor       *v7fakes.FakeSetLabelActor
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		testUI          *ui.UI

		executeErr error
	)

	Context("shared validations", func() {
		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeActor = new(v7fakes.FakeSetLabelActor)
			fakeConfig = new(commandfakes.FakeConfig)
			fakeSharedActor = new(commandfakes.FakeSharedActor)
			cmd = SetLabelCommand{
				Actor:       fakeActor,
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
			}
		})

		DescribeTable(
			"Combination of --stack with resource type",
			func(resourceType string) {
				cmd.BuildpackStack = "cflinuxfs3"
				cmd.RequiredArgs = flag.SetLabelArgs{
					ResourceType: resourceType,
				}

				err := cmd.Execute(nil)

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

		When("some provided labels do not have a value part", func() {
			BeforeEach(func() {
				cmd.RequiredArgs = flag.SetLabelArgs{
					ResourceType: "anything",
					ResourceName: resourceName,
					Labels:       []string{"FOO=BAR", "MISSING_EQUALS", "ENV=FAKE"},
				}
			})

			It("complains about the missing equal sign", func() {
				err := cmd.Execute(nil)
				Expect(err).To(MatchError("Metadata error: no value provided for label 'MISSING_EQUALS'"))
				Expect(err).To(HaveOccurred())
			})
		})

		When("fetching the current user's name fails", func() {
			BeforeEach(func() {
				cmd.RequiredArgs = flag.SetLabelArgs{
					ResourceType: "anything",
					ResourceName: resourceName,
				}
				fakeConfig.CurrentUserNameReturns("some-user", errors.New("boom"))
			})

			It("returns an error", func() {
				err := cmd.Execute(nil)
				Expect(err).To(MatchError("boom"))
			})
		})
	})

	When("setting labels on apps", func() {

		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeActor = new(v7fakes.FakeSetLabelActor)
			fakeConfig = new(commandfakes.FakeConfig)
			fakeSharedActor = new(commandfakes.FakeSharedActor)
			cmd = SetLabelCommand{
				Actor:       fakeActor,
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
			}
			cmd.RequiredArgs = flag.SetLabelArgs{
				ResourceType: "app",
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
			Expect(checkSpace).To(BeTrue())
		})

		When("checking target succeeds", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(nil)
			})

			When("fetching current user's name succeeds", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("some-user", nil)
					fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org"})
					fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "fake-space", GUID: "some-space-guid"})

					u, err := uuid.NewV4()
					Expect(err).NotTo(HaveOccurred())
					resourceName = u.String()
				})

				When("all the provided labels are valid", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "app",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateApplicationLabelsByApplicationNameReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							nil,
						)
					})

					When("updating the application labels succeeds", func() {
						It("sets the provided labels on the app", func() {
							name, spaceGUID, labels := fakeActor.UpdateApplicationLabelsByApplicationNameArgsForCall(0)
							Expect(name).To(Equal(resourceName), "failed to pass app name")
							Expect(spaceGUID).To(Equal("some-space-guid"))
							Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
								"FOO": types.NewNullString("BAR"),
								"ENV": types.NewNullString("FAKE"),
							}))
						})

						It("displays a message", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))

							Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for app %s in org fake-org / space fake-space as some-user...`), resourceName))
							Expect(testUI.Out).To(Say("OK"))
						})

						It("prints all warnings", func() {
							Expect(testUI.Err).To(Say("some-warning-1"))
							Expect(testUI.Err).To(Say("some-warning-2"))
						})
					})
				})

				When("updating the application labels fail", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "app",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateApplicationLabelsByApplicationNameReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							errors.New("some-updating-error"),
						)
					})
					It("displays warnings and an error message", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
						Expect(executeErr).To(HaveOccurred())
						Expect(executeErr).To(MatchError("some-updating-error"))
					})
				})

				When("some provided labels do not have a value part", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "app",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "MISSING_EQUALS", "ENV=FAKE"},
						}
					})

					It("complains about the missing equal sign", func() {
						Expect(executeErr).To(MatchError("Metadata error: no value provided for label 'MISSING_EQUALS'"))
					})
				})

				When("when the --stack flag is specified", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "app",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						cmd.BuildpackStack = "im-a-stack"
					})

					It("complains about the --stack flag being present", func() {
						Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{Args: []string{"app", "--stack, -s"}}))
					})
				})

				When("the resource type argument is not lowercase", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "ApP",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateApplicationLabelsByApplicationNameReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							nil,
						)
					})

					It("sets the provided labels on the app", func() {
						name, spaceGUID, labels := fakeActor.UpdateApplicationLabelsByApplicationNameArgsForCall(0)
						Expect(name).To(Equal(resourceName), "failed to pass app name")
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
							"FOO": types.NewNullString("BAR"),
							"ENV": types.NewNullString("FAKE"),
						}))
					})

					It("prints the flavor text in lowercase", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for app %s`), resourceName))
					})
				})
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
	})

	When("setting labels on domains", func() {
		BeforeEach(func() {
			resourceName = "example.com"
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeActor = new(v7fakes.FakeSetLabelActor)
			fakeConfig = new(commandfakes.FakeConfig)
			fakeSharedActor = new(commandfakes.FakeSharedActor)
			cmd = SetLabelCommand{
				Actor:       fakeActor,
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
			}
			cmd.RequiredArgs = flag.SetLabelArgs{
				ResourceType: "domain",
				ResourceName: resourceName,
				Labels:       []string{"FOO=BAR", "ENV=FAKE"},
			}

			fakeConfig.CurrentUserNameReturns("some-user", nil)

			fakeActor.UpdateDomainLabelsByDomainNameReturns(
				v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
				nil,
			)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		It("doesn't error", func() {
			Expect(executeErr).ToNot(HaveOccurred())
		})

		It("outputs that the label is being set", func() {
			Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for domain %s as some-user...`), resourceName))
			Expect(testUI.Out).To(Say("OK"))
		})

		It("sets the provided labels on the domain", func() {
			Expect(fakeActor.UpdateDomainLabelsByDomainNameCallCount()).To(Equal(1))
			name, labels := fakeActor.UpdateDomainLabelsByDomainNameArgsForCall(0)
			Expect(name).To(Equal(resourceName), "failed to pass domain name")
			Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
				"FOO": types.NewNullString("BAR"),
				"ENV": types.NewNullString("FAKE"),
			}))
		})

		It("prints all warnings", func() {
			Expect(testUI.Err).To(Say("some-warning-1"))
			Expect(testUI.Err).To(Say("some-warning-2"))
		})

		When("updating the domain labels fails", func() {
			BeforeEach(func() {
				fakeActor.UpdateDomainLabelsByDomainNameReturns(
					v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
					errors.New("some-updating-error"),
				)
			})
			It("displays warnings and an error message", func() {
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError("some-updating-error"))
			})
		})

		When("some provided labels do not have a value part", func() {
			BeforeEach(func() {
				cmd.RequiredArgs = flag.SetLabelArgs{
					ResourceType: "domain",
					ResourceName: resourceName,
					Labels:       []string{"FOO=BAR", "MISSING_EQUALS", "ENV=FAKE"},
				}
			})

			It("complains about the missing equal sign", func() {
				Expect(executeErr).To(MatchError("Metadata error: no value provided for label 'MISSING_EQUALS'"))
			})
		})

		When("the resource type argument is not lowercase", func() {
			BeforeEach(func() {
				cmd.RequiredArgs = flag.SetLabelArgs{
					ResourceType: "DoMaiN",
					ResourceName: resourceName,
					Labels:       []string{"FOO=BAR", "ENV=FAKE"},
				}
			})

			It("sets the provided labels on the domain", func() {
				Expect(fakeActor.UpdateDomainLabelsByDomainNameCallCount()).To(Equal(1))
				name, labels := fakeActor.UpdateDomainLabelsByDomainNameArgsForCall(0)
				Expect(name).To(Equal(resourceName), "failed to pass domain name")
				Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
					"FOO": types.NewNullString("BAR"),
					"ENV": types.NewNullString("FAKE"),
				}))
			})

			It("prints the flavor text in lowercase", func() {
				Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for domain %s as some-user...`), resourceName))
			})
		})
	})

	When("setting labels on orgs", func() {
		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeActor = new(v7fakes.FakeSetLabelActor)
			fakeConfig = new(commandfakes.FakeConfig)
			fakeSharedActor = new(commandfakes.FakeSharedActor)
			resourceName = "some-org"
			cmd = SetLabelCommand{
				Actor:       fakeActor,
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
			}
			cmd.RequiredArgs = flag.SetLabelArgs{
				ResourceType: "org",
				ResourceName: resourceName,
			}
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

		When("checking target succeeds", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(nil)
			})

			When("fetching current user's name succeeds", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("some-user", nil)
				})

				When("all the provided labels are valid", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "org",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateOrganizationLabelsByOrganizationNameReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							nil,
						)
					})

					When("updating the application labels succeeds", func() {
						It("sets the provided labels on the app", func() {
							Expect(fakeActor.UpdateOrganizationLabelsByOrganizationNameCallCount()).To(Equal(1))
							name, labels := fakeActor.UpdateOrganizationLabelsByOrganizationNameArgsForCall(0)
							Expect(name).To(Equal(resourceName), "failed to pass app name")
							Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
								"FOO": types.NewNullString("BAR"),
								"ENV": types.NewNullString("FAKE"),
							}))
						})

						It("displays a message", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))

							Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for org %s as some-user...`), resourceName))
							Expect(testUI.Out).To(Say("OK"))
						})

						It("prints all warnings", func() {
							Expect(testUI.Err).To(Say("some-warning-1"))
							Expect(testUI.Err).To(Say("some-warning-2"))
						})
					})
				})

				When("updating the application labels fail", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "org",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateOrganizationLabelsByOrganizationNameReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							errors.New("some-updating-error"),
						)
					})

					It("displays warnings and an error message", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
						Expect(executeErr).To(MatchError("some-updating-error"))
					})
				})

				When("some provided labels do not have a value part", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "org",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "MISSING_EQUALS", "ENV=FAKE"},
						}
					})

					It("complains about the missing equal sign", func() {
						Expect(executeErr).To(MatchError("Metadata error: no value provided for label 'MISSING_EQUALS'"))
						Expect(executeErr).To(HaveOccurred())
					})
				})

				When("when the --stack flag is specified", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "org",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						cmd.BuildpackStack = "im-a-stack"
					})

					It("complains about the --stack flag being present", func() {
						Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{Args: []string{"org", "--stack, -s"}}))
					})
				})

				When("the resource type argument is not lowercase", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "OrG",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateOrganizationLabelsByOrganizationNameReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							nil,
						)
					})

					It("sets the provided labels on the org", func() {
						name, labels := fakeActor.UpdateOrganizationLabelsByOrganizationNameArgsForCall(0)
						Expect(name).To(Equal(resourceName), "failed to pass org name")
						Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
							"FOO": types.NewNullString("BAR"),
							"ENV": types.NewNullString("FAKE"),
						}))
					})

					It("prints the flavor text in lowercase", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for org %s`), resourceName))
					})
				})
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
	})

	When("setting labels on routes", func() {

		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeActor = new(v7fakes.FakeSetLabelActor)
			fakeConfig = new(commandfakes.FakeConfig)
			fakeSharedActor = new(commandfakes.FakeSharedActor)
			cmd = SetLabelCommand{
				Actor:       fakeActor,
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
			}
			cmd.RequiredArgs = flag.SetLabelArgs{
				ResourceType: "route",
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
			Expect(checkSpace).To(BeTrue())
		})

		When("checking target succeeds", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(nil)
			})

			When("fetching current user's name succeeds", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("some-user", nil)
					fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org"})
					fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "fake-space", GUID: "some-space-guid"})

					u, err := uuid.NewV4()
					Expect(err).NotTo(HaveOccurred())
					resourceName = u.String()
				})

				When("all the provided labels are valid", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "route",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateRouteLabelsReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							nil,
						)
					})

					When("updating the route labels succeeds", func() {
						It("sets the provided labels on the route", func() {
							name, spaceGUID, labels := fakeActor.UpdateRouteLabelsArgsForCall(0)
							Expect(name).To(Equal(resourceName), "failed to pass route name")
							Expect(spaceGUID).To(Equal("some-space-guid"))
							Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
								"FOO": types.NewNullString("BAR"),
								"ENV": types.NewNullString("FAKE"),
							}))
						})

						It("displays a message", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))

							Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for route %s in org fake-org / space fake-space as some-user...`), resourceName))
							Expect(testUI.Out).To(Say("OK"))
						})

						It("prints all warnings", func() {
							Expect(testUI.Err).To(Say("some-warning-1"))
							Expect(testUI.Err).To(Say("some-warning-2"))
						})
					})
				})

				When("updating the route labels fail", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "route",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateRouteLabelsReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							errors.New("some-updating-error"),
						)
					})
					It("displays warnings and an error message", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
						Expect(executeErr).To(HaveOccurred())
						Expect(executeErr).To(MatchError("some-updating-error"))
					})
				})

				When("some provided labels do not have a value part", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "route",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "MISSING_EQUALS", "ENV=FAKE"},
						}
					})

					It("complains about the missing equal sign", func() {
						Expect(executeErr).To(MatchError("Metadata error: no value provided for label 'MISSING_EQUALS'"))
					})
				})

				When("when the --stack flag is specified", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "route",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						cmd.BuildpackStack = "im-a-stack"
					})

					It("complains about the --stack flag being present", func() {
						Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{Args: []string{"route", "--stack, -s"}}))
					})
				})

				When("the resource type argument is not lowercase", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "RoUTe",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateRouteLabelsReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							nil,
						)
					})

					It("sets the provided labels on the route", func() {
						name, spaceGUID, labels := fakeActor.UpdateRouteLabelsArgsForCall(0)
						Expect(name).To(Equal(resourceName), "failed to pass route name")
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
							"FOO": types.NewNullString("BAR"),
							"ENV": types.NewNullString("FAKE"),
						}))
					})

					It("prints the flavor text in lowercase", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for route %s`), resourceName))
					})
				})
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
	})

	When("setting labels on buildpacks", func() {
		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeActor = new(v7fakes.FakeSetLabelActor)
			fakeConfig = new(commandfakes.FakeConfig)
			fakeSharedActor = new(commandfakes.FakeSharedActor)
			resourceName = "some-buildpack"
			cmd = SetLabelCommand{
				Actor:       fakeActor,
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
			}
			cmd.RequiredArgs = flag.SetLabelArgs{
				ResourceType: "buildpack",
				ResourceName: resourceName,
			}
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

		When("checking target succeeds", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(nil)
			})

			When("fetching current user's name succeeds", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("some-user", nil)
				})

				When("all the provided labels are valid", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "buildpack",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						cmd.BuildpackStack = "globinski"

						fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							nil,
						)
					})

					When("updating the buildpack labels succeeds", func() {
						When("the buildpack stack is specified", func() {
							It("sets the provided labels on the buildpack", func() {
								Expect(fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackCallCount()).To(Equal(1))
								name, stack, labels := fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackArgsForCall(0)
								Expect(name).To(Equal(resourceName), "failed to pass buildpack name")
								Expect(stack).To(Equal("globinski"), "failed to pass stack name")
								Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
									"FOO": types.NewNullString("BAR"),
									"ENV": types.NewNullString("FAKE"),
								}))
							})

							It("displays a message", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))

								Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for buildpack %s with stack %s as some-user...`), resourceName, cmd.BuildpackStack))
								Expect(testUI.Out).To(Say("OK"))
							})

							It("prints all warnings", func() {
								Expect(testUI.Err).To(Say("some-warning-1"))
								Expect(testUI.Err).To(Say("some-warning-2"))
							})
						})

						When("the buildpack stack is not specified", func() {
							BeforeEach(func() {
								cmd.BuildpackStack = ""
							})

							It("sets the provided labels on the buildpack", func() {
								Expect(fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackCallCount()).To(Equal(1))
								name, stack, labels := fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackArgsForCall(0)
								Expect(name).To(Equal(resourceName), "failed to pass buildpack name")
								Expect(stack).To(Equal(""), "got a non-empty stack name")
								Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
									"FOO": types.NewNullString("BAR"),
									"ENV": types.NewNullString("FAKE"),
								}))
							})

							It("displays a message", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))

								Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for buildpack %s as some-user...`), resourceName))
								Expect(testUI.Out).To(Say("OK"))
							})

							It("prints all warnings", func() {
								Expect(testUI.Err).To(Say("some-warning-1"))
								Expect(testUI.Err).To(Say("some-warning-2"))
							})
						})

						When("no stack is provided", func() {
							BeforeEach(func() {
								cmd.BuildpackStack = ""
							})

							It("displays a message that includes the stack name", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
								Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for buildpack %s as some-user...`), resourceName))

								Expect(testUI.Out).To(Say("OK"))
							})
						})
					})
				})

				When("updating the buildpack labels fail", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "buildpack",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							errors.New("some-updating-error"),
						)
					})

					It("displays warnings and an error message", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
						Expect(executeErr).To(MatchError("some-updating-error"))
					})
				})

				When("some provided labels do not have a value part", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "buildpack",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "MISSING_EQUALS", "ENV=FAKE"},
						}
					})

					It("complains about the missing equal sign", func() {
						Expect(executeErr).To(MatchError("Metadata error: no value provided for label 'MISSING_EQUALS'"))
						Expect(executeErr).To(HaveOccurred())
					})
				})

				When("the resource type argument is not lowercase", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "bUiLdPaCk",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							nil,
						)
					})

					It("sets the provided labels on the buildpack", func() {
						name, stack, labels := fakeActor.UpdateBuildpackLabelsByBuildpackNameAndStackArgsForCall(0)
						Expect(name).To(Equal(resourceName), "failed to pass buildpack name")
						Expect(stack).To(Equal(""), "failed to pass buildpack stack")
						Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
							"FOO": types.NewNullString("BAR"),
							"ENV": types.NewNullString("FAKE"),
						}))
					})

					It("prints the flavor text in lowercase", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for buildpack %s`), resourceName))
					})

					When("setting the stack argument", func() {
						BeforeEach(func() {
							cmd.BuildpackStack = "cflinuxfs3"
						})

						It("does not display an argument combination error", func() {
							Expect(executeErr).ToNot(HaveOccurred())
						})

					})
				})
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

		When("setting the stack argument", func() {

			BeforeEach(func() {
				cmd.BuildpackStack = "cflinuxfs3"
			})

			It("does not display an argument combination error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
			})

		})
	})

	When("setting labels on spaces", func() {
		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeActor = new(v7fakes.FakeSetLabelActor)
			fakeConfig = new(commandfakes.FakeConfig)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org", GUID: "some-org-guid"})

			fakeSharedActor = new(commandfakes.FakeSharedActor)
			resourceName = "some-space"
			cmd = SetLabelCommand{
				Actor:       fakeActor,
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
			}
			cmd.RequiredArgs = flag.SetLabelArgs{
				ResourceType: "space",
				ResourceName: resourceName,
			}
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		It("checks that the user is logged in and targeted to an org", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkOrg).To(BeTrue())
			Expect(checkSpace).To(BeFalse())
		})

		When("checking target succeeds", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(nil)
			})

			When("fetching current user's name succeeds", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("some-user", nil)
				})

				When("all the provided labels are valid", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "space",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateSpaceLabelsBySpaceNameReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							nil,
						)
					})

					When("updating the space labels succeeds", func() {
						It("sets the provided labels on the space", func() {
							Expect(fakeActor.UpdateSpaceLabelsBySpaceNameCallCount()).To(Equal(1))
							spaceName, orgGUID, labels := fakeActor.UpdateSpaceLabelsBySpaceNameArgsForCall(0)
							Expect(spaceName).To(Equal(resourceName), "failed to pass space name")
							Expect(orgGUID).To(Equal("some-org-guid"))
							Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
								"FOO": types.NewNullString("BAR"),
								"ENV": types.NewNullString("FAKE"),
							}))
						})

						It("displays a message", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
							Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for space %s in org fake-org as some-user...`), resourceName))
							Expect(testUI.Out).To(Say("OK"))
						})

						It("prints all warnings", func() {
							Expect(testUI.Err).To(Say("some-warning-1"))
							Expect(testUI.Err).To(Say("some-warning-2"))
						})
					})
				})

				When("updating the application labels fail", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "org",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateOrganizationLabelsByOrganizationNameReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							errors.New("some-updating-error"),
						)
					})

					It("displays warnings and an error message", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
						Expect(executeErr).To(MatchError("some-updating-error"))
					})
				})

				When("some provided labels do not have a value part", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "org",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "MISSING_EQUALS", "ENV=FAKE"},
						}
					})

					It("complains about the missing equal sign", func() {
						Expect(executeErr).To(MatchError("Metadata error: no value provided for label 'MISSING_EQUALS'"))
						Expect(executeErr).To(HaveOccurred())
					})
				})

				When("the resource type argument is not lowercase", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "sPaCe",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateSpaceLabelsBySpaceNameReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							nil,
						)
					})

					It("sets the provided labels on the app", func() {
						name, orgGUID, labels := fakeActor.UpdateSpaceLabelsBySpaceNameArgsForCall(0)
						Expect(name).To(Equal(resourceName), "failed to pass space name")
						Expect(orgGUID).To(Equal("some-org-guid"))
						Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
							"FOO": types.NewNullString("BAR"),
							"ENV": types.NewNullString("FAKE"),
						}))
					})

					It("prints the flavor text in lowercase", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for space %s`), resourceName))
					})
				})
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
	})

	When("setting labels on stacks", func() {
		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeActor = new(v7fakes.FakeSetLabelActor)
			fakeConfig = new(commandfakes.FakeConfig)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org", GUID: "some-org-guid"})

			fakeSharedActor = new(commandfakes.FakeSharedActor)
			resourceName = "some-stack"
			cmd = SetLabelCommand{
				Actor:       fakeActor,
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
			}
			cmd.RequiredArgs = flag.SetLabelArgs{
				ResourceType: "stack",
				ResourceName: resourceName,
			}
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		It("checks that the user is logged in but not necessarily targeted to an org", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkOrg).To(BeFalse())
			Expect(checkSpace).To(BeFalse())
		})

		When("checking target succeeds", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(nil)
			})

			When("fetching current user's name succeeds", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("some-user", nil)
				})

				When("all the provided labels are valid", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "stack",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateStackLabelsByStackNameReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							nil,
						)
					})

					When("updating the stack labels succeeds", func() {
						It("sets the provided labels on the stack", func() {
							Expect(fakeActor.UpdateStackLabelsByStackNameCallCount()).To(Equal(1))
							stackName, labels := fakeActor.UpdateStackLabelsByStackNameArgsForCall(0)
							Expect(stackName).To(Equal(resourceName), "failed to pass stack name")
							Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
								"FOO": types.NewNullString("BAR"),
								"ENV": types.NewNullString("FAKE"),
							}))
						})

						It("displays a message", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
							Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for stack %s as some-user...`), resourceName))
							Expect(testUI.Out).To(Say("OK"))
						})

						It("prints all warnings", func() {
							Expect(testUI.Err).To(Say("some-warning-1"))
							Expect(testUI.Err).To(Say("some-warning-2"))
						})
					})
				})

				When("updating the application labels fail", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "org",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateOrganizationLabelsByOrganizationNameReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							errors.New("some-updating-error"),
						)
					})

					It("displays warnings and an error message", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
						Expect(executeErr).To(MatchError("some-updating-error"))
					})
				})

				When("some provided labels do not have a value part", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "org",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "MISSING_EQUALS", "ENV=FAKE"},
						}
					})

					It("complains about the missing equal sign", func() {
						Expect(executeErr).To(MatchError("Metadata error: no value provided for label 'MISSING_EQUALS'"))
						Expect(executeErr).To(HaveOccurred())
					})
				})

				When("the resource type argument is not lowercase", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "sTaCk",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateStackLabelsByStackNameReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							nil,
						)
					})

					It("sets the provided labels on the stack", func() {
						name, labels := fakeActor.UpdateStackLabelsByStackNameArgsForCall(0)
						Expect(name).To(Equal(resourceName), "failed to pass stack name")
						Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
							"FOO": types.NewNullString("BAR"),
							"ENV": types.NewNullString("FAKE"),
						}))
					})

					It("prints the flavor text in lowercase", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for stack %s`), resourceName))
					})
				})
			})
		})

		When("checking targeted org and stack fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(errors.New("nope"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("nope"))
			})
		})
	})

	When("setting labels on service-brokers", func() {
		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeActor = new(v7fakes.FakeSetLabelActor)
			fakeConfig = new(commandfakes.FakeConfig)
			//fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org", GUID: "some-org-guid"})

			fakeSharedActor = new(commandfakes.FakeSharedActor)
			resourceName = "some-service-broker"
			cmd = SetLabelCommand{
				Actor:       fakeActor,
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
			}
			cmd.RequiredArgs = flag.SetLabelArgs{
				ResourceType: "service-broker",
				ResourceName: resourceName,
			}
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		It("checks that the user is logged in but not necessarily targeted to an org", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkOrg).To(BeFalse())
			Expect(checkSpace).To(BeFalse())
		})

		When("checking target succeeds", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(nil)
			})

			When("fetching current user's name succeeds", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("some-user", nil)
				})

				When("all the provided labels are valid", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "service-broker",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							nil,
						)
					})

					When("updating the service-broker labels succeeds", func() {
						It("sets the provided labels on the service-broker", func() {
							Expect(executeErr).NotTo(HaveOccurred())
							Expect(fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameCallCount()).To(Equal(1))
							serviceBroker, labels := fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameArgsForCall(0)
							Expect(serviceBroker).To(Equal(resourceName))
							Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
								"FOO": types.NewNullString("BAR"),
								"ENV": types.NewNullString("FAKE"),
							}))
						})

						It("displays a message", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for service-broker %s as some-user...`), resourceName))
							Expect(testUI.Out).To(Say("OK"))
						})

						It("prints all warnings", func() {
							Expect(testUI.Err).To(Say("some-warning-1"))
							Expect(testUI.Err).To(Say("some-warning-2"))
						})
					})
				})

				When("updating the service-broker labels fail", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "service-broker",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							errors.New("some-updating-error"),
						)
					})

					It("displays warnings and an error message", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
						Expect(executeErr).To(MatchError("some-updating-error"))
					})
				})

				When("the resource type argument is not lowercase", func() {
					BeforeEach(func() {
						cmd.RequiredArgs = flag.SetLabelArgs{
							ResourceType: "sErvicE-BrokEr",
							ResourceName: resourceName,
							Labels:       []string{"FOO=BAR", "ENV=FAKE"},
						}
						fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameReturns(
							v7action.Warnings([]string{"some-warning-1", "some-warning-2"}),
							nil,
						)
					})

					It("sets the provided labels on the service-broker", func() {
						name, labels := fakeActor.UpdateServiceBrokerLabelsByServiceBrokerNameArgsForCall(0)
						Expect(name).To(Equal(resourceName), "failed to pass stack name")
						Expect(labels).To(BeEquivalentTo(map[string]types.NullString{
							"FOO": types.NewNullString("BAR"),
							"ENV": types.NewNullString("FAKE"),
						}))
					})

					It("prints the flavor text in lowercase", func() {
						Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Setting label(s) for service-broker %s`), resourceName))
					})
				})
			})
		})

		When("checking targeted org and stack fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(errors.New("nope"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("nope"))
			})
		})
	})

	When("an unrecognized resource is specified", func() {
		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeActor = new(v7fakes.FakeSetLabelActor)
			fakeConfig = new(commandfakes.FakeConfig)
			fakeSharedActor = new(commandfakes.FakeSharedActor)
			resourceName = "some-unrecognized-resource"
			cmd = SetLabelCommand{
				Actor:       fakeActor,
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
			}
			cmd.RequiredArgs = flag.SetLabelArgs{
				ResourceType: "unrecognized-resource",
				ResourceName: resourceName,
			}
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		It("errors", func() {
			Expect(executeErr).To(MatchError("Unsupported resource type of 'unrecognized-resource'"))
		})
	})
})
