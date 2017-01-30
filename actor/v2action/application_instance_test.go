package v2action_test

import (
	"errors"
	"time"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Instance Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil)
	})

	Describe("ApplicationInstance", func() {
		var instance ApplicationInstance

		BeforeEach(func() {
			instance = ApplicationInstance{}
		})

		Describe("StartTime", func() {
			It("returns a the time the instance started", func() {
				instance.Uptime = 0
				Expect(instance.StartTime()).To(BeTemporally("~", time.Now(), time.Second))

				instance.Uptime = 10
				Expect(instance.StartTime()).To(BeTemporally("~", time.Now(), 11*time.Second))
			})
		})
	})

	Describe("GetApplicationInstancesByApplication", func() {
		Context("when the application exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationInstanceStatusesByApplicationReturns(
					[]ccv2.ApplicationInstanceStatus{
						{ID: 0},
						{ID: 1},
					},
					ccv2.Warnings{"instance-warning-1", "instance-warning-2"},
					nil,
				)
			})

			It("returns the application and warnings", func() {
				instances, warnings, err := actor.GetApplicationInstancesByApplication("some-app-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(instances).To(Equal([]ApplicationInstance{
					{ID: 0},
					{ID: 1},
				}))
				Expect(warnings).To(ConsistOf("instance-warning-1", "instance-warning-2"))

				Expect(fakeCloudControllerClient.GetApplicationInstanceStatusesByApplicationCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationInstanceStatusesByApplicationArgsForCall(0)).To(Equal("some-app-guid"))
			})
		})

		Context("when the client returns an error", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("banana")
				fakeCloudControllerClient.GetApplicationInstanceStatusesByApplicationReturns(
					nil,
					ccv2.Warnings{"instance-warning-1", "instance-warning-2"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, _, err := actor.GetApplicationInstancesByApplication("some-app-guid")
				Expect(err).To(MatchError(expectedErr))
			})
		})

		Context("the application does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationInstanceStatusesByApplicationReturns(nil, ccv2.Warnings{"instance-warning-1", "instance-warning-2"}, cloudcontroller.ResourceNotFoundError{})
			})

			It("returns an ApplicationInstancesNotFoundError", func() {
				_, warnings, err := actor.GetApplicationInstancesByApplication("some-app-guid")
				Expect(err).To(MatchError(ApplicationInstancesNotFoundError{ApplicationGUID: "some-app-guid"}))
				Expect(warnings).To(ConsistOf("instance-warning-1", "instance-warning-2"))
			})
		})
	})
})
