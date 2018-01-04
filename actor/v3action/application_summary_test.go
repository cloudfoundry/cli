package v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Summary Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("GetApplicationSummaryByNameAndSpace", func() {
		Context("when the app exists", func() {
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

				fakeCloudControllerClient.GetApplicationProcessesReturns(
					[]ccv3.Process{
						{
							GUID:       "some-process-guid",
							Type:       "some-type",
							MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
						},
					},
					ccv3.Warnings{"some-process-warning"},
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

			Context("when app has droplet", func() {
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
					summary, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app-name", "some-space-guid")
					Expect(err).ToNot(HaveOccurred())
					Expect(summary).To(Equal(ApplicationSummary{
						Application: Application{
							Name:  "some-app-name",
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
						CurrentDroplet: Droplet{
							Stack: "some-stack",
							Image: "docker/some-image",
							Buildpacks: []Buildpack{
								{
									Name: "some-buildpack",
								},
							},
						},
						ProcessSummaries: []ProcessSummary{
							{
								Process: Process{
									GUID:       "some-process-guid",
									MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
									Type:       "some-type",
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
					Expect(warnings).To(Equal(Warnings{"some-warning", "some-process-warning", "some-process-stats-warning", "some-droplet-warning"}))

					Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-app-name"}},
						ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
					))

					Expect(fakeCloudControllerClient.GetApplicationDropletCurrentCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationDropletCurrentArgsForCall(0)).To(Equal("some-app-guid"))

					Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationProcessesArgsForCall(0)).To(Equal("some-app-guid"))

					Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("some-process-guid"))
				})

				Context("when getting the current droplet returns an error", func() {
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
						_, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app-name", "some-space-guid")
						Expect(err).To(Equal(expectedErr))
						Expect(warnings).To(Equal(Warnings{"some-warning", "some-process-warning", "some-process-stats-warning", "some-droplet-warning"}))
					})
				})
			})

			Context("when app does not have current droplet", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationDropletCurrentReturns(
						ccv3.Droplet{},
						ccv3.Warnings{"some-droplet-warning"},
						ccerror.DropletNotFoundError{},
					)
				})

				It("returns the summary and warnings without droplet information", func() {
					summary, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app-name", "some-space-guid")
					Expect(err).ToNot(HaveOccurred())
					Expect(summary).To(Equal(ApplicationSummary{
						Application: Application{
							Name:  "some-app-name",
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
						ProcessSummaries: []ProcessSummary{
							{
								Process: Process{
									GUID:       "some-process-guid",
									MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
									Type:       "some-type",
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
					Expect(warnings).To(Equal(Warnings{"some-warning", "some-process-warning", "some-process-stats-warning", "some-droplet-warning"}))

					Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-app-name"}},
						ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
					))

					Expect(fakeCloudControllerClient.GetApplicationDropletCurrentCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationDropletCurrentArgsForCall(0)).To(Equal("some-app-guid"))

					Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationProcessesArgsForCall(0)).To(Equal("some-app-guid"))

					Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("some-process-guid"))
				})
			})
		})

		Context("when the app is not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app-name", "some-space-guid")
				Expect(err).To(Equal(actionerror.ApplicationNotFoundError{Name: "some-app-name"}))
				Expect(warnings).To(Equal(Warnings{"some-warning"}))
			})
		})

		Context("when getting the app processes returns an error", func() {
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
				_, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app-name", "some-space-guid")
				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(Equal(Warnings{"some-warning", "some-process-warning"}))
			})
		})

		Context("when getting the app process instances returns an error", func() {
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

				fakeCloudControllerClient.GetDropletsReturns(
					[]ccv3.Droplet{
						{
							Stack: "some-stack",
							Buildpacks: []ccv3.DropletBuildpack{
								{
									Name: "some-buildpack",
								},
							},
						},
					},
					ccv3.Warnings{"some-droplet-warning"},
					nil,
				)

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

				expectedErr = errors.New("some error")
				fakeCloudControllerClient.GetProcessInstancesReturns(
					[]ccv3.ProcessInstance{},
					ccv3.Warnings{"some-process-stats-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app-name", "some-space-guid")
				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(Equal(Warnings{"some-warning", "some-process-warning", "some-process-stats-warning"}))
			})
		})
	})
})
