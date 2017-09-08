package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Summary Actions", func() {
	Describe("ApplicationSummary", func() {
		Describe("StartingOrRunningInstanceCount", func() {
			It("only counts the running and starting instances", func() {
				app := ApplicationSummary{
					RunningInstances: []ApplicationInstanceWithStats{
						{State: ApplicationInstanceState(ccv2.ApplicationInstanceCrashed)},
						{State: ApplicationInstanceState(ccv2.ApplicationInstanceDown)},
						{State: ApplicationInstanceState(ccv2.ApplicationInstanceFlapping)},
						{State: ApplicationInstanceState(ccv2.ApplicationInstanceRunning)},
						{State: ApplicationInstanceState(ccv2.ApplicationInstanceStarting)},
						{State: ApplicationInstanceState(ccv2.ApplicationInstanceUnknown)},
					},
				}
				Expect(app.StartingOrRunningInstanceCount()).To(Equal(2))
			})
		})
	})

	Describe("GetApplicationSummaryByNameSpace", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
			app                       ccv2.Application
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
			actor = NewActor(fakeCloudControllerClient, nil, nil)
			app = ccv2.Application{
				GUID: "some-app-guid",
				Name: "some-app",
			}
		})

		Context("when the application does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{},
					ccv2.Warnings{"app-warning"},
					nil)
			})

			It("returns an ApplicationNotFoundError and all warnings", func() {
				_, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
				Expect(err).To(MatchError(actionerror.ApplicationNotFoundError{Name: "some-app"}))
				Expect(warnings).To(ConsistOf("app-warning"))
			})
		})

		Context("when the application exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{app},
					ccv2.Warnings{"app-warning"},
					nil)
			})

			Context("when the application is STARTED", func() {
				BeforeEach(func() {
					app.State = ccv2.ApplicationStarted
					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{app},
						ccv2.Warnings{"app-warning"},
						nil)
				})

				Context("when instance information is available", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationInstanceStatusesByApplicationReturns(
							map[int]ccv2.ApplicationInstanceStatus{
								0: {ID: 0, IsolationSegment: "isolation-segment-1"},
								1: {ID: 1, IsolationSegment: "isolation-segment-2"}, // should never happen; iso segs for 2 instances of the same app should match.
							},
							ccv2.Warnings{"stats-warning"},
							nil)
						fakeCloudControllerClient.GetApplicationInstancesByApplicationReturns(
							map[int]ccv2.ApplicationInstance{
								0: {ID: 0},
								1: {ID: 1},
							},
							ccv2.Warnings{"instance-warning"},
							nil)
					})

					It("returns the application with instance information and warnings and populates isolation segment from the first instance", func() {
						app, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
						Expect(err).ToNot(HaveOccurred())
						Expect(app).To(Equal(ApplicationSummary{
							Application: Application{
								GUID:  "some-app-guid",
								Name:  "some-app",
								State: ccv2.ApplicationStarted,
							},
							RunningInstances: []ApplicationInstanceWithStats{
								{ID: 0, IsolationSegment: "isolation-segment-1"},
								{ID: 1, IsolationSegment: "isolation-segment-2"},
							},
							IsolationSegment: "isolation-segment-1",
						}))
						Expect(warnings).To(ConsistOf("app-warning", "stats-warning", "instance-warning"))
					})
				})

				Context("when instance information is not available", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationInstanceStatusesByApplicationReturns(
							nil,
							ccv2.Warnings{"stats-warning"},
							ccerror.ApplicationStoppedStatsError{})
					})

					It("returns the empty list of instances and all warnings", func() {
						app, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
						Expect(err).ToNot(HaveOccurred())
						Expect(app.RunningInstances).To(BeEmpty())
						Expect(warnings).To(ConsistOf("app-warning", "stats-warning"))
					})
				})
			})

			Context("when the application is not STARTED", func() {
				BeforeEach(func() {
					app.State = ccv2.ApplicationStopped
				})

				It("does not try and get application instance information", func() {
					app, _, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
					Expect(err).ToNot(HaveOccurred())
					Expect(app.RunningInstances).To(BeEmpty())

					Expect(fakeCloudControllerClient.GetApplicationInstanceStatusesByApplicationCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetApplicationInstancesByApplicationCallCount()).To(Equal(0))
				})
			})

			Context("when the app has routes", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationRoutesReturns(
						[]ccv2.Route{
							{
								GUID: "some-route-1-guid",
								Host: "host-1",
							},
							{
								GUID: "some-route-2-guid",
								Host: "host-2",
							},
						},
						ccv2.Warnings{"get-application-routes-warning"},
						nil)
				})

				It("returns the routes and all warnings", func() {
					app, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("app-warning", "get-application-routes-warning"))
					Expect(app.Routes).To(ConsistOf(
						Route{
							GUID: "some-route-1-guid",
							Host: "host-1",
						},
						Route{
							GUID: "some-route-2-guid",
							Host: "host-2",
						},
					))
				})

				Context("when an error is encountered while getting routes", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("get routes error")
						fakeCloudControllerClient.GetApplicationRoutesReturns(
							nil,
							ccv2.Warnings{"get-application-routes-warning"},
							expectedErr,
						)
					})

					It("returns the error and all warnings", func() {
						app, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
						Expect(err).To(MatchError(expectedErr))
						Expect(app.Routes).To(BeEmpty())
						Expect(warnings).To(ConsistOf("app-warning", "get-application-routes-warning"))
					})
				})
			})

			Context("when the app has stack information", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetStackReturns(
						ccv2.Stack{Name: "some-stack"},
						ccv2.Warnings{"get-application-stack-warning"},
						nil)
				})

				It("returns the stack information and all warnings", func() {
					app, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("app-warning", "get-application-stack-warning"))
					Expect(app.Stack).To(Equal(Stack{Name: "some-stack"}))
				})

				Context("when an error is encountered while getting stack", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("get stack error")
						fakeCloudControllerClient.GetStackReturns(
							ccv2.Stack{},
							ccv2.Warnings{"get-application-stack-warning"},
							expectedErr,
						)
					})

					It("returns the error and all warnings", func() {
						app, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
						Expect(err).To(MatchError(expectedErr))
						Expect(app.Stack).To(Equal(Stack{}))
						Expect(warnings).To(ConsistOf("app-warning", "get-application-stack-warning"))
					})
				})
			})
		})
	})
})
