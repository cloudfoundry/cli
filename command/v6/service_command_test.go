package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("service Command", func() {
	var (
		cmd             ServiceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeServiceActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeServiceActor)

		cmd = ServiceCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd.RequiredArgs.ServiceInstance = "some-service-instance"

		fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionMaintenanceInfoInSummaryV2)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("an error is encountered checking if the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrgArg, checkTargetedSpaceArg := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrgArg).To(BeTrue())
			Expect(checkTargetedSpaceArg).To(BeTrue())
		})
	})

	When("the user is logged in and an org and space are targeted", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space",
			})
		})

		When("getting the current user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("get-user-error"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("get-user-error"))
				Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))
			})
		})

		When("getting the current user succeeds", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			})

			When("the '--guid' flag is provided", func() {
				BeforeEach(func() {
					cmd.GUID = true
				})

				When("the service instance does not exist", func() {
					BeforeEach(func() {
						fakeActor.GetServiceInstanceByNameAndSpaceReturns(
							v2action.ServiceInstance{},
							v2action.Warnings{"get-service-instance-warning"},
							actionerror.ServiceInstanceNotFoundError{
								GUID: "non-existant-service-instance-guid",
								Name: "non-existant-service-instance",
							})
					})

					It("returns ServiceInstanceNotFoundError", func() {
						Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotFoundError{
							GUID: "non-existant-service-instance-guid",
							Name: "non-existant-service-instance",
						}))

						Expect(testUI.Err).To(Say("get-service-instance-warning"))

						Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
						serviceInstanceNameArg, spaceGUIDArg := fakeActor.GetServiceInstanceByNameAndSpaceArgsForCall(0)
						Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
						Expect(spaceGUIDArg).To(Equal("some-space-guid"))

						Expect(fakeActor.GetServiceInstanceSummaryByNameAndSpaceCallCount()).To(Equal(0))
					})
				})

				When("an error is encountered getting the service instance", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("get-service-instance-error")
						fakeActor.GetServiceInstanceByNameAndSpaceReturns(
							v2action.ServiceInstance{},
							v2action.Warnings{"get-service-instance-warning"},
							expectedErr,
						)
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(testUI.Err).To(Say("get-service-instance-warning"))

						Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
						serviceInstanceNameArg, spaceGUIDArg := fakeActor.GetServiceInstanceByNameAndSpaceArgsForCall(0)
						Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
						Expect(spaceGUIDArg).To(Equal("some-space-guid"))

						Expect(fakeActor.GetServiceInstanceSummaryByNameAndSpaceCallCount()).To(Equal(0))
					})
				})

				When("no errors are encountered getting the service instance", func() {
					BeforeEach(func() {
						fakeActor.GetServiceInstanceByNameAndSpaceReturns(
							v2action.ServiceInstance{
								GUID: "some-service-instance-guid",
								Name: "some-service-instance",
							},
							v2action.Warnings{"get-service-instance-warning"},
							nil,
						)
					})

					It("displays the service instance guid", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("some-service-instance-guid"))
						Expect(testUI.Err).To(Say("get-service-instance-warning"))

						Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
						serviceInstanceNameArg, spaceGUIDArg := fakeActor.GetServiceInstanceByNameAndSpaceArgsForCall(0)
						Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
						Expect(spaceGUIDArg).To(Equal("some-space-guid"))

						Expect(fakeActor.GetServiceInstanceSummaryByNameAndSpaceCallCount()).To(Equal(0))
					})
				})
			})

			When("the '--guid' flag is not provided", func() {
				When("the service instance does not exist", func() {
					BeforeEach(func() {
						fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
							v2action.ServiceInstanceSummary{},
							v2action.Warnings{"get-service-instance-summary-warning"},
							actionerror.ServiceInstanceNotFoundError{
								GUID: "non-existant-service-instance-guid",
								Name: "non-existant-service-instance",
							})
					})

					It("returns ServiceInstanceNotFoundError", func() {
						Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotFoundError{
							GUID: "non-existant-service-instance-guid",
							Name: "non-existant-service-instance",
						}))

						Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))

						Expect(testUI.Err).To(Say("get-service-instance-summary-warning"))

						Expect(fakeActor.GetServiceInstanceSummaryByNameAndSpaceCallCount()).To(Equal(1))
						serviceInstanceNameArg, spaceGUIDArg := fakeActor.GetServiceInstanceSummaryByNameAndSpaceArgsForCall(0)
						Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
						Expect(spaceGUIDArg).To(Equal("some-space-guid"))

						Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(0))
					})
				})

				When("an error is encountered getting the service instance summary", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("get-service-instance-summary-error")
						fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
							v2action.ServiceInstanceSummary{},
							v2action.Warnings{"get-service-instance-summary-warning"},
							expectedErr)
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))

						Expect(testUI.Err).To(Say("get-service-instance-summary-warning"))

						Expect(fakeActor.GetServiceInstanceSummaryByNameAndSpaceCallCount()).To(Equal(1))
						serviceInstanceNameArg, spaceGUIDArg := fakeActor.GetServiceInstanceSummaryByNameAndSpaceArgsForCall(0)
						Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
						Expect(spaceGUIDArg).To(Equal("some-space-guid"))

						Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(0))
					})
				})

				When("no errors are encountered getting the service instance summary", func() {
					When("the service instance is a managed service instance", func() {
						var returnedSummary v2action.ServiceInstanceSummary

						BeforeEach(func() {
							returnedSummary = v2action.ServiceInstanceSummary{
								ServiceInstance: v2action.ServiceInstance{
									Name:         "some-service-instance",
									Type:         constant.ManagedService,
									Tags:         []string{"tag-1", "tag-2", "tag-3"},
									DashboardURL: "some-dashboard",
									LastOperation: ccv2.LastOperation{
										Type:        "some-type",
										State:       "some-state",
										Description: "some-last-operation-description",
										UpdatedAt:   "some-updated-at-time",
										CreatedAt:   "some-created-at-time",
									},
								},
								ServicePlan: v2action.ServicePlan{
									Name: "some-plan",
								},
								Service: v2action.Service{
									Label:            "some-service",
									Description:      "some-description",
									DocumentationURL: "some-docs-url",
								},
							}

							fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
								returnedSummary,
								v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
								nil,
							)
						})

						When("the service instance is not shared and is not shareable", func() {
							BeforeEach(func() {
								returnedSummary.ServiceInstanceShareType = v2action.ServiceInstanceIsNotShared
								returnedSummary.Service.Extra.Shareable = false
								returnedSummary.Service.ServiceBrokerName = "broker-1"
								returnedSummary.BoundApplications = []v2action.BoundApplication{
									{
										AppName:            "app-1",
										ServiceBindingName: "binding-name-1",
									},
									{
										AppName:            "app-2",
										ServiceBindingName: "binding-name-2",
										LastOperation: v2action.LastOperation{
											Type:        "delete",
											State:       constant.LastOperationInProgress,
											Description: "10% complete",
											UpdatedAt:   "some-updated-at-time",
											CreatedAt:   "some-created-at-time",
										},
									},
									{
										AppName:            "app-3",
										ServiceBindingName: "binding-name-3",
										LastOperation: v2action.LastOperation{
											Type:        "create",
											State:       constant.LastOperationFailed,
											Description: "Binding failed",
											UpdatedAt:   "some-updated-at-time",
											CreatedAt:   "some-created-at-time",
										},
									},
									{
										AppName:            "app-4",
										ServiceBindingName: "binding-name-4",
										LastOperation: v2action.LastOperation{
											Type:        "create",
											State:       constant.LastOperationSucceeded,
											Description: "Binding created",
											UpdatedAt:   "some-updated-at-time",
											CreatedAt:   "some-created-at-time",
										},
									},
								}
								fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
									returnedSummary,
									v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
									nil)

							})

							It("displays the service instance summary, all warnings and bound applications in the correct position", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).To(Say(`name:\s+some-service-instance`))
								Expect(testUI.Out).ToNot(Say("shared from org/space:"))
								Expect(testUI.Out).To(Say(`service:\s+some-service`))
								Expect(testUI.Out).To(Say(`tags:\s+tag-1, tag-2, tag-3`))
								Expect(testUI.Out).To(Say(`plan:\s+some-plan`))
								Expect(testUI.Out).To(Say(`description:\s+some-description`))
								Expect(testUI.Out).To(Say(`documentation:\s+some-docs-url`))
								Expect(testUI.Out).To(Say(`dashboard:\s+some-dashboard`))
								Expect(testUI.Out).To(Say(`service broker:\s+broker-1`))
								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).ToNot(Say("shared with spaces:"))
								Expect(testUI.Out).ToNot(Say(`org\s+space\s+bindings`))
								Expect(testUI.Out).ToNot(Say("This service is not currently shared."))

								Expect(testUI.Out).To(Say(`Showing status of last operation from service some-service-instance\.\.\.`))
								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).To(Say(`status:\s+some-type some-state`))
								Expect(testUI.Out).To(Say(`message:\s+some-last-operation-description`))
								Expect(testUI.Out).To(Say(`started:\s+some-created-at-time`))
								Expect(testUI.Out).To(Say(`updated:\s+some-updated-at-time`))
								Expect(testUI.Out).To(Say("\n\n"))

								Expect(testUI.Out).To(Say("bound apps:"))
								Expect(testUI.Out).To(Say(`name\s+binding name\s+status\s+message`))
								Expect(testUI.Out).To(Say(`app-1\s+binding-name-1\s+`))
								Expect(testUI.Out).To(Say(`app-2\s+binding-name-2\s+delete in progress\s+10\% complete`))
								Expect(testUI.Out).To(Say(`app-3\s+binding-name-3\s+create failed\s+Binding failed`))
								Expect(testUI.Out).To(Say(`app-4\s+binding-name-4\s+create succeeded\s+Binding created`))

								Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
								Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))

								Expect(fakeActor.GetServiceInstanceSummaryByNameAndSpaceCallCount()).To(Equal(1))
								serviceInstanceNameArg, spaceGUIDArg := fakeActor.GetServiceInstanceSummaryByNameAndSpaceArgsForCall(0)
								Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
								Expect(spaceGUIDArg).To(Equal("some-space-guid"))

								Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(0))
							})
						})

						When("the service instance is not shared and is shareable", func() {
							BeforeEach(func() {
								returnedSummary.ServiceInstanceShareType = v2action.ServiceInstanceIsNotShared
								returnedSummary.ServiceInstanceSharingFeatureFlag = true
								returnedSummary.Service.Extra.Shareable = true
								fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
									returnedSummary,
									v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
									nil)
							})

							It("displays the service instance summary and all warnings", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).To(Say(`name:\s+some-service-instance`))
								Expect(testUI.Out).ToNot(Say("shared from org/space:"))
								Expect(testUI.Out).To(Say(`service:\s+some-service`))
								Expect(testUI.Out).To(Say(`tags:\s+tag-1, tag-2, tag-3`))
								Expect(testUI.Out).To(Say(`plan:\s+some-plan`))
								Expect(testUI.Out).To(Say(`description:\s+some-description`))
								Expect(testUI.Out).To(Say(`documentation:\s+some-docs-url`))
								Expect(testUI.Out).To(Say(`dashboard:\s+some-dashboard`))
								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).ToNot(Say("shared with spaces:"))
								Expect(testUI.Out).ToNot(Say(`org\s+space\s+bindings`))
								Expect(testUI.Out).To(Say("This service is not currently shared."))
								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).To(Say(`Showing status of last operation from service some-service-instance\.\.\.`))
								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).To(Say(`status:\s+some-type some-state`))
								Expect(testUI.Out).To(Say(`message:\s+some-last-operation-description`))
								Expect(testUI.Out).To(Say(`started:\s+some-created-at-time`))
								Expect(testUI.Out).To(Say(`updated:\s+some-updated-at-time`))

								Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
								Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))

								Expect(fakeActor.GetServiceInstanceSummaryByNameAndSpaceCallCount()).To(Equal(1))
								serviceInstanceNameArg, spaceGUIDArg := fakeActor.GetServiceInstanceSummaryByNameAndSpaceArgsForCall(0)
								Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
								Expect(spaceGUIDArg).To(Equal("some-space-guid"))

								Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(0))
							})
						})

						When("the service instance is shared from another space", func() {
							BeforeEach(func() {
								returnedSummary.ServiceInstanceShareType = v2action.ServiceInstanceIsSharedFrom
								returnedSummary.ServiceInstanceSharedFrom = v2action.ServiceInstanceSharedFrom{
									SpaceGUID:        "some-space-guid",
									SpaceName:        "some-space-name",
									OrganizationName: "some-org-name",
								}
								fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
									returnedSummary,
									v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
									nil)
							})

							It("displays the shared from info and does not display the shared to info", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
								Expect(testUI.Out).To(Say(`name:\s+some-service-instance`))
								Expect(testUI.Out).To(Say(`shared from org/space:\s+some-org-name / some-space-name`))
								Expect(testUI.Out).To(Say(`service:\s+some-service`))
								Expect(testUI.Out).ToNot(Say("shared with spaces:"))

								Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
								Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))
							})
						})

						When("the service instance is shared to other spaces", func() {
							BeforeEach(func() {
								returnedSummary.ServiceInstanceShareType = v2action.ServiceInstanceIsSharedTo
								returnedSummary.ServiceInstanceSharedTos = []v2action.ServiceInstanceSharedTo{
									{
										SpaceGUID:        "another-space-guid",
										SpaceName:        "another-space-name",
										OrganizationName: "another-org-name",
										BoundAppCount:    2,
									},
									{
										SpaceGUID:        "yet-another-space-guid",
										SpaceName:        "yet-another-space-name",
										OrganizationName: "yet-another-org-name",
										BoundAppCount:    3,
									},
								}
								fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
									returnedSummary,
									v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
									nil)
							})

							When("the service instance is still shareable", func() {
								It("displays the shared to info and does not display the shared from info", func() {
									Expect(executeErr).ToNot(HaveOccurred())

									Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
									Expect(testUI.Out).ToNot(Say("shared from org/space:"))
									Expect(testUI.Out).To(Say(`dashboard:\s+some-dashboard`))
									Expect(testUI.Out).To(Say("shared with spaces:"))
									Expect(testUI.Out).To(Say(`org\s+space\s+bindings`))
									Expect(testUI.Out).To(Say(`another-org-name\s+another-space-name\s+2`))
									Expect(testUI.Out).To(Say(`yet-another-org-name\s+yet-another-space-name\s+3`))
									Expect(testUI.Out).To(Say(`Showing status of last operation from service some-service-instance\.\.\.`))

									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))
								})
							})

							When("the service instance is no longer shareable due to global settings only", func() {
								BeforeEach(func() {
									returnedSummary.ServiceInstanceSharingFeatureFlag = false
									returnedSummary.Service.Extra.Shareable = true
									fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
										returnedSummary,
										v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
										nil)
								})

								It("displays the shared to info and message that the service instance feature flag is disabled", func() {
									Expect(executeErr).ToNot(HaveOccurred())
									Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
									Expect(testUI.Out).ToNot(Say("shared from org/space:"))
									Expect(testUI.Out).To(Say(`dashboard:\s+some-dashboard`))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform.`))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say("shared with spaces:"))
									Expect(testUI.Out).To(Say(`org\s+space\s+bindings`))
									Expect(testUI.Out).To(Say(`another-org-name\s+another-space-name\s+2`))
									Expect(testUI.Out).To(Say(`yet-another-org-name\s+yet-another-space-name\s+3`))
									Expect(testUI.Out).To(Say(`Showing status of last operation from service some-service-instance\.\.\.`))

									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))
								})
							})

							When("the service instance is no longer shareable due to service broker settings only", func() {
								BeforeEach(func() {
									returnedSummary.ServiceInstanceSharingFeatureFlag = true
									returnedSummary.Service.Extra.Shareable = false
									fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
										returnedSummary,
										v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
										nil)
								})

								It("displays the shared to info and message that service instance sharing is disabled for the service", func() {
									Expect(executeErr).ToNot(HaveOccurred())
									Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
									Expect(testUI.Out).ToNot(Say("shared from org/space:"))
									Expect(testUI.Out).To(Say(`dashboard:\s+some-dashboard`))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say("Service instance sharing is disabled for this service."))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say("shared with spaces:"))
									Expect(testUI.Out).To(Say(`org\s+space\s+bindings`))
									Expect(testUI.Out).To(Say(`another-org-name\s+another-space-name\s+2`))
									Expect(testUI.Out).To(Say(`yet-another-org-name\s+yet-another-space-name\s+3`))
									Expect(testUI.Out).To(Say(`Showing status of last operation from service some-service-instance\.\.\.`))

									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))
								})
							})
							When("the service instance is no longer shareable due to  global settings AND service broker settings", func() {
								BeforeEach(func() {
									returnedSummary.ServiceInstanceSharingFeatureFlag = false
									returnedSummary.Service.Extra.Shareable = false
									fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
										returnedSummary,
										v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
										nil)
								})

								It("displays the shared to info, the message that the service instance feature flag is disabled and that service instance sharing is disabled for the service", func() {
									Expect(executeErr).ToNot(HaveOccurred())
									Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
									Expect(testUI.Out).ToNot(Say("shared from org/space:"))
									Expect(testUI.Out).To(Say(`dashboard:\s+some-dashboard`))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform. Also, service instance sharing is disabled for this service.`))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say("shared with spaces:"))
									Expect(testUI.Out).To(Say(`org\s+space\s+bindings`))
									Expect(testUI.Out).To(Say(`another-org-name\s+another-space-name\s+2`))
									Expect(testUI.Out).To(Say(`yet-another-org-name\s+yet-another-space-name\s+3`))
									Expect(testUI.Out).To(Say(`Showing status of last operation from service some-service-instance\.\.\.`))

									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))
								})
							})
						})

						When("the service instance has bound apps", func() {
							When("the service bindings have binding names", func() {
								BeforeEach(func() {
									returnedSummary.BoundApplications = []v2action.BoundApplication{
										{
											AppName:            "app-1",
											ServiceBindingName: "binding-name-1",
										},
										{
											AppName:            "app-2",
											ServiceBindingName: "binding-name-2",
										},
										{
											AppName:            "app-3",
											ServiceBindingName: "binding-name-3",
										},
									}

									fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
										returnedSummary,
										v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
										nil)
								})

								It("displays the bound apps table with service binding names", func() {
									Expect(executeErr).ToNot(HaveOccurred())

									Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say(`name:\s+some-service-instance`))
									Expect(testUI.Out).To(Say(`dashboard:\s+some-dashboard`))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say(`Showing status of last operation from service some-service-instance\.\.\.`))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say("bound apps:"))
									Expect(testUI.Out).To(Say(`name\s+binding name\s+status\s+message`))
									Expect(testUI.Out).To(Say(`app-1\s+binding-name-1`))
									Expect(testUI.Out).To(Say(`app-2\s+binding-name-2`))
									Expect(testUI.Out).To(Say(`app-3\s+binding-name-3`))

									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))

									Expect(fakeActor.GetServiceInstanceSummaryByNameAndSpaceCallCount()).To(Equal(1))
									serviceInstanceNameArg, spaceGUIDArg := fakeActor.GetServiceInstanceSummaryByNameAndSpaceArgsForCall(0)
									Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
									Expect(spaceGUIDArg).To(Equal("some-space-guid"))

									Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(0))
								})
							})

							When("the service bindings do not have binding names", func() {
								BeforeEach(func() {
									returnedSummary.BoundApplications = []v2action.BoundApplication{
										{AppName: "app-1"},
										{AppName: "app-2"},
										{AppName: "app-3"},
									}

									fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
										returnedSummary,
										v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
										nil)
								})

								It("displays the bound apps table with NO service binding names", func() {
									Expect(executeErr).ToNot(HaveOccurred())

									Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say(`name:\s+some-service-instance`))
									Expect(testUI.Out).To(Say(`dashboard:\s+some-dashboard`))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say(`Showing status of last operation from service some-service-instance\.\.\.`))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say("bound apps:"))
									Expect(testUI.Out).To(Say(`name\s+binding name\s+status\s+message`))
									Expect(testUI.Out).To(Say("app-1"))
									Expect(testUI.Out).To(Say("app-2"))
									Expect(testUI.Out).To(Say("app-3"))

									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))

									Expect(fakeActor.GetServiceInstanceSummaryByNameAndSpaceCallCount()).To(Equal(1))
									serviceInstanceNameArg, spaceGUIDArg := fakeActor.GetServiceInstanceSummaryByNameAndSpaceArgsForCall(0)
									Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
									Expect(spaceGUIDArg).To(Equal("some-space-guid"))

									Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(0))
								})
							})
						})

						When("the service instance does not have bound apps", func() {
							It("displays a message indicating that there are no bound apps", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).To(Say(`name:\s+some-service-instance`))
								Expect(testUI.Out).To(Say(`dashboard:\s+some-dashboard`))
								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).To(Say(`Showing status of last operation from service some-service-instance\.\.\.`))
								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).To(Say("There are no bound apps for this service."))

								Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
								Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))

								Expect(fakeActor.GetServiceInstanceSummaryByNameAndSpaceCallCount()).To(Equal(1))
								serviceInstanceNameArg, spaceGUIDArg := fakeActor.GetServiceInstanceSummaryByNameAndSpaceArgsForCall(0)
								Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
								Expect(spaceGUIDArg).To(Equal("some-space-guid"))

								Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(0))
							})

						})

						When("the service instance does not support an upgrade", func() {
							It("displays a message indicating that there are no upgrade available", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
								Expect(testUI.Out).To(Say(`name:\s+some-service-instance`))
								Expect(testUI.Out).To(Say("There are no bound apps for this service."))

								Expect(testUI.Out).To(Say("\n\n"), "Expect an empty line between 'bound apps' and 'upgrade available' messages")
								Expect(testUI.Out).To(Say("Upgrades are not supported by this broker."))

								Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
								Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))
							})
						})

						When("the service instance does not have an upgrade available", func() {
							BeforeEach(func() {
								maintenanceInfo := ccv2.MaintenanceInfo{
									Version:     "2.0.0",
									Description: "This is the maintenance description",
								}

								returnedSummary.ServicePlan.MaintenanceInfo = maintenanceInfo
								returnedSummary.ServiceInstance.MaintenanceInfo = maintenanceInfo
								fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
									returnedSummary,
									v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
									nil,
								)
							})

							It("displays a message indicating that there are no upgrade available", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
								Expect(testUI.Out).To(Say(`name:\s+some-service-instance`))
								Expect(testUI.Out).To(Say("There are no bound apps for this service."))

								Expect(testUI.Out).To(Say("\n\n"), "Expect an empty line between 'bound apps' and 'upgrade available' messages")
								Expect(testUI.Out).To(Say("There is no upgrade available for this service."))

								Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
								Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))
							})
						})

						When("the service instance has an upgrade available", func() {
							BeforeEach(func() {
								returnedSummary.ServicePlan.MaintenanceInfo = ccv2.MaintenanceInfo{
									Version:     "2.0.0",
									Description: "This is the maintenance description",
								}

								returnedSummary.ServiceInstance.MaintenanceInfo = ccv2.MaintenanceInfo{
									Version:     "1.0.0",
									Description: "This is old maintenance description",
								}

								fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
									returnedSummary,
									v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
									nil,
								)
							})

							It("displays a message with the upgrade available description", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
								Expect(testUI.Out).To(Say(`name:\s+some-service-instance`))
								Expect(testUI.Out).To(Say("There are no bound apps for this service."))

								Expect(testUI.Out).To(Say("\n\n"), "Expect an empty line between 'bound apps' and 'upgrade available' messages")
								Expect(testUI.Out).To(Say("Showing available upgrade details for this service..."))
								Expect(testUI.Out).To(Say("\n\n"), "Expect an empty line between messages")
								Expect(testUI.Out).To(Say("upgrade description: This is the maintenance description"))
								Expect(testUI.Out).To(Say("\n\n"), "Expect an empty line before tips")
								Expect(testUI.Out).To(Say("TIP: You can upgrade using 'cf update-service some-service-instance --upgrade'"))

								Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
								Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))
							})
						})

						When("the version of CC API is less than minimum version supporting maintenance_info", func() {
							BeforeEach(func() {
								fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionUpdateServiceInstanceMaintenanceInfoV2)
							})

							It("should not output anything about upgrades", func() {
								Consistently(testUI.Out).ShouldNot(Say("There is no upgrade available for this service."))
								Consistently(testUI.Out).ShouldNot(Say("Upgrades are not supported by this broker."))
								Consistently(testUI.Out).ShouldNot(Say("Showing available upgrade details for this service..."))
								Consistently(testUI.Out).ShouldNot(Say("upgrade description: This is the maintenance description"))
								Consistently(testUI.Out).ShouldNot(Say("TIP: You can upgrade using 'cf update-service some-service-instance --upgrade'"))
							})
						})

						When("the version of CC API is broken", func() {
							BeforeEach(func() {
								fakeActor.CloudControllerAPIVersionReturns("broken")
							})

							It("should fail", func() {
								Expect(executeErr).To(HaveOccurred())
							})
						})
					})

					When("the service instance is a user provided service instance", func() {
						BeforeEach(func() {
							fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
								v2action.ServiceInstanceSummary{
									ServiceInstance: v2action.ServiceInstance{
										Name: "some-service-instance",
										Type: constant.UserProvidedService,
									},
								},
								v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
								nil,
							)
						})

						It("displays a smaller service instance summary than for managed service instances", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
							Expect(testUI.Out).To(Say(""))
							Expect(testUI.Out).To(Say(`name:\s+some-service-instance`))
							Expect(testUI.Out).ToNot(Say("shared from"))
							Expect(testUI.Out).To(Say(`service:\s+user-provided`))
							Expect(testUI.Out).To(Say("tags:"))
							Expect(testUI.Out).ToNot(Say("plan:"))
							Expect(testUI.Out).ToNot(Say("description:"))
							Expect(testUI.Out).ToNot(Say("documentation:"))
							Expect(testUI.Out).ToNot(Say("dashboard:"))
							Expect(testUI.Out).ToNot(Say("shared with spaces"))
							Expect(testUI.Out).ToNot(Say("last operation"))
							Expect(testUI.Out).ToNot(Say("status:"))
							Expect(testUI.Out).ToNot(Say("message:"))
							Expect(testUI.Out).ToNot(Say("started:"))
							Expect(testUI.Out).ToNot(Say("updated:"))

							Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
							Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))

							Expect(fakeActor.GetServiceInstanceSummaryByNameAndSpaceCallCount()).To(Equal(1))
							serviceInstanceNameArg, spaceGUIDArg := fakeActor.GetServiceInstanceSummaryByNameAndSpaceArgsForCall(0)
							Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
							Expect(spaceGUIDArg).To(Equal("some-space-guid"))

							Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(0))
						})

						When("the service instance has bound apps", func() {
							When("the service bindings have binding names", func() {
								BeforeEach(func() {
									fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
										v2action.ServiceInstanceSummary{
											ServiceInstance: v2action.ServiceInstance{
												Name: "some-service-instance",
												Type: constant.UserProvidedService,
											},
											BoundApplications: []v2action.BoundApplication{
												{
													AppName:            "app-1",
													ServiceBindingName: "binding-name-1",
												},
												{
													AppName:            "app-2",
													ServiceBindingName: "binding-name-2",
												},
												{
													AppName:            "app-3",
													ServiceBindingName: "binding-name-3",
												},
											},
										},
										v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
										nil,
									)
								})

								It("displays the bound apps table with service binding names", func() {
									Expect(executeErr).ToNot(HaveOccurred())

									Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say(`name:\s+some-service-instance`))
									Expect(testUI.Out).To(Say(`service:\s+user-provided`))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say("bound apps:"))
									Expect(testUI.Out).To(Say(`name\s+binding name\s+status\s+message`))
									Expect(testUI.Out).To(Say(`app-1\s+binding-name-1`))
									Expect(testUI.Out).To(Say(`app-2\s+binding-name-2`))
									Expect(testUI.Out).To(Say(`app-3\s+binding-name-3`))

									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))

									Expect(fakeActor.GetServiceInstanceSummaryByNameAndSpaceCallCount()).To(Equal(1))
									serviceInstanceNameArg, spaceGUIDArg := fakeActor.GetServiceInstanceSummaryByNameAndSpaceArgsForCall(0)
									Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
									Expect(spaceGUIDArg).To(Equal("some-space-guid"))

									Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(0))
								})
							})

							When("the service bindings do not have binding names", func() {
								BeforeEach(func() {
									fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
										v2action.ServiceInstanceSummary{
											ServiceInstance: v2action.ServiceInstance{
												Name: "some-service-instance",
												Type: constant.UserProvidedService,
											},
											BoundApplications: []v2action.BoundApplication{
												{AppName: "app-1"},
												{AppName: "app-2"},
												{AppName: "app-3"},
											},
										},
										v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
										nil,
									)
								})

								It("displays the bound apps table with NO service binding names", func() {
									Expect(executeErr).ToNot(HaveOccurred())

									Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say(`name:\s+some-service-instance`))
									Expect(testUI.Out).To(Say(`service:\s+user-provided`))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say("bound apps:"))
									Expect(testUI.Out).To(Say(`name\s+binding name\s+status\s+message`))
									Expect(testUI.Out).To(Say("app-1"))
									Expect(testUI.Out).To(Say("app-2"))
									Expect(testUI.Out).To(Say("app-3"))

									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-1"))
									Expect(testUI.Err).To(Say("get-service-instance-summary-warning-2"))

									Expect(fakeActor.GetServiceInstanceSummaryByNameAndSpaceCallCount()).To(Equal(1))
									serviceInstanceNameArg, spaceGUIDArg := fakeActor.GetServiceInstanceSummaryByNameAndSpaceArgsForCall(0)
									Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
									Expect(spaceGUIDArg).To(Equal("some-space-guid"))

									Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(0))
								})

							})
						})

						When("the service instance does not have bound apps", func() {
							It("displays a message indicating that there are no bound apps", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say(`Showing info of service some-service-instance in org some-org / space some-space as some-user\.\.\.`))
								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).To(Say(`name:\s+some-service-instance`))
								Expect(testUI.Out).To(Say(`service:\s+user-provided`))
								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).To(Say("There are no bound apps for this service."))
							})
						})

						When("the service instance have tags", func() {
							BeforeEach(func() {
								fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
									v2action.ServiceInstanceSummary{
										ServiceInstance: v2action.ServiceInstance{
											Name: "some-service-instance",
											Type: constant.UserProvidedService,
											Tags: []string{"tag-1", "tag-2", "tag-3"},
										},
									},
									v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
									nil,
								)
							})

							It("displays the tags", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say(`tags:\s+tag-1, tag-2, tag-3`))
							})
						})

						When("the service instance has route service url", func() {
							BeforeEach(func() {
								fakeActor.GetServiceInstanceSummaryByNameAndSpaceReturns(
									v2action.ServiceInstanceSummary{
										ServiceInstance: v2action.ServiceInstance{
											Name:            "some-service-instance",
											Type:            constant.UserProvidedService,
											RouteServiceURL: "some-route-service-url",
										},
									},
									v2action.Warnings{"get-service-instance-summary-warning-1", "get-service-instance-summary-warning-2"},
									nil,
								)
							})

							It("displays the route service url", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say(`route service url:\s+some-route-service-url`))
							})
						})
					})
				})
			})
		})
	})
})
