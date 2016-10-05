package v2actions_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/actors/v2actions/v2actionsfakes"
	"code.cloudfoundry.org/cli/api/cloudcontrollerv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v2actionsfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionsfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient)
	})

	Describe("GetApplicationBySpace", func() {
		Context("when the application exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]cloudcontrollerv2.Application{
						{
							GUID: "some-app-guid",
							Name: "some-app",
						},
					},
					cloudcontrollerv2.Warnings{"foo"},
					nil,
				)
			})

			It("returns the application and warnings", func() {
				app, warnings, err := actor.GetApplicationBySpace("some-app", "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(app).To(Equal(Application{
					GUID: "some-app-guid",
					Name: "some-app",
				}))
				Expect(warnings).To(Equal(Warnings{"foo"}))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf([]cloudcontrollerv2.Query{
					cloudcontrollerv2.Query{
						Filter:   cloudcontrollerv2.NameFilter,
						Operator: cloudcontrollerv2.EqualOperator,
						Value:    "some-app",
					},
					cloudcontrollerv2.Query{
						Filter:   cloudcontrollerv2.SpaceGUIDFilter,
						Operator: cloudcontrollerv2.EqualOperator,
						Value:    "some-space-guid",
					},
				}))
			})
		})

		Context("when the application does not exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]cloudcontrollerv2.Application{}, nil, nil)
			})

			It("returns an ApplicationNotFoundError", func() {
				_, _, err := actor.GetApplicationBySpace("some-app", "some-space-guid")
				Expect(err).To(MatchError(ApplicationNotFoundError{Name: "some-app"}))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationsReturns([]cloudcontrollerv2.Application{}, nil, expectedError)
			})

			It("returns the error", func() {
				_, _, err := actor.GetApplicationBySpace("some-app", "some-space-guid")
				Expect(err).To(MatchError(expectedError))
			})
		})
	})
})
