package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v2action"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Summary Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
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

	Describe("GetApplicationSummaryByNameAndSpace", func() {
		var (
			appName              string
			spaceGUID            string
			withObfuscatedValues bool
			fakeRouteActor       *v7actionfakes.FakeRouteActor

			summary    ApplicationSummary
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			appName = "some-app-name"
			spaceGUID = "some-space-guid"
			withObfuscatedValues = true
			fakeRouteActor = new(v7actionfakes.FakeRouteActor)
		})

		Context("when a route actor is not provided", func() {
			JustBeforeEach(func() {
				summary, warnings, executeErr = actor.GetApplicationSummaryByNameAndSpace(appName, spaceGUID, withObfuscatedValues, nil)
			})

			When("retrieving the application is successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv3.Application{
							{
								Name:  "some-app-name",
								GUID:  "some-app-guid",
								State: constant.ApplicationStarted,
							},
						},
						ccv3.Warnings{"some-warning"},
						nil,
					)
				})

				When("retrieving the process information is successful", func() {
					BeforeEach(func() {
						listedProcesses := []ccv3.Process{
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
							ccv3.Warnings{"some-process-warning"},
							nil,
						)

						explicitlyCalledProcess := listedProcesses[0]
						explicitlyCalledProcess.Command = *types.NewFilteredString("some-start-command")
						fakeCloudControllerClient.GetApplicationProcessByTypeReturnsOnCall(
							0,
							explicitlyCalledProcess,
							ccv3.Warnings{"get-process-by-type-warning"},
							nil,
						)

						fakeCloudControllerClient.GetApplicationProcessByTypeReturnsOnCall(
							1,
							listedProcesses[1],
							ccv3.Warnings{"get-process-by-type-warning"},
							nil,
						)

						fakeCloudControllerClient.GetProcessInstancesReturns(
							[]ccv3.ProcessInstance{
								{
									State:       constant.ProcessInstanceRunning,
									CPU:         0.01,
									MemoryUsage: 1000000,
									DiskUsage:   2000000,
									MemoryQuota: 3000000,
									DiskQuota:   4000000,
									Index:       0,
								},
							},
							ccv3.Warnings{"some-process-stats-warning"},
							nil,
						)
					})

					When("app has droplet", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationDropletCurrentReturns(
								ccv3.Droplet{
									Stack: "some-stack",
									Buildpacks: []ccv3.DropletBuildpack{
										{
											Name: "some-buildpack",
										},
									},
									Image: "docker/some-image",
								},
								ccv3.Warnings{"some-droplet-warning"},
								nil,
							)
						})

						It("returns the summary and warnings with droplet information", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(summary).To(Equal(ApplicationSummary{
								Application: Application{
									Name:  "some-app-name",
									GUID:  "some-app-guid",
									State: constant.ApplicationStarted,
								},
								CurrentDroplet: Droplet{
									Stack: "some-stack",
									Image: "docker/some-image",
									Buildpacks: []DropletBuildpack{
										{
											Name: "some-buildpack",
										},
									},
								},
								ProcessSummaries: []ProcessSummary{
									{
										Process: Process{
											GUID:       "some-process-web-guid",
											Type:       "web",
											Command:    *types.NewFilteredString("[Redacted Value]"),
											MemoryInMB: types.NullUint64{Value: 64, IsSet: true},
										},
										InstanceDetails: []ProcessInstance{
											{
												State:       constant.ProcessInstanceRunning,
												CPU:         0.01,
												MemoryUsage: 1000000,
												DiskUsage:   2000000,
												MemoryQuota: 3000000,
												DiskQuota:   4000000,
												Index:       0,
											},
										},
									},
									{
										Process: Process{
											GUID:       "some-process-guid",
											MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
											Type:       "some-type",
											Command:    *types.NewFilteredString("some-start-command"),
										},
										InstanceDetails: []ProcessInstance{
											{
												State:       constant.ProcessInstanceRunning,
												CPU:         0.01,
												MemoryUsage: 1000000,
												DiskUsage:   2000000,
												MemoryQuota: 3000000,
												DiskQuota:   4000000,
												Index:       0,
											},
										},
									},
								},
							}))
							Expect(warnings).To(ConsistOf("some-warning", "some-process-warning", "get-process-by-type-warning", "get-process-by-type-warning", "some-process-stats-warning", "some-process-stats-warning", "some-droplet-warning"))

							Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
								ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-app-name"}},
								ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
							))

							Expect(fakeCloudControllerClient.GetApplicationDropletCurrentCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetApplicationDropletCurrentArgsForCall(0)).To(Equal("some-app-guid"))

							Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetApplicationProcessesArgsForCall(0)).To(Equal("some-app-guid"))

							Expect(fakeCloudControllerClient.GetApplicationProcessByTypeCallCount()).To(Equal(2))
							appGUID, processType := fakeCloudControllerClient.GetApplicationProcessByTypeArgsForCall(0)
							Expect(appGUID).To(Equal("some-app-guid"))
							Expect(processType).To(Equal("some-type"))

							Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(2))
							Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("some-process-guid"))
						})

						When("getting the current droplet returns an error", func() {
							var expectedErr error

							BeforeEach(func() {
								expectedErr = errors.New("some error")
								fakeCloudControllerClient.GetApplicationDropletCurrentReturns(
									ccv3.Droplet{},
									ccv3.Warnings{"some-droplet-warning"},
									expectedErr,
								)
							})

							It("returns the error", func() {
								Expect(executeErr).To(Equal(expectedErr))
								Expect(warnings).To(ConsistOf("some-warning", "some-process-warning", "get-process-by-type-warning", "get-process-by-type-warning", "some-process-stats-warning", "some-process-stats-warning", "some-droplet-warning"))
							})
						})
					})

					When("app does not have current droplet", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationDropletCurrentReturns(
								ccv3.Droplet{},
								ccv3.Warnings{"some-droplet-warning"},
								ccerror.DropletNotFoundError{},
							)
						})

						It("returns the summary and warnings without droplet information", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(summary).To(Equal(ApplicationSummary{
								Application: Application{
									Name:  "some-app-name",
									GUID:  "some-app-guid",
									State: constant.ApplicationStarted,
								},
								ProcessSummaries: []ProcessSummary{
									{
										Process: Process{
											GUID:       "some-process-web-guid",
											Type:       "web",
											Command:    *types.NewFilteredString("[Redacted Value]"),
											MemoryInMB: types.NullUint64{Value: 64, IsSet: true},
										},
										InstanceDetails: []ProcessInstance{
											{
												State:       constant.ProcessInstanceRunning,
												CPU:         0.01,
												MemoryUsage: 1000000,
												DiskUsage:   2000000,
												MemoryQuota: 3000000,
												DiskQuota:   4000000,
												Index:       0,
											},
										},
									},
									{
										Process: Process{
											GUID:       "some-process-guid",
											MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
											Type:       "some-type",
											Command:    *types.NewFilteredString("some-start-command"),
										},
										InstanceDetails: []ProcessInstance{
											{
												State:       constant.ProcessInstanceRunning,
												CPU:         0.01,
												MemoryUsage: 1000000,
												DiskUsage:   2000000,
												MemoryQuota: 3000000,
												DiskQuota:   4000000,
												Index:       0,
											},
										},
									},
								},
							}))
							Expect(warnings).To(ConsistOf("some-warning", "some-process-warning", "get-process-by-type-warning", "get-process-by-type-warning", "some-process-stats-warning", "some-process-stats-warning", "some-droplet-warning"))
						})
					})

				})

				When("getting the app process instances returns an error", func() {
					var expectedErr error

					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationProcessesReturns(
							[]ccv3.Process{
								{
									GUID: "some-process-guid",
									Type: "some-type",
								},
							},
							ccv3.Warnings{"some-process-warning"},
							nil,
						)

						fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
							ccv3.Process{},
							ccv3.Warnings{"get-process-by-type-warning"},
							nil,
						)

						expectedErr = errors.New("some error")
						fakeCloudControllerClient.GetProcessInstancesReturns(
							[]ccv3.ProcessInstance{},
							ccv3.Warnings{"some-process-stats-warning"},
							expectedErr,
						)
					})

					It("returns the error", func() {
						Expect(executeErr).To(Equal(expectedErr))
						Expect(warnings).To(ConsistOf("some-warning", "some-process-warning", "get-process-by-type-warning", "some-process-stats-warning"))
					})
				})
			})

			When("getting the app processes returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv3.Application{
							{
								Name:  "some-app-name",
								GUID:  "some-app-guid",
								State: constant.ApplicationStarted,
							},
						},
						ccv3.Warnings{"some-warning"},
						nil,
					)

					expectedErr = errors.New("some error")
					fakeCloudControllerClient.GetApplicationProcessesReturns(
						[]ccv3.Process{{Type: constant.ProcessTypeWeb}},
						ccv3.Warnings{"some-process-warning"},
						expectedErr,
					)
				})

				It("returns the error", func() {
					Expect(executeErr).To(Equal(expectedErr))
					Expect(warnings).To(ConsistOf("some-warning", "some-process-warning"))
				})
			})
		})

		When("passed with a routeActor", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{
							Name:  "some-app-name",
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			JustBeforeEach(func() {
				summary, warnings, executeErr = actor.GetApplicationSummaryByNameAndSpace(appName, spaceGUID, withObfuscatedValues, fakeRouteActor)
			})

			When("getting the routes is successful", func() {
				BeforeEach(func() {
					routes := v2action.Routes{
						{GUID: "some-route-guid"},
						{GUID: "some-other-route-guid"},
					}

					fakeRouteActor.GetApplicationRoutesReturns(routes, v2action.Warnings{"v2-routes-warnings"}, nil)
				})

				It("retrieves and sets the application routes", func() {
					Expect(warnings).To(ConsistOf("some-warning", "v2-routes-warnings"))
					Expect(summary.Routes).To(ConsistOf(
						v2action.Route{GUID: "some-route-guid"},
						v2action.Route{GUID: "some-other-route-guid"},
					))

					Expect(fakeRouteActor.GetApplicationRoutesCallCount()).To(Equal(1))
					Expect(fakeRouteActor.GetApplicationRoutesArgsForCall(0)).To(Equal("some-app-guid"))
				})
			})

			When("getting the application routes errors", func() {
				When("a generic error is returned", func() {
					BeforeEach(func() {
						fakeRouteActor.GetApplicationRoutesReturns(nil, v2action.Warnings{"v2-routes-warnings"}, errors.New("some-error"))
					})

					It("returns warnings and the error", func() {
						Expect(executeErr).To(MatchError("some-error"))
						Expect(warnings).To(ConsistOf("some-warning", "v2-routes-warnings"))
					})
				})

				When("a ResourceNotFoundError is returned", func() {
					BeforeEach(func() {
						fakeRouteActor.GetApplicationRoutesReturns(nil, v2action.Warnings{"v2-routes-warnings"}, ccerror.ResourceNotFoundError{})
					})

					It("adds warnings and continues", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("some-warning", "v2-routes-warnings"))
					})
				})
			})
		})
	})
})
