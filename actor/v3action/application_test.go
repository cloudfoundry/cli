package v3action_test

import (
	"errors"
	"net/url"
	"time"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil)
	})

	Describe("GetApplicationByNameAndSpace", func() {
		Context("when the app exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{
							Name: "some-app-name",
							GUID: "some-app-guid",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns the application and warnings", func() {
				app, warnings, err := actor.GetApplicationByNameAndSpace("some-app-name", "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(app).To(Equal(Application{
					Name: "some-app-name",
					GUID: "some-app-guid",
				}))
				Expect(warnings).To(Equal(Warnings{"some-warning"}))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				expectedQuery := url.Values{
					"names":       []string{"some-app-name"},
					"space_guids": []string{"some-space-guid"},
				}
				query := fakeCloudControllerClient.GetApplicationsArgsForCall(0)
				Expect(query).To(Equal(expectedQuery))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the warnings and the error", func() {
				_, warnings, err := actor.GetApplicationByNameAndSpace("some-app-name", "some-space-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(expectedError))
				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				expectedQuery := url.Values{
					"names":       []string{"some-app-name"},
					"space_guids": []string{"some-space-guid"},
				}
				query := fakeCloudControllerClient.GetApplicationsArgsForCall(0)
				Expect(query).To(Equal(expectedQuery))
			})
		})

		Context("when the app does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns an ApplicationNotFoundError and the warnings", func() {
				_, warnings, err := actor.GetApplicationByNameAndSpace("some-app-name", "some-space-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(
					ApplicationNotFoundError{Name: "some-app-name"}))
				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				expectedQuery := url.Values{
					"names":       []string{"some-app-name"},
					"space_guids": []string{"some-space-guid"},
				}
				query := fakeCloudControllerClient.GetApplicationsArgsForCall(0)
				Expect(query).To(Equal(expectedQuery))
			})
		})
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
							GUID: "some-process-guid",
							Type: "some-type",
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
							Type: "some-type",
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

	Describe("CreateApplicationByNameAndSpace", func() {
		Context("when the app successfully gets created", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationReturns(
					ccv3.Application{
						Name: "some-app-name",
						GUID: "some-app-guid",
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("creates and returns the application and warnings", func() {
				app, warnings, err := actor.CreateApplicationByNameAndSpace("some-app-name", "some-space-guid")

				Expect(err).ToNot(HaveOccurred())
				Expect(app).To(Equal(Application{
					Name: "some-app-name",
					GUID: "some-app-guid",
				}))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.CreateApplicationCallCount()).To(Equal(1))
				expectedApp := ccv3.Application{
					Name: "some-app-name",
					Relationships: ccv3.Relationships{
						ccv3.SpaceRelationship: ccv3.Relationship{GUID: "some-space-guid"},
					},
				}
				Expect(fakeCloudControllerClient.CreateApplicationArgsForCall(0)).To(Equal(expectedApp))
			})
		})

		Context("when the cc client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.CreateApplicationReturns(
					ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError,
				)
			})

			It("raises the error and warnings", func() {
				_, warnings, err := actor.CreateApplicationByNameAndSpace("some-app-name", "some-space-guid")

				Expect(err).To(MatchError(expectedError))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})

		Context("when the cc client response contains an UnprocessableEntityError", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationReturns(
					ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					ccerror.UnprocessableEntityError{},
				)
			})

			It("raises the error as ApplicationAlreadyExistsError and warnings", func() {
				_, warnings, err := actor.CreateApplicationByNameAndSpace("some-app-name", "some-space-guid")

				Expect(err).To(MatchError(ApplicationAlreadyExistsError{Name: "some-app-name"}))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})
	})

	Describe("SetApplicationDroplet", func() {
		Context("when there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.SetApplicationDropletReturns(
					ccv3.Relationship{GUID: "some-droplet-guid"},
					ccv3.Warnings{"set-application-droplet-warning"},
					nil,
				)
			})

			It("sets the app's droplet", func() {
				warnings, err := actor.SetApplicationDroplet("some-app-name", "some-space-guid", "some-droplet-guid")

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-applications-warning", "set-application-droplet-warning"))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				queryURL := fakeCloudControllerClient.GetApplicationsArgsForCall(0)
				query := url.Values{"names": []string{"some-app-name"}, "space_guids": []string{"some-space-guid"}}
				Expect(queryURL).To(Equal(query))

				Expect(fakeCloudControllerClient.SetApplicationDropletCallCount()).To(Equal(1))
				appGUID, dropletGUID := fakeCloudControllerClient.SetApplicationDropletArgsForCall(0)
				Expect(appGUID).To(Equal("some-app-guid"))
				Expect(dropletGUID).To(Equal("some-droplet-guid"))
			})
		})

		Context("when getting the application fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some get application error")

				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"get-applications-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				warnings, err := actor.SetApplicationDroplet("some-app-name", "some-space-guid", "some-droplet-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("get-applications-warning"))
			})
		})

		Context("when setting the droplet fails", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("some set application-droplet error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.SetApplicationDropletReturns(
					ccv3.Relationship{},
					ccv3.Warnings{"set-application-droplet-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				warnings, err := actor.SetApplicationDroplet("some-app-name", "some-space-guid", "some-droplet-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("get-applications-warning", "set-application-droplet-warning"))
			})
		})
	})

	Describe("StartApplication", func() {
		Context("when there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.StartApplicationReturns(
					ccv3.Application{GUID: "some-app-guid"},
					ccv3.Warnings{"start-application-warning"},
					nil,
				)
			})

			It("sets the app's droplet", func() {
				app, warnings, err := actor.StartApplication("some-app-name", "some-space-guid")

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-applications-warning", "start-application-warning"))
				Expect(app).To(Equal(Application{GUID: "some-app-guid"}))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				queryURL := fakeCloudControllerClient.GetApplicationsArgsForCall(0)
				query := url.Values{"names": []string{"some-app-name"}, "space_guids": []string{"some-space-guid"}}
				Expect(queryURL).To(Equal(query))

				Expect(fakeCloudControllerClient.StartApplicationCallCount()).To(Equal(1))
				appGUID := fakeCloudControllerClient.StartApplicationArgsForCall(0)
				Expect(appGUID).To(Equal("some-app-guid"))
			})
		})

		Context("when getting the application fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some get application error")

				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"get-applications-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, warnings, err := actor.StartApplication("some-app-name", "some-space-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("get-applications-warning"))
			})
		})

		Context("when starting the application fails", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("some set start-application error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.StartApplicationReturns(
					ccv3.Application{},
					ccv3.Warnings{"start-application-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, warnings, err := actor.StartApplication("some-app-name", "some-space-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("get-applications-warning", "start-application-warning"))
			})
		})
	})

	Describe("ApplicationSummary", func() {
		var summary ApplicationSummary

		BeforeEach(func() {
			summary = ApplicationSummary{
				Processes: []Process{
					Process{
						Instances: []Instance{
							Instance{MemoryUsage: 1000000},
							Instance{MemoryUsage: 2000000},
						},
					},
					Process{
						Instances: []Instance{
							Instance{MemoryUsage: 3000000},
							Instance{MemoryUsage: 4000000},
						},
					},
				},
			}
		})

		Describe("TotalMemoryUsage", func() {
			It("returns the total memory usage of all processes and instances", func() {
				Expect(summary.TotalMemoryUsage()).To(Equal(uint64(10000000)))
			})
		})
	})

	Describe("Instance", func() {
		Describe("StartTime", func() {
			It("returns the time that the instance started", func() {
				instance := Instance{Uptime: 86400}
				Expect(instance.StartTime()).To(BeTemporally("~", time.Now().Add(-24*time.Hour), 10*time.Second))
			})
		})
	})

	Describe("Process", func() {
		var process Process
		BeforeEach(func() {
			process = Process{
				Instances: []Instance{
					Instance{State: "RUNNING"},
					Instance{State: "RUNNING"},
					Instance{State: "STOPPED"},
				},
			}
		})

		Describe("TotalInstanceCount", func() {
			It("returns the total number of instances", func() {
				Expect(process.TotalInstanceCount()).To(Equal(3))
			})
		})

		Describe("HealthyInstanceCount", func() {
			It("returns the total number of RUNNING instances", func() {
				Expect(process.HealthyInstanceCount()).To(Equal(2))
			})
		})
	})
})
