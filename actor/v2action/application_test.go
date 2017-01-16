package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil)
	})

	Describe("GetApplicationBySpace", func() {
		Context("when the application exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{
						{
							GUID: "some-app-guid",
							Name: "some-app",
						},
					},
					ccv2.Warnings{"foo"},
					nil,
				)
			})

			It("returns the application and warnings", func() {
				app, warnings, err := actor.GetApplicationByNameAndSpace("some-app", "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(app).To(Equal(Application{
					GUID: "some-app-guid",
					Name: "some-app",
				}))
				Expect(warnings).To(Equal(Warnings{"foo"}))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf([]ccv2.Query{
					ccv2.Query{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Value:    "some-app",
					},
					ccv2.Query{
						Filter:   ccv2.SpaceGUIDFilter,
						Operator: ccv2.EqualOperator,
						Value:    "some-space-guid",
					},
				}))
			})
		})

		Context("when the application does not exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv2.Application{}, nil, nil)
			})

			It("returns an ApplicationNotFoundError", func() {
				_, _, err := actor.GetApplicationByNameAndSpace("some-app", "some-space-guid")
				Expect(err).To(MatchError(ApplicationNotFoundError{Name: "some-app"}))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationsReturns([]ccv2.Application{}, nil, expectedError)
			})

			It("returns the error", func() {
				_, _, err := actor.GetApplicationByNameAndSpace("some-app", "some-space-guid")
				Expect(err).To(MatchError(expectedError))
			})
		})
	})

	Describe("GetRouteApplications", func() {
		Context("when the CC client returns no errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRouteApplicationsReturns(
					[]ccv2.Application{
						{
							GUID: "application-guid",
							Name: "application-name",
						},
					}, ccv2.Warnings{"route-applications-warning"}, nil)
			})
			It("returns the applications bound to the route and warnings", func() {
				applications, warnings, err := actor.GetRouteApplications("route-guid", nil)
				Expect(fakeCloudControllerClient.GetRouteApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetRouteApplicationsArgsForCall(0)).To(Equal("route-guid"))

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("route-applications-warning"))
				Expect(applications).To(ConsistOf(
					Application{
						GUID: "application-guid",
						Name: "application-name",
					},
				))
			})
		})

		Context("when the CC client returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRouteApplicationsReturns(
					[]ccv2.Application{}, ccv2.Warnings{"route-applications-warning"}, errors.New("get-route-applications-error"))
			})

			It("returns the error and warnings", func() {
				apps, warnings, err := actor.GetRouteApplications("route-guid", nil)
				Expect(fakeCloudControllerClient.GetRouteApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetRouteApplicationsArgsForCall(0)).To(Equal("route-guid"))

				Expect(err).To(MatchError("get-route-applications-error"))
				Expect(warnings).To(ConsistOf("route-applications-warning"))
				Expect(apps).To(BeNil())
			})
		})

		Context("when a query parameter exists", func() {
			It("passes the query to the client", func() {
				expectedQuery := []ccv2.Query{
					{
						Filter:   "route_guid",
						Operator: ":",
						Value:    "route-guid",
					}}

				_, _, err := actor.GetRouteApplications("route-guid", expectedQuery)
				Expect(err).ToNot(HaveOccurred())
				_, query := fakeCloudControllerClient.GetRouteApplicationsArgsForCall(0)
				Expect(query).To(Equal(expectedQuery))
			})
		})
	})
})
