package v7action_test

import (
	"errors"
	"fmt"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/clock"

	"code.cloudfoundry.org/cli/actor/actionerror"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Summary Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, clock.NewClock())
	})

	Describe("ApplicationSummary", func() {
		DescribeTable("GetIsolationSegmentName",
			func(summary ApplicationSummary, isoName string, exists bool) {
				name, ok := summary.GetIsolationSegmentName()
				Expect(ok).To(Equal(exists))
				Expect(name).To(Equal(isoName))
			},

			Entry("when the there are application instances and the isolationSegmentName is set",
				ApplicationSummary{
					ProcessSummaries: ProcessSummaries{
						ProcessSummary{
							InstanceDetails: []ProcessInstance{
								{IsolationSegment: "some-name"},
							},
						},
					},
				},
				"some-name",
				true,
			),

			Entry("when the there are application instances and the isolationSegmentName is blank",
				ApplicationSummary{
					ProcessSummaries: ProcessSummaries{
						ProcessSummary{InstanceDetails: []ProcessInstance{{}}},
					},
				},
				"",
				false,
			),

			Entry("when the there are no application instances", ApplicationSummary{ProcessSummaries: ProcessSummaries{{}}}, "", false),
			Entry("when the there are no processes", ApplicationSummary{}, "", false),
		)
	})

	Describe("GetAppSummariesForSpace", func() {
		var (
			spaceGUID     string
			labelSelector string
			omitStats     bool

			summaries  []ApplicationSummary
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			spaceGUID = "some-space-guid"
			labelSelector = "some-key=some-value"
			omitStats = false
		})

		JustBeforeEach(func() {
			summaries, warnings, executeErr = actor.GetAppSummariesForSpace(spaceGUID, labelSelector, omitStats)
		})

		When("getting the application is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{
							Name:  "some-app-name",
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
					},
					ccv3.Warnings{"get-apps-warning"},
					nil,
				)

				listedProcesses := []resources.Process{
					{
						GUID:       "some-process-guid",
						Type:       "some-type",
						Command:    *types.NewFilteredString("[Redacted Value]"),
						MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
						AppGUID:    "some-app-guid",
					},
					{
						GUID:       "some-process-web-guid",
						Type:       "web",
						Command:    *types.NewFilteredString("[Redacted Value]"),
						MemoryInMB: types.NullUint64{Value: 64, IsSet: true},
						AppGUID:    "some-app-guid",
					},
				}

				fakeCloudControllerClient.GetProcessesReturns(
					listedProcesses,
					ccv3.Warnings{"get-app-processes-warning"},
					nil,
				)

				explicitlyCalledProcess := listedProcesses[0]
				explicitlyCalledProcess.Command = *types.NewFilteredString("some-start-command")
				fakeCloudControllerClient.GetProcessReturnsOnCall(
					0,
					explicitlyCalledProcess,
					ccv3.Warnings{"get-process-by-type-warning"},
					nil,
				)

				fakeCloudControllerClient.GetProcessReturnsOnCall(
					1,
					listedProcesses[1],
					ccv3.Warnings{"get-process-by-type-warning"},
					nil,
				)

				fakeCloudControllerClient.GetProcessInstancesReturns(
					[]ccv3.ProcessInstance{
						{
							State:          constant.ProcessInstanceRunning,
							CPUEntitlement: types.NullFloat64{Value: 0.01, IsSet: true},
							MemoryUsage:    1000000,
							DiskUsage:      2000000,
							MemoryQuota:    3000000,
							DiskQuota:      4000000,
							Index:          0,
						},
					},
					ccv3.Warnings{"get-process-instances-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{
						{
							GUID: "some-route-guid",
							Destinations: []resources.RouteDestination{
								{
									App: resources.RouteDestinationApp{
										GUID: "some-app-guid",
									},
								},
							},
						},
						{
							GUID: "some-other-route-guid",
							Destinations: []resources.RouteDestination{
								{
									App: resources.RouteDestinationApp{
										GUID: "some-app-guid",
									},
								},
							},
						},
					},
					ccv3.Warnings{"get-routes-warning"},
					nil,
				)
			})

			It("returns the summary and warnings with droplet information", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(summaries).To(Equal([]ApplicationSummary{
					{
						Application: resources.Application{
							Name:  "some-app-name",
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
						ProcessSummaries: []ProcessSummary{
							{
								Process: resources.Process{
									GUID:       "some-process-web-guid",
									Type:       "web",
									Command:    *types.NewFilteredString("[Redacted Value]"),
									MemoryInMB: types.NullUint64{Value: 64, IsSet: true},
									AppGUID:    "some-app-guid",
								},
								InstanceDetails: []ProcessInstance{
									{
										State:          constant.ProcessInstanceRunning,
										CPUEntitlement: types.NullFloat64{Value: 0.01, IsSet: true},
										MemoryUsage:    1000000,
										DiskUsage:      2000000,
										MemoryQuota:    3000000,
										DiskQuota:      4000000,
										Index:          0,
									},
								},
							},
							{
								Process: resources.Process{
									GUID:       "some-process-guid",
									MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
									Type:       "some-type",
									Command:    *types.NewFilteredString("[Redacted Value]"),
									AppGUID:    "some-app-guid",
								},
								InstanceDetails: []ProcessInstance{
									{
										State:          constant.ProcessInstanceRunning,
										CPUEntitlement: types.NullFloat64{Value: 0.01, IsSet: true},
										MemoryUsage:    1000000,
										DiskUsage:      2000000,
										MemoryQuota:    3000000,
										DiskQuota:      4000000,
										Index:          0,
									},
								},
							},
						},
						Routes: []resources.Route{
							{
								GUID: "some-route-guid",
								Destinations: []resources.RouteDestination{
									{
										App: resources.RouteDestinationApp{
											GUID: "some-app-guid",
										},
									},
								},
							},
							{
								GUID: "some-other-route-guid",
								Destinations: []resources.RouteDestination{
									{
										App: resources.RouteDestinationApp{
											GUID: "some-app-guid",
										},
									},
								},
							},
						},
					},
				}))

				Expect(warnings).To(ConsistOf(
					"get-apps-warning",
					"get-app-processes-warning",
					"get-process-instances-warning",
					"get-process-instances-warning",
					"get-routes-warning",
				))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.OrderBy, Values: []string{"name"}},
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
					ccv3.Query{Key: ccv3.LabelSelectorFilter, Values: []string{"some-key=some-value"}},
					ccv3.Query{Key: ccv3.PerPage, Values: []string{ccv3.MaxPerPage}},
				))

				Expect(fakeCloudControllerClient.GetProcessesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetProcessesArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{"some-app-guid"}},
				))

				Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(2))
				Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("some-process-guid"))
			})

			When("there is no label selector", func() {
				BeforeEach(func() {
					labelSelector = ""
				})
				It("doesn't pass a label selection filter", func() {
					Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.OrderBy, Values: []string{"name"}},
						ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
						ccv3.Query{Key: ccv3.PerPage, Values: []string{ccv3.MaxPerPage}},
					))
				})
			})
			When("there is long list of app guids", func() {
				BeforeEach(func() {
					lotsOfApps := []resources.Application{}
					for i := 0; i < 600; i++ {
						lotsOfApps = append(lotsOfApps, resources.Application{
							Name:  "some-app-name",
							GUID:  fmt.Sprintf("some-app-guid-%d", i),
							State: constant.ApplicationStarted,
						})
					}
					fakeCloudControllerClient.GetApplicationsReturns(
						lotsOfApps,
						ccv3.Warnings{"get-apps-warning"},
						nil,
					)
				})

				It("batches the calls to the API for processes", func() {
					Expect(fakeCloudControllerClient.GetProcessesCallCount()).To(Equal(3))

					getProcessesArgs := fakeCloudControllerClient.GetProcessesArgsForCall(0)
					Expect(len(getProcessesArgs[0].Values)).To(Equal(200))
					getProcessesArgs = fakeCloudControllerClient.GetProcessesArgsForCall(1)
					Expect(len(getProcessesArgs[0].Values)).To(Equal(200))
					getProcessesArgs = fakeCloudControllerClient.GetProcessesArgsForCall(2)
					Expect(len(getProcessesArgs[0].Values)).To(Equal(200))

					Expect(len(summaries)).To(Equal(600))
				})
				It("batches the calls to the API for routes", func() {
					Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(3))

					getArgs := fakeCloudControllerClient.GetRoutesArgsForCall(0)
					Expect(len(getArgs[0].Values)).To(Equal(200))
					getArgs = fakeCloudControllerClient.GetRoutesArgsForCall(1)
					Expect(len(getArgs[0].Values)).To(Equal(200))
					getArgs = fakeCloudControllerClient.GetRoutesArgsForCall(2)
					Expect(len(getArgs[0].Values)).To(Equal(200))

					Expect(len(summaries)).To(Equal(600))
				})
			})
		})

		When("getting the application fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{
							Name:  "some-app-name",
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
					},
					ccv3.Warnings{"get-apps-warning"},
					errors.New("failed to get app"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError("failed to get app"))
				Expect(warnings).To(ConsistOf("get-apps-warning"))
			})
		})

		When("omitStats flag is provided", func() {
			BeforeEach(func() {
				omitStats = true

				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{
							Name:  "some-app-name",
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
					},
					ccv3.Warnings{"get-apps-warning"},
					nil,
				)

				listedProcesses := []resources.Process{
					{
						GUID:       "some-process-web-guid",
						Type:       "web",
						Command:    *types.NewFilteredString("[Redacted Value]"),
						MemoryInMB: types.NullUint64{Value: 64, IsSet: true},
						AppGUID:    "some-app-guid",
					},
				}

				fakeCloudControllerClient.GetProcessesReturns(
					listedProcesses,
					ccv3.Warnings{"get-app-processes-warning"},
					nil,
				)
			})
			It("doesn't call the stats endpoint", func() {
				Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(0))
			})
		})

		When("an application is deleted in between", func() {

			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{
							Name:  "some-app-name",
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
					},
					nil,
					nil,
				)

				fakeCloudControllerClient.GetProcessesReturns(
					[]resources.Process{
						{
							GUID:       "some-process-web-guid",
							Type:       "web",
							Command:    *types.NewFilteredString("[Redacted Value]"),
							MemoryInMB: types.NullUint64{Value: 64, IsSet: true},
							AppGUID:    "some-app-guid",
						},
					},
					nil,
					nil,
				)

				fakeCloudControllerClient.GetProcessInstancesReturns(nil, nil, ccerror.ProcessNotFoundError{})
			})

			It("does not fail and has empty ProcessSummaries & Routes", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(summaries).To(Equal([]ApplicationSummary{
					{
						Application: resources.Application{
							Name:  "some-app-name",
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
						ProcessSummaries: nil,
						Routes:           nil,
					},
				}))

			})
		})
	})

	Describe("GetDetailedAppSummary", func() {
		var (
			appName              string
			spaceGUID            string
			withObfuscatedValues bool

			summary    DetailedApplicationSummary
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			appName = "some-app-name"
			spaceGUID = "some-space-guid"
			withObfuscatedValues = true
		})

		JustBeforeEach(func() {
			summary, warnings, executeErr = actor.GetDetailedAppSummary(appName, spaceGUID, withObfuscatedValues)
		})

		When("getting the application is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{
							Name:  "some-app-name",
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
					},
					ccv3.Warnings{"get-apps-warning"},
					nil,
				)
			})

			When("getting the process information is successful", func() {
				BeforeEach(func() {
					listedProcesses := []resources.Process{
						{
							GUID:       "some-process-guid",
							Type:       "some-type",
							Command:    *types.NewFilteredString("[Redacted Value]"),
							MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
						},
						{
							GUID:       "some-process-web-guid",
							Type:       "web",
							Command:    *types.NewFilteredString("[Redacted Value]"),
							MemoryInMB: types.NullUint64{Value: 64, IsSet: true},
						},
					}
					fakeCloudControllerClient.GetApplicationProcessesReturns(
						listedProcesses,
						ccv3.Warnings{"get-app-processes-warning"},
						nil,
					)

					explicitlyCalledProcess := listedProcesses[0]
					explicitlyCalledProcess.Command = *types.NewFilteredString("some-start-command")
					fakeCloudControllerClient.GetProcessReturnsOnCall(
						0,
						explicitlyCalledProcess,
						ccv3.Warnings{"get-process-by-type-warning"},
						nil,
					)

					fakeCloudControllerClient.GetProcessReturnsOnCall(
						1,
						listedProcesses[1],
						ccv3.Warnings{"get-process-by-type-warning"},
						nil,
					)

					fakeCloudControllerClient.GetProcessSidecarsReturns(
						[]resources.Sidecar{
							{
								GUID:    "sidecar-guid",
								Name:    "sidecar_name",
								Command: *types.NewFilteredString("my-sidecar-command"),
							},
						},
						ccv3.Warnings{"get-process-sidecars-warning"},
						nil,
					)

					fakeCloudControllerClient.GetProcessInstancesReturns(
						[]ccv3.ProcessInstance{
							{
								State:          constant.ProcessInstanceRunning,
								CPUEntitlement: types.NullFloat64{Value: 0.01, IsSet: true},
								MemoryUsage:    1000000,
								DiskUsage:      2000000,
								MemoryQuota:    3000000,
								DiskQuota:      4000000,
								Index:          0,
							},
						},
						ccv3.Warnings{"get-process-instances-warning"},
						nil,
					)
				})

				When("getting current droplet succeeds", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationDropletCurrentReturns(
							resources.Droplet{
								Stack: "some-stack",
								Buildpacks: []resources.DropletBuildpack{
									{
										Name: "some-buildpack",
									},
								},
								Image: "docker/some-image",
							},
							ccv3.Warnings{"get-app-droplet-warning"},
							nil,
						)
					})

					When("getting application deployment succeeds", func() {
						When("the deployment is active", func() {
							When("the deployment strategy is rolling", func() {
								When("the deployment is in progress", func() {
									BeforeEach(func() {
										fakeCloudControllerClient.GetDeploymentsReturns(
											[]resources.Deployment{
												{
													GUID:         "some-deployment-guid",
													Strategy:     "rolling",
													StatusValue:  "ACTIVE",
													StatusReason: "DEPLOYING",
												},
											},
											nil,
											nil,
										)
									})
									It("returns the deployment information", func() {
										Expect(summary.Deployment).To(Equal(resources.Deployment{
											GUID:         "some-deployment-guid",
											Strategy:     "rolling",
											StatusValue:  "ACTIVE",
											StatusReason: "DEPLOYING",
										}))
									})
								})

								When("the deployment is canceled", func() {
									When("the deployment is in progress", func() {
										BeforeEach(func() {
											fakeCloudControllerClient.GetDeploymentsReturns(
												[]resources.Deployment{
													{
														GUID:         "some-deployment-guid",
														Strategy:     "rolling",
														StatusValue:  "ACTIVE",
														StatusReason: "CANCELLING",
													},
												},
												nil,
												nil,
											)
										})
										It("returns the deployment information", func() {
											Expect(summary.Deployment).To(Equal(resources.Deployment{
												GUID:         "some-deployment-guid",
												Strategy:     "rolling",
												StatusValue:  "ACTIVE",
												StatusReason: "CANCELLING",
											}))
										})
									})
								})
							})

							When("the deployment strategy is canary", func() {
								When("the deployment is paused", func() {
									BeforeEach(func() {
										fakeCloudControllerClient.GetDeploymentsReturns(
											[]resources.Deployment{
												{
													GUID:         "some-deployment-guid",
													Strategy:     "canary",
													StatusValue:  "ACTIVE",
													StatusReason: "PAUSED",
												},
											},
											nil,
											nil,
										)
									})
									It("returns the deployment information", func() {
										Expect(summary.Deployment).To(Equal(resources.Deployment{
											GUID:         "some-deployment-guid",
											Strategy:     "canary",
											StatusValue:  "ACTIVE",
											StatusReason: "PAUSED",
										}))
									})
								})
							})
						})

						When("the deployment is not active", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.GetDeploymentsReturns(
									[]resources.Deployment{
										{
											GUID:         "",
											Strategy:     "",
											StatusValue:  "",
											StatusReason: "",
										},
									},
									nil,
									nil,
								)
							})
							It("returns no deployment information", func() {
								Expect(summary.Deployment).To(Equal(resources.Deployment{
									GUID:         "",
									Strategy:     "",
									StatusValue:  "",
									StatusReason: "",
								}))
							})
						})
					})

					When("getting application deployment fails", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetDeploymentsReturns(
								nil,
								ccv3.Warnings{"get-deployments-warning"},
								errors.New("some-error"),
							)
						})

						It("returns the warnings and error", func() {
							Expect(executeErr).To(MatchError("some-error"))
							Expect(warnings).To(ConsistOf(
								"get-apps-warning",
								"get-app-processes-warning",
								"get-process-by-type-warning",
								"get-process-sidecars-warning",
								"get-process-instances-warning",
								"get-process-by-type-warning",
								"get-process-sidecars-warning",
								"get-process-instances-warning",
								"get-app-droplet-warning",
								"get-deployments-warning",
							))
						})
					})

					When("getting application routes succeeds", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationRoutesReturns(
								[]resources.Route{
									{GUID: "some-route-guid"},
									{GUID: "some-other-route-guid"},
								},
								ccv3.Warnings{"get-routes-warning"},
								nil,
							)
						})

						It("returns the summary and warnings with droplet information", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(summary).To(Equal(DetailedApplicationSummary{
								ApplicationSummary: ApplicationSummary{
									Application: resources.Application{
										Name:  "some-app-name",
										GUID:  "some-app-guid",
										State: constant.ApplicationStarted,
									},
									ProcessSummaries: []ProcessSummary{
										{
											Process: resources.Process{
												GUID:       "some-process-web-guid",
												Type:       "web",
												Command:    *types.NewFilteredString("[Redacted Value]"),
												MemoryInMB: types.NullUint64{Value: 64, IsSet: true},
											},
											Sidecars: []resources.Sidecar{
												{
													GUID:    "sidecar-guid",
													Name:    "sidecar_name",
													Command: *types.NewFilteredString("my-sidecar-command"),
												},
											},
											InstanceDetails: []ProcessInstance{
												{
													State:          constant.ProcessInstanceRunning,
													CPUEntitlement: types.NullFloat64{Value: 0.01, IsSet: true},
													MemoryUsage:    1000000,
													DiskUsage:      2000000,
													MemoryQuota:    3000000,
													DiskQuota:      4000000,
													Index:          0,
												},
											},
										},
										{
											Process: resources.Process{
												GUID:       "some-process-guid",
												MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
												Type:       "some-type",
												Command:    *types.NewFilteredString("some-start-command"),
											},
											Sidecars: []resources.Sidecar{
												{
													GUID:    "sidecar-guid",
													Name:    "sidecar_name",
													Command: *types.NewFilteredString("my-sidecar-command"),
												},
											},
											InstanceDetails: []ProcessInstance{
												{
													State:          constant.ProcessInstanceRunning,
													CPUEntitlement: types.NullFloat64{Value: 0.01, IsSet: true},
													MemoryUsage:    1000000,
													DiskUsage:      2000000,
													MemoryQuota:    3000000,
													DiskQuota:      4000000,
													Index:          0,
												},
											},
										},
									},
									Routes: []resources.Route{
										{GUID: "some-route-guid"},
										{GUID: "some-other-route-guid"},
									},
								},
								CurrentDroplet: resources.Droplet{
									Stack: "some-stack",
									Image: "docker/some-image",
									Buildpacks: []resources.DropletBuildpack{
										{
											Name: "some-buildpack",
										},
									},
								},
							}))

							Expect(warnings).To(ConsistOf(
								"get-apps-warning",
								"get-app-processes-warning",
								"get-process-by-type-warning",
								"get-process-by-type-warning",
								"get-process-instances-warning",
								"get-process-instances-warning",
								"get-process-sidecars-warning",
								"get-process-sidecars-warning",
								"get-app-droplet-warning",
								"get-routes-warning",
							))

							Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
								ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-app-name"}},
								ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
							))

							Expect(fakeCloudControllerClient.GetApplicationDropletCurrentCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetApplicationDropletCurrentArgsForCall(0)).To(Equal("some-app-guid"))

							Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetApplicationProcessesArgsForCall(0)).To(Equal("some-app-guid"))

							Expect(fakeCloudControllerClient.GetProcessCallCount()).To(Equal(2))
							processGUID := fakeCloudControllerClient.GetProcessArgsForCall(0)
							Expect(processGUID).To(Equal("some-process-guid"))

							Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(2))
							Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("some-process-guid"))
						})
					})

					When("getting application routes fails", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationRoutesReturns(nil, ccv3.Warnings{"get-routes-warnings"}, errors.New("some-error"))
						})

						It("returns the warnings and error", func() {
							Expect(executeErr).To(MatchError("some-error"))
							Expect(warnings).To(ConsistOf(
								"get-apps-warning",
								"get-app-processes-warning",
								"get-process-by-type-warning",
								"get-process-by-type-warning",
								"get-process-instances-warning",
								"get-process-instances-warning",
								"get-process-sidecars-warning",
								"get-process-sidecars-warning",
								"get-routes-warnings",
							))
						})
					})
				})

				When("app does not have current droplet", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationDropletCurrentReturns(
							resources.Droplet{},
							ccv3.Warnings{"get-app-droplet-warning"},
							ccerror.DropletNotFoundError{},
						)
					})

					It("returns the summary and warnings without droplet information", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(summary).To(Equal(DetailedApplicationSummary{
							ApplicationSummary: ApplicationSummary{
								Application: resources.Application{
									Name:  "some-app-name",
									GUID:  "some-app-guid",
									State: constant.ApplicationStarted,
								},
								ProcessSummaries: []ProcessSummary{
									{
										Process: resources.Process{
											GUID:       "some-process-web-guid",
											Type:       "web",
											Command:    *types.NewFilteredString("[Redacted Value]"),
											MemoryInMB: types.NullUint64{Value: 64, IsSet: true},
										},
										Sidecars: []resources.Sidecar{
											{
												GUID:    "sidecar-guid",
												Name:    "sidecar_name",
												Command: *types.NewFilteredString("my-sidecar-command"),
											},
										},
										InstanceDetails: []ProcessInstance{
											{
												State:          constant.ProcessInstanceRunning,
												CPUEntitlement: types.NullFloat64{Value: 0.01, IsSet: true},
												MemoryUsage:    1000000,
												DiskUsage:      2000000,
												MemoryQuota:    3000000,
												DiskQuota:      4000000,
												Index:          0,
											},
										},
									},
									{
										Process: resources.Process{
											GUID:       "some-process-guid",
											MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
											Type:       "some-type",
											Command:    *types.NewFilteredString("some-start-command"),
										},
										Sidecars: []resources.Sidecar{
											{
												GUID:    "sidecar-guid",
												Name:    "sidecar_name",
												Command: *types.NewFilteredString("my-sidecar-command"),
											},
										},
										InstanceDetails: []ProcessInstance{
											{
												State:          constant.ProcessInstanceRunning,
												CPUEntitlement: types.NullFloat64{Value: 0.01, IsSet: true},
												MemoryUsage:    1000000,
												DiskUsage:      2000000,
												MemoryQuota:    3000000,
												DiskQuota:      4000000,
												Index:          0,
											},
										},
									},
								},
							},
						}))

						Expect(warnings).To(ConsistOf(
							"get-apps-warning",
							"get-app-processes-warning",
							"get-process-by-type-warning",
							"get-process-by-type-warning",
							"get-process-instances-warning",
							"get-process-instances-warning",
							"get-process-sidecars-warning",
							"get-process-sidecars-warning",
							"get-app-droplet-warning",
						))
					})
				})

				When("getting the current droplet returns an error", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("some error")
						fakeCloudControllerClient.GetApplicationDropletCurrentReturns(
							resources.Droplet{},
							ccv3.Warnings{"get-droplet-warning"},
							expectedErr,
						)
					})

					It("returns the error", func() {
						Expect(executeErr).To(Equal(expectedErr))
						Expect(warnings).To(ConsistOf(
							"get-apps-warning",
							"get-app-processes-warning",
							"get-process-by-type-warning",
							"get-process-by-type-warning",
							"get-process-instances-warning",
							"get-process-instances-warning",
							"get-process-sidecars-warning",
							"get-process-sidecars-warning",
							"get-droplet-warning",
						))
					})
				})
			})

			When("getting the app processes returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationProcessesReturns(
						[]resources.Process{
							{
								GUID: "some-process-guid",
								Type: "some-type",
							},
						},
						ccv3.Warnings{"get-app-processes-warning"},
						nil,
					)

					fakeCloudControllerClient.GetProcessReturns(
						resources.Process{},
						ccv3.Warnings{"get-process-warning"},
						nil,
					)

					expectedErr = errors.New("some error")
					fakeCloudControllerClient.GetProcessInstancesReturns(
						[]ccv3.ProcessInstance{},
						ccv3.Warnings{"get-process-instances-warning"},
						expectedErr,
					)
				})

				It("returns the error and warnings", func() {
					Expect(executeErr).To(Equal(expectedErr))
					Expect(warnings).To(ConsistOf("get-apps-warning", "get-app-processes-warning", "get-process-warning", "get-process-instances-warning"))
				})
			})
		})

		When("no applications are returned", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{},
					ccv3.Warnings{"get-apps-warning"},
					nil,
				)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError(actionerror.ApplicationNotFoundError{Name: appName}))
				Expect(warnings).To(ConsistOf("get-apps-warning"))
			})
		})

		When("getting the application fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{
							Name:  "some-app-name",
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
					},
					ccv3.Warnings{"get-apps-warning"},
					errors.New("failed to get app"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError("failed to get app"))
				Expect(warnings).To(ConsistOf("get-apps-warning"))
			})
		})
	})
})
