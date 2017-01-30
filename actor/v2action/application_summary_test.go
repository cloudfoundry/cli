package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Summary Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil)
	})

	Describe("GetApplicationSummaryByNameSpace", func() {
		Context("when the application exists", func() {
			Context("when the application is running", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{
							{
								GUID:  "some-app-guid",
								Name:  "some-app",
								State: ccv2.ApplicationStarted,
							},
						},
						ccv2.Warnings{"app-warning"},
						nil,
					)
				})

				Context("when instance information is available", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationInstanceStatusesByApplicationReturns(
							[]ccv2.ApplicationInstanceStatus{
								{ID: 0},
								{ID: 1},
							},
							ccv2.Warnings{"instance-warning"},
							nil,
						)
					})

					It("returns the application and warnings", func() {
						app, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
						Expect(err).ToNot(HaveOccurred())
						Expect(app).To(Equal(ApplicationSummary{
							Application: Application{
								GUID:  "some-app-guid",
								Name:  "some-app",
								State: ccv2.ApplicationStarted,
							},
							RunningInstances: []ApplicationInstance{
								{ID: 0},
								{ID: 1},
							},
						}))
						Expect(warnings).To(Equal(Warnings{"app-warning", "instance-warning"}))
					})
				})

				Context("when instance information says the application is stopped", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationInstanceStatusesByApplicationReturns(
							nil,
							ccv2.Warnings{"instance-warning"},
							ccv2.AppStoppedStatsError{},
						)
					})

					It("running instances is empty and no error is returned", func() {
						app, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
						Expect(err).ToNot(HaveOccurred())
						Expect(app.RunningInstances).To(BeEmpty())
						Expect(warnings).To(Equal(Warnings{"app-warning", "instance-warning"}))
					})
				})
			})

			Context("when the application is stopped", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{
							{
								GUID:  "some-app-guid",
								Name:  "some-app",
								State: ccv2.ApplicationStopped,
							},
						},
						ccv2.Warnings{"app-warning"},
						nil,
					)
				})

				It("does not try and get application instance information", func() {
					app, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
					Expect(err).ToNot(HaveOccurred())
					Expect(app.RunningInstances).To(BeEmpty())
					Expect(warnings).To(Equal(Warnings{"app-warning"}))
					Expect(fakeCloudControllerClient.GetApplicationInstanceStatusesByApplicationCallCount()).To(Equal(0))
				})
			})
		})

		Context("when the application has routes", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{
						{
							GUID:  "some-app-guid",
							Name:  "some-app",
							State: ccv2.ApplicationStopped,
						},
					},
					ccv2.Warnings{"app-warning"},
					nil,
				)
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
					nil,
				)
			})

			It("returns the routes", func() {
				app, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
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
				Expect(warnings).To(Equal(Warnings{"app-warning", "get-application-routes-warning"}))
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

				It("returns the error", func() {
					app, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
					Expect(err).To(MatchError(expectedErr))
					Expect(app.Routes).To(BeEmpty())
					Expect(warnings).To(Equal(Warnings{"app-warning", "get-application-routes-warning"}))
				})
			})
		})

		Context("when the application's stack information exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{
						{
							GUID:  "some-app-guid",
							Name:  "some-app",
							State: ccv2.ApplicationStopped,
						},
					},
					ccv2.Warnings{"app-warning"},
					nil,
				)
				fakeCloudControllerClient.GetStackReturns(
					ccv2.Stack{Name: "some-stack"},
					ccv2.Warnings{"get-application-stack-warning"},
					nil,
				)
			})

			It("returns the stack information", func() {
				app, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(app.Stack).To(Equal(Stack{Name: "some-stack"}))
				Expect(warnings).To(Equal(Warnings{"app-warning", "get-application-stack-warning"}))
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

				It("returns the error", func() {
					app, warnings, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
					Expect(err).To(MatchError(expectedErr))
					Expect(app.Stack).To(Equal(Stack{}))
					Expect(warnings).To(Equal(Warnings{"app-warning", "get-application-stack-warning"}))
				})
			})
		})

		Context("when the application does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv2.Application{}, nil, nil)
			})

			It("returns an ApplicationNotFoundError", func() {
				_, _, err := actor.GetApplicationSummaryByNameAndSpace("some-app", "some-space-guid")
				Expect(err).To(MatchError(ApplicationNotFoundError{Name: "some-app"}))
			})
		})
	})
})
