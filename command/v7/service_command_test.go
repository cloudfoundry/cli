package v7_test

import (
	"errors"
	"strings"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("service command", func() {
	const (
		serviceInstanceName = "fake-service-instance-name"
		serviceInstanceGUID = "fake-service-instance-guid"
		spaceName           = "fake-space-name"
		spaceGUID           = "fake-space-guid"
		orgName             = "fake-org-name"
		username            = "fake-user-name"
	)
	var (
		cmd             ServiceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
	)

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = ServiceCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		fakeActor.GetCurrentUserReturns(configv3.User{Name: username}, nil)

		fakeConfig.TargetedSpaceReturns(configv3.Space{
			GUID: spaceGUID,
			Name: spaceName,
		})

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: orgName,
		})

		fakeActor.GetServiceInstanceDetailsReturns(
			v7action.ServiceInstanceDetails{
				ServiceInstance: resources.ServiceInstance{
					GUID: serviceInstanceGUID,
					Name: serviceInstanceName,
				},
				SharedStatus: v7action.SharedStatus{
					IsSharedToOtherSpaces: false,
				},
			},
			v7action.Warnings{"warning one", "warning two"},
			nil,
		)

		setPositionalFlags(&cmd, serviceInstanceName)
	})

	It("checks the user is logged in, and targeting an org and space", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		orgChecked, spaceChecked := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(orgChecked).To(BeTrue())
		Expect(spaceChecked).To(BeTrue())
	})

	When("the --guid flag is specified", func() {
		BeforeEach(func() {
			fakeActor.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{
					GUID: serviceInstanceGUID,
					Name: serviceInstanceName,
				},
				v7action.Warnings{"warning one", "warning two"},
				nil,
			)

			setFlag(&cmd, "--guid")
		})

		It("looks up the service instance and prints the GUID and no warnings", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
			actualName, actualSpaceGUID := fakeActor.GetServiceInstanceByNameAndSpaceArgsForCall(0)
			Expect(actualName).To(Equal(serviceInstanceName))
			Expect(actualSpaceGUID).To(Equal(spaceGUID))

			Expect(testUI.Out).To(Say(`^%s\n$`, serviceInstanceGUID))
			Expect(testUI.Err).NotTo(Say("warning"))
		})

		When("getting the service instance fails", func() {
			BeforeEach(func() {
				fakeActor.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{
						GUID: serviceInstanceGUID,
						Name: serviceInstanceName,
					},
					v7action.Warnings{"warning one", "warning two"},
					errors.New("yuck"),
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("yuck"))
			})
		})
	})

	When("it is a user-provided service instance", func() {
		const (
			routeServiceURL          = "https://route.com"
			syslogURL                = "https://syslog.com"
			tags                     = "foo, bar"
			lastOperationType        = "update"
			lastOperationState       = "succeeded"
			lastOperationDescription = "doing amazing work"
			lastOperationStartTime   = "a second ago"
			lastOperationUpdatedTime = "just now"
		)

		BeforeEach(func() {
			fakeActor.GetServiceInstanceDetailsReturns(
				v7action.ServiceInstanceDetails{
					ServiceInstance: resources.ServiceInstance{
						GUID:            serviceInstanceGUID,
						Name:            serviceInstanceName,
						Type:            resources.UserProvidedServiceInstance,
						SyslogDrainURL:  types.NewOptionalString(syslogURL),
						RouteServiceURL: types.NewOptionalString(routeServiceURL),
						Tags:            types.NewOptionalStringSlice(strings.Split(tags, ", ")...),
						LastOperation: resources.LastOperation{
							Type:        lastOperationType,
							State:       lastOperationState,
							Description: lastOperationDescription,
							CreatedAt:   lastOperationStartTime,
							UpdatedAt:   lastOperationUpdatedTime,
						},
					},
				},
				v7action.Warnings{"warning one", "warning two"},
				nil,
			)
		})

		It("looks up the service instance and prints the details and warnings", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(fakeActor.GetServiceInstanceDetailsCallCount()).To(Equal(1))
			actualName, actualSpaceGUID, actualOmitApps := fakeActor.GetServiceInstanceDetailsArgsForCall(0)
			Expect(actualName).To(Equal(serviceInstanceName))
			Expect(actualSpaceGUID).To(Equal(spaceGUID))
			Expect(actualOmitApps).To(BeFalse())

			Expect(testUI.Out).To(SatisfyAll(
				Say(`Showing info of service %s in org %s / space %s as %s...\n`, serviceInstanceName, orgName, spaceName, username),
				Say(`\n`),
				Say(`name:\s+%s\n`, serviceInstanceName),
				Say(`guid:\s+\S+\n`),
				Say(`type:\s+user-provided`),
				Say(`tags:\s+%s\n`, tags),
				Say(`route service url:\s+%s\n`, routeServiceURL),
				Say(`syslog drain url:\s+%s\n`, syslogURL),
				Say(`\n`),
				Say(`Showing status of last operation:\n`),
				Say(`\s+status:\s+%s %s\n`, lastOperationType, lastOperationState),
				Say(`\s+message:\s+%s\n`, lastOperationDescription),
				Say(`\s+started:\s+%s\n`, lastOperationStartTime),
				Say(`\s+updated:\s+%s\n`, lastOperationUpdatedTime),
				Say(`\n`),
				Say(`Showing bound apps:\n`),
				Say(`There are no bound apps for this service instance\.\n`),
			))

			Expect(testUI.Err).To(SatisfyAll(
				Say("warning one"),
				Say("warning two"),
			))
		})

		When("last operation is not set", func() {
			BeforeEach(func() {
				fakeActor.GetServiceInstanceDetailsReturns(
					v7action.ServiceInstanceDetails{
						ServiceInstance: resources.ServiceInstance{
							GUID:            serviceInstanceGUID,
							Name:            serviceInstanceName,
							Type:            resources.UserProvidedServiceInstance,
							SyslogDrainURL:  types.NewOptionalString(syslogURL),
							RouteServiceURL: types.NewOptionalString(routeServiceURL),
							Tags:            types.NewOptionalStringSlice(strings.Split(tags, ", ")...),
						},
					},
					v7action.Warnings{"warning one", "warning two"},
					nil,
				)
			})

			It("when last operation does not exist", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(fakeActor.GetServiceInstanceDetailsCallCount()).To(Equal(1))
				actualName, actualSpaceGUID, actualOmitApps := fakeActor.GetServiceInstanceDetailsArgsForCall(0)
				Expect(actualName).To(Equal(serviceInstanceName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualOmitApps).To(BeFalse())

				Expect(testUI.Out).To(SatisfyAll(
					Say(`Showing info of service %s in org %s / space %s as %s...\n`, serviceInstanceName, orgName, spaceName, username),
					Say(`\n`),
					Say(`name:\s+%s\n`, serviceInstanceName),
					Say(`guid:\s+\S+\n`),
					Say(`type:\s+user-provided`),
					Say(`tags:\s+%s\n`, tags),
					Say(`route service url:\s+%s\n`, routeServiceURL),
					Say(`syslog drain url:\s+%s\n`, syslogURL),
					Say(`\n`),
					Say(`Showing status of last operation:\n`),
					Say(`There is no last operation available for this service instance\.\n`),
					Say(`\n`),
					Say(`Showing bound apps:\n`),
					Say(`There are no bound apps for this service instance\.\n`),
				))

				Expect(testUI.Err).To(SatisfyAll(
					Say("warning one"),
					Say("warning two"),
				))
			})
		})

		When("there are bound apps", func() {
			BeforeEach(func() {
				fakeActor.GetServiceInstanceDetailsReturns(
					v7action.ServiceInstanceDetails{
						ServiceInstance: resources.ServiceInstance{
							GUID:            serviceInstanceGUID,
							Name:            serviceInstanceName,
							Type:            resources.UserProvidedServiceInstance,
							SyslogDrainURL:  types.NewOptionalString(syslogURL),
							RouteServiceURL: types.NewOptionalString(routeServiceURL),
							Tags:            types.NewOptionalStringSlice(strings.Split(tags, ", ")...),
						},
						BoundApps: []resources.ServiceCredentialBinding{
							{
								Name:    "named-binding",
								AppName: "app-1",
								LastOperation: resources.LastOperation{
									Type:        resources.CreateOperation,
									State:       resources.OperationSucceeded,
									Description: "great",
								},
							},
							{
								AppName: "app-2",
								LastOperation: resources.LastOperation{
									Type:        resources.UpdateOperation,
									State:       resources.OperationFailed,
									Description: "sorry",
								},
							},
						},
					},
					v7action.Warnings{"warning one", "warning two"},
					nil,
				)
			})

			It("prints the bound apps table", func() {
				Expect(testUI.Out).To(SatisfyAll(
					Say(`Showing bound apps:\n`),
					Say(`   name\s+binding name\s+status\s+message\n`),
					Say(`   app-1\s+named-binding\s+create succeeded\s+great\n`),
					Say(`   app-2\s+update failed\s+sorry\n`),
				))
			})
		})
	})

	When("it is a managed service instance", func() {
		const (
			dashboardURL               = "https://dashboard.com"
			tags                       = "foo, bar"
			servicePlanName            = "fake-service-plan-name"
			serviceOfferingName        = "fake-service-offering-name"
			serviceOfferingDescription = "an amazing service"
			serviceOfferingDocs        = "https://service.docs.com"
			serviceBrokerName          = "fake-service-broker-name"
			lastOperationType          = "create"
			lastOperationState         = "in progress"
			lastOperationDescription   = "doing amazing work"
			lastOperationStartTime     = "a second ago"
			lastOperationUpdatedTime   = "just now"
		)

		BeforeEach(func() {
			fakeActor.GetServiceInstanceDetailsReturns(
				v7action.ServiceInstanceDetails{
					ServiceInstance: resources.ServiceInstance{
						GUID:         serviceInstanceGUID,
						Name:         serviceInstanceName,
						Type:         resources.ManagedServiceInstance,
						DashboardURL: types.NewOptionalString(dashboardURL),
						Tags:         types.NewOptionalStringSlice(strings.Split(tags, ", ")...),
						LastOperation: resources.LastOperation{
							Type:        lastOperationType,
							State:       lastOperationState,
							Description: lastOperationDescription,
							CreatedAt:   lastOperationStartTime,
							UpdatedAt:   lastOperationUpdatedTime,
						},
						SpaceGUID: spaceGUID,
					},
					SpaceName:        spaceName,
					OrganizationName: orgName,
					ServiceOffering: resources.ServiceOffering{
						Name:             serviceOfferingName,
						Description:      serviceOfferingDescription,
						DocumentationURL: serviceOfferingDocs,
					},
					ServicePlan:       resources.ServicePlan{Name: servicePlanName},
					ServiceBrokerName: serviceBrokerName,
					SharedStatus: v7action.SharedStatus{
						IsSharedToOtherSpaces: true,
						UsageSummary:          []v7action.UsageSummaryWithSpaceAndOrg{{"shared-to-space", "some-org", 3}},
					},
				},
				v7action.Warnings{"warning one", "warning two"},
				nil,
			)
		})

		It("looks up the service instance and prints the details and warnings", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(fakeActor.GetServiceInstanceDetailsCallCount()).To(Equal(1))
			actualName, actualSpaceGUID, actualOmitApps := fakeActor.GetServiceInstanceDetailsArgsForCall(0)
			Expect(actualName).To(Equal(serviceInstanceName))
			Expect(actualSpaceGUID).To(Equal(spaceGUID))
			Expect(actualOmitApps).To(BeFalse())

			Expect(testUI.Out).To(SatisfyAll(
				Say(`Showing info of service %s in org %s / space %s as %s...\n`, serviceInstanceName, orgName, spaceName, username),
				Say(`\n`),
				Say(`name:\s+%s\n`, serviceInstanceName),
				Say(`guid:\s+\S+\n`),
				Say(`type:\s+managed`),
				Say(`broker:\s+%s`, serviceBrokerName),
				Say(`offering:\s+%s`, serviceOfferingName),
				Say(`plan:\s+%s`, servicePlanName),
				Say(`tags:\s+%s\n`, tags),
				Say(`description:\s+%s\n`, serviceOfferingDescription),
				Say(`documentation:\s+%s\n`, serviceOfferingDocs),
				Say(`dashboard url:\s+%s\n`, dashboardURL),
				Say(`\n`),
				Say(`Showing status of last operation:`),
				Say(`status:\s+%s %s\n`, lastOperationType, lastOperationState),
				Say(`message:\s+%s\n`, lastOperationDescription),
				Say(`started:\s+%s\n`, lastOperationStartTime),
				Say(`updated:\s+%s\n`, lastOperationUpdatedTime),
				Say(`\n`),
				Say(`Showing bound apps:`),
				Say(`There are no bound apps for this service instance\.\n`),
				Say(`\n`),
				Say(`Showing sharing info:`),
				Say(`Shared with spaces:\n`),
				Say(`Showing upgrade status:`),
				Say(`Upgrades are not supported by this broker.\n`),
			))

			Expect(testUI.Err).To(SatisfyAll(
				Say("warning one"),
				Say("warning two"),
			))
		})

		Context("sharing", func() {
			When("targeting original space", func() {
				When("service instance is shared", func() {
					It("shows shared information", func() {
						Expect(testUI.Out).To(SatisfyAll(
							Say(`Showing sharing info:`),
							Say(`Shared with spaces:`),
							Say(`org\s+space\s+bindings\s*\n`),
							Say(`some-org\s+shared-to-space\s+3\s*\n`),
						))
					})
				})

				When("service is not shared", func() {
					BeforeEach(func() {
						fakeActor.GetServiceInstanceDetailsReturns(
							v7action.ServiceInstanceDetails{
								ServiceInstance: resources.ServiceInstance{
									SpaceGUID: spaceGUID,
								},
								SharedStatus: v7action.SharedStatus{
									IsSharedToOtherSpaces: false,
								},
							},
							v7action.Warnings{},
							nil,
						)
					})

					It("displays that the service is not shared", func() {
						Expect(testUI.Out).To(SatisfyAll(
							Say(`Showing sharing info:`),
							Say(`This service instance is not currently being shared.`),
						))
					})
				})

				When("the service instance sharing feature is disabled", func() {
					BeforeEach(func() {
						fakeActor.GetServiceInstanceDetailsReturns(
							v7action.ServiceInstanceDetails{
								ServiceInstance: resources.ServiceInstance{
									SpaceGUID: spaceGUID,
								},
								SharedStatus: v7action.SharedStatus{
									FeatureFlagIsDisabled: true,
								},
							},
							v7action.Warnings{},
							nil,
						)
					})

					It("displays that the sharing feature is disabled", func() {
						Expect(testUI.Out).To(SatisfyAll(
							Say(`Showing sharing info:\n`),
							Say(`\n`),
							Say(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform.\n`),
							Say(`\n`),
						))
					})
				})

				When("the service instance sharing feature is enabled", func() {
					BeforeEach(func() {
						fakeActor.GetServiceInstanceDetailsReturns(
							v7action.ServiceInstanceDetails{
								ServiceInstance: resources.ServiceInstance{},
								SharedStatus: v7action.SharedStatus{
									FeatureFlagIsDisabled: false,
								},
							},
							v7action.Warnings{},
							nil,
						)
					})

					It("does not display a warning", func() {
						Expect(testUI.Out).NotTo(
							Say(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform.`),
						)
					})
				})

				When("the offering does not allow service instance sharing", func() {
					BeforeEach(func() {
						fakeActor.GetServiceInstanceDetailsReturns(
							v7action.ServiceInstanceDetails{
								ServiceInstance: resources.ServiceInstance{
									SpaceGUID: spaceGUID,
								},
								SharedStatus: v7action.SharedStatus{
									OfferingDisablesSharing: true,
								},
							},
							v7action.Warnings{},
							nil,
						)
					})

					It("displays that the sharing feature is disabled", func() {
						Expect(testUI.Out).To(SatisfyAll(
							Say(`Showing sharing info:\n`),
							Say(`\n`),
							Say(`Service instance sharing is disabled for this service offering.\n`),
							Say(`\n`),
						))
					})
				})

				When("the offering does allow service instance sharing", func() {
					BeforeEach(func() {
						fakeActor.GetServiceInstanceDetailsReturns(
							v7action.ServiceInstanceDetails{
								ServiceInstance: resources.ServiceInstance{},
								SharedStatus: v7action.SharedStatus{
									OfferingDisablesSharing: false,
								},
							},
							v7action.Warnings{},
							nil,
						)
					})

					It("does not display a warning", func() {
						Expect(testUI.Out).NotTo(
							Say(`Service instance sharing is disabled for this service offering.`),
						)
					})
				})
			})

			When("targeting shared to space", func() {
				BeforeEach(func() {
					sharedToSpaceGUID := "fake-shared-to-space-guid"
					sharedToSpaceName := "fake-shared-to-space-name"
					fakeConfig.TargetedSpaceReturns(configv3.Space{
						GUID: sharedToSpaceGUID,
						Name: sharedToSpaceName,
					})

					fakeActor.GetServiceInstanceDetailsReturns(
						v7action.ServiceInstanceDetails{
							ServiceInstance: resources.ServiceInstance{
								SpaceGUID: spaceGUID,
							},
							SpaceName:        spaceName,
							OrganizationName: orgName,
							SharedStatus: v7action.SharedStatus{
								IsSharedFromOriginalSpace: true,
								IsSharedToOtherSpaces:     true,
								OfferingDisablesSharing:   true,
								FeatureFlagIsDisabled:     true,
							},
						},
						v7action.Warnings{},
						nil,
					)
				})

				It("shows original space information", func() {
					Expect(testUI.Out).To(SatisfyAll(
						Say(`Showing sharing info:\n`),
						Say(`This service instance is shared from space %s of org %s.\n`, spaceName, orgName),
					))
				})

				It("doesn't display any sharing warning", func() {
					Expect(testUI.Out).NotTo(SatisfyAny(
						Say(`This service instance is currently shared.`),
						Say(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform.`),
						Say(`Service instance sharing is disabled for this service offering.`),
					))
				})
			})
		})

		Context("upgrading", func() {
			When("an upgrade is available", func() {
				BeforeEach(func() {
					fakeActor.GetServiceInstanceDetailsReturns(
						v7action.ServiceInstanceDetails{
							ServiceInstance: resources.ServiceInstance{
								GUID:         serviceInstanceGUID,
								Name:         serviceInstanceName,
								Type:         resources.ManagedServiceInstance,
								DashboardURL: types.NewOptionalString(dashboardURL),
								Tags:         types.NewOptionalStringSlice(strings.Split(tags, ", ")...),
								LastOperation: resources.LastOperation{
									Type:        lastOperationType,
									State:       lastOperationState,
									Description: lastOperationDescription,
									CreatedAt:   lastOperationStartTime,
									UpdatedAt:   lastOperationUpdatedTime,
								},
							},
							ServiceOffering: resources.ServiceOffering{
								Name:             serviceOfferingName,
								Description:      serviceOfferingDescription,
								DocumentationURL: serviceOfferingDocs,
							},
							ServicePlan:       resources.ServicePlan{Name: servicePlanName},
							ServiceBrokerName: serviceBrokerName,
							UpgradeStatus: v7action.ServiceInstanceUpgradeStatus{
								State:       v7action.ServiceInstanceUpgradeAvailable,
								Description: "really cool upgrade\nwith juicy bits",
							},
						},
						v7action.Warnings{"warning one", "warning two"},
						nil,
					)
				})

				It("says an upgrade is available and shows the description", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Out).To(SatisfyAll(
						Say(`Showing upgrade status:\n`),
						Say(`There is an upgrade available for this service.\n`),
						Say(`Upgrade description: really cool upgrade\n`),
						Say(`with juicy bits\n`),
						Say(`TIP: You can upgrade using 'cf upgrade-service %s'\n`, serviceInstanceName),
					))
				})
			})

			When("the service instance is up to date", func() {
				BeforeEach(func() {
					fakeActor.GetServiceInstanceDetailsReturns(
						v7action.ServiceInstanceDetails{
							ServiceInstance: resources.ServiceInstance{
								GUID:         serviceInstanceGUID,
								Name:         serviceInstanceName,
								Type:         resources.ManagedServiceInstance,
								DashboardURL: types.NewOptionalString(dashboardURL),
								Tags:         types.NewOptionalStringSlice(strings.Split(tags, ", ")...),
								LastOperation: resources.LastOperation{
									Type:        lastOperationType,
									State:       lastOperationState,
									Description: lastOperationDescription,
									CreatedAt:   lastOperationStartTime,
									UpdatedAt:   lastOperationUpdatedTime,
								},
							},
							ServiceOffering: resources.ServiceOffering{
								Name:             serviceOfferingName,
								Description:      serviceOfferingDescription,
								DocumentationURL: serviceOfferingDocs,
							},
							ServicePlan:       resources.ServicePlan{Name: servicePlanName},
							ServiceBrokerName: serviceBrokerName,
							UpgradeStatus: v7action.ServiceInstanceUpgradeStatus{
								State: v7action.ServiceInstanceUpgradeNotAvailable,
							},
						},
						v7action.Warnings{"warning one", "warning two"},
						nil,
					)
				})

				It("says an upgrade is available and shows the description", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Out).To(SatisfyAll(
						Say(`Showing upgrade status:\n`),
						Say(`There is no upgrade available for this service.\n`),
					))
				})
			})
		})

		Context("bound apps", func() {
			BeforeEach(func() {
				fakeActor.GetServiceInstanceDetailsReturns(
					v7action.ServiceInstanceDetails{
						ServiceInstance: resources.ServiceInstance{
							GUID:         serviceInstanceGUID,
							Name:         serviceInstanceName,
							Type:         resources.ManagedServiceInstance,
							DashboardURL: types.NewOptionalString(dashboardURL),
							Tags:         types.NewOptionalStringSlice(strings.Split(tags, ", ")...),
							LastOperation: resources.LastOperation{
								Type:        lastOperationType,
								State:       lastOperationState,
								Description: lastOperationDescription,
								CreatedAt:   lastOperationStartTime,
								UpdatedAt:   lastOperationUpdatedTime,
							},
							SpaceGUID: spaceGUID,
						},
						SpaceName:        spaceName,
						OrganizationName: orgName,
						ServiceOffering: resources.ServiceOffering{
							Name:             serviceOfferingName,
							Description:      serviceOfferingDescription,
							DocumentationURL: serviceOfferingDocs,
						},
						ServicePlan:       resources.ServicePlan{Name: servicePlanName},
						ServiceBrokerName: serviceBrokerName,
						SharedStatus: v7action.SharedStatus{
							IsSharedToOtherSpaces: true,
						},
						BoundApps: []resources.ServiceCredentialBinding{
							{
								Name:    "named-binding",
								AppName: "app-1",
								LastOperation: resources.LastOperation{
									Type:        resources.CreateOperation,
									State:       resources.OperationSucceeded,
									Description: "great",
								},
							},
							{
								AppName: "app-2",
								LastOperation: resources.LastOperation{
									Type:        resources.UpdateOperation,
									State:       resources.OperationFailed,
									Description: "sorry",
								},
							},
						},
					},
					v7action.Warnings{"warning one", "warning two"},
					nil,
				)
			})

			It("prints the bound apps table", func() {
				Expect(testUI.Out).To(SatisfyAll(
					Say(`Showing bound apps:\n`),
					Say(`name\s+binding name\s+status\s+message\n`),
					Say(`app-1\s+named-binding\s+create succeeded\s+great\n`),
					Say(`app-2\s+update failed\s+sorry\n`),
				))
			})
		})
	})

	When("the --params flag is specified", func() {
		BeforeEach(func() {
			setFlag(&cmd, "--params")
		})

		When("parameters are set", func() {
			BeforeEach(func() {
				fakeActor.GetServiceInstanceParametersReturns(
					map[string]interface{}{"foo": "bar"},
					v7action.Warnings{"warning one", "warning two"},
					nil,
				)
			})

			It("returns parameters JSON", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(fakeActor.GetServiceInstanceParametersCallCount()).To(Equal(1))
				actualName, actualSpaceGUID := fakeActor.GetServiceInstanceParametersArgsForCall(0)
				Expect(actualName).To(Equal(serviceInstanceName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))

				Expect(testUI.Out).To(SatisfyAll(
					Say(`\{\n`),
					Say(`  "foo": "bar"\n`),
					Say(`\}\n`),
				))
				Expect(testUI.Err).To(SatisfyAll(
					Say("warning one"),
					Say("warning two"),
				))
			})
		})

		When("there was a problem retrieving the parameters", func() {
			BeforeEach(func() {
				fakeActor.GetServiceInstanceParametersReturns(
					v7action.ServiceInstanceParameters{},
					v7action.Warnings{"warning one", "warning two"},
					errors.New("something happened"),
				)
			})

			It("displays the reason", func() {
				Expect(executeErr).To(MatchError("something happened"))
				Expect(testUI.Err).To(SatisfyAll(
					Say("warning one"),
					Say("warning two"),
				))
			})
		})

		When("the parameters are empty", func() {
			BeforeEach(func() {
				fakeActor.GetServiceInstanceParametersReturns(
					v7action.ServiceInstanceParameters{},
					v7action.Warnings{"warning one", "warning two"},
					nil,
				)
			})

			It("displays empty parameters", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Out).To(Say(`{}\n`))
			})
		})
	})

	When("there is a problem looking up the service instance", func() {
		BeforeEach(func() {
			fakeActor.GetServiceInstanceDetailsReturns(
				v7action.ServiceInstanceDetails{},
				v7action.Warnings{"warning one", "warning two"},
				errors.New("boom"),
			)
		})

		It("prints warnings and returns an error", func() {
			Expect(executeErr).To(MatchError("boom"))

			Expect(testUI.Err).To(SatisfyAll(
				Say("warning one"),
				Say("warning two"),
			))
		})
	})

	When("checking the target returns an error", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("explode"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("explode"))
		})
	})
})
