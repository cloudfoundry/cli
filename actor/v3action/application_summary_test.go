package v3action_test

import (
	"errors"
	"net/url"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

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
		actor = NewActor(fakeCloudControllerClient, nil)
	})

	Describe("GetApplicationSummaryByNameAndSpace", func() {
		Context("when the app exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{
							Name:  "some-app-name",
							GUID:  "some-app-guid",
							State: "RUNNING",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)

				fakeCloudControllerClient.GetApplicationCurrentDropletReturns(
					ccv3.Droplet{
						Stack: "some-stack",
						Buildpacks: []ccv3.Buildpack{
							{
								Name: "some-buildpack",
							},
						},
					},
					ccv3.Warnings{"some-droplet-warning"},
					nil,
				)

				fakeCloudControllerClient.GetApplicationProcessesReturns(
					[]ccv3.Process{
						{
							GUID:       "some-process-guid",
							Type:       "some-type",
							MemoryInMB: 32,
						},
					},
					ccv3.Warnings{"some-process-warning"},
					nil,
				)

				fakeCloudControllerClient.GetProcessInstancesReturns(
					[]ccv3.Instance{
						{
							State:       "RUNNING",
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

			It("returns the summary and warnings", func() {
				summary, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app-name", "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(summary).To(Equal(ApplicationSummary{
					Application: Application{
						Name:  "some-app-name",
						GUID:  "some-app-guid",
						State: "RUNNING",
					},
					CurrentDroplet: Droplet{
						Stack: "some-stack",
						Buildpacks: []Buildpack{
							{
								Name: "some-buildpack",
							},
						},
					},
					Processes: []Process{
						Process{
							MemoryInMB: 32,
							Type:       "some-type",
							Instances: []Instance{
								{
									State:       "RUNNING",
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
				Expect(warnings).To(Equal(Warnings{"some-warning", "some-droplet-warning", "some-process-warning", "some-process-stats-warning"}))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				expectedQuery := url.Values{
					"names":       []string{"some-app-name"},
					"space_guids": []string{"some-space-guid"},
				}
				query := fakeCloudControllerClient.GetApplicationsArgsForCall(0)
				Expect(query).To(Equal(expectedQuery))

				Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(1))
				appGUID := fakeCloudControllerClient.GetApplicationProcessesArgsForCall(0)
				Expect(appGUID).To(Equal("some-app-guid"))

				Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(1))
				processGUID := fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)
				Expect(processGUID).To(Equal("some-process-guid"))
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
				Expect(err).To(Equal(ApplicationNotFoundError{Name: "some-app-name"}))
				Expect(warnings).To(Equal(Warnings{"some-warning"}))
			})
		})

		Context("when getting the current droplet returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{
							Name:  "some-app-name",
							GUID:  "some-app-guid",
							State: "RUNNING",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)

				expectedErr = errors.New("some error")
				fakeCloudControllerClient.GetApplicationCurrentDropletReturns(
					ccv3.Droplet{},
					ccv3.Warnings{"some-droplet-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app-name", "some-space-guid")
				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(Equal(Warnings{"some-warning", "some-droplet-warning"}))
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
							State: "RUNNING",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)

				fakeCloudControllerClient.GetApplicationCurrentDropletReturns(
					ccv3.Droplet{
						Stack: "some-stack",
						Buildpacks: []ccv3.Buildpack{
							{
								Name: "some-buildpack",
							},
						},
					},
					ccv3.Warnings{"some-droplet-warning"},
					nil,
				)

				expectedErr = errors.New("some error")
				fakeCloudControllerClient.GetApplicationProcessesReturns(
					[]ccv3.Process{},
					ccv3.Warnings{"some-process-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app-name", "some-space-guid")
				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(Equal(Warnings{"some-warning", "some-droplet-warning", "some-process-warning"}))
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
							State: "RUNNING",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)

				fakeCloudControllerClient.GetApplicationCurrentDropletReturns(
					ccv3.Droplet{
						Stack: "some-stack",
						Buildpacks: []ccv3.Buildpack{
							{
								Name: "some-buildpack",
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
					[]ccv3.Instance{},
					ccv3.Warnings{"some-process-stats-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app-name", "some-space-guid")
				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(Equal(Warnings{"some-warning", "some-droplet-warning", "some-process-warning", "some-process-stats-warning"}))
			})
		})
	})
})
