package v2action_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Instance With Stats Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("ApplicationInstanceWithStats", func() {
		var instance ApplicationInstanceWithStats

		BeforeEach(func() {
			instance = ApplicationInstanceWithStats{}
		})

		Describe("TimeSinceCreation", func() {
			It("returns the time the instance started", func() {
				instance.Since = 1485985587.12345
				Expect(instance.TimeSinceCreation()).To(Equal(time.Unix(1485985587, 0)))
			})
		})
	})

	Describe("GetApplicationInstancesWithStatsByApplication", func() {
		When("the application exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationApplicationInstanceStatusesReturns(
					map[int]ccv2.ApplicationInstanceStatus{
						0: {
							ID:               0,
							CPU:              100,
							Memory:           100,
							MemoryQuota:      200,
							Disk:             50,
							DiskQuota:        100,
							IsolationSegment: "some-isolation-segment",
						},
						1: {ID: 1, CPU: 200},
					},
					ccv2.Warnings{"stats-warning-1", "stats-warning-2"},
					nil)

				fakeCloudControllerClient.GetApplicationApplicationInstancesReturns(
					map[int]ccv2.ApplicationInstance{
						0: {ID: 0, Details: "hello", Since: 1485985587.12345, State: constant.ApplicationInstanceRunning},
						1: {ID: 1, Details: "hi", Since: 1485985587.567},
					},
					ccv2.Warnings{"instance-warning-1", "instance-warning-2"},
					nil)
			})

			It("returns the application instances and all warnings", func() {
				instances, warnings, err := actor.GetApplicationInstancesWithStatsByApplication("some-app-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(instances).To(ConsistOf(
					ApplicationInstanceWithStats{
						ID:               0,
						CPU:              100,
						Memory:           100,
						MemoryQuota:      200,
						Disk:             50,
						DiskQuota:        100,
						Details:          "hello",
						IsolationSegment: "some-isolation-segment",
						Since:            1485985587.12345,
						State:            ApplicationInstanceState(constant.ApplicationInstanceRunning),
					},
					ApplicationInstanceWithStats{ID: 1, CPU: 200, Details: "hi", Since: 1485985587.567}))
				Expect(warnings).To(ConsistOf(
					"stats-warning-1",
					"stats-warning-2",
					"instance-warning-1",
					"instance-warning-2"))

				Expect(fakeCloudControllerClient.GetApplicationApplicationInstanceStatusesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationApplicationInstanceStatusesArgsForCall(0)).To(Equal("some-app-guid"))
				Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesArgsForCall(0)).To(Equal("some-app-guid"))
			})
		})

		When("an error is encountered", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("banana")
				fakeCloudControllerClient.GetApplicationApplicationInstanceStatusesReturns(
					nil,
					ccv2.Warnings{"stats-warning"},
					nil)
				fakeCloudControllerClient.GetApplicationApplicationInstancesReturns(
					nil,
					ccv2.Warnings{"instances-warning"},
					expectedErr)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := actor.GetApplicationInstancesWithStatsByApplication("some-app-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("stats-warning", "instances-warning"))
			})

			When("the application does not exist", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationApplicationInstanceStatusesReturns(
						nil,
						nil,
						ccerror.ResourceNotFoundError{})
				})

				It("returns an ApplicationInstancesNotFoundError", func() {
					_, _, err := actor.GetApplicationInstancesWithStatsByApplication("some-app-guid")
					Expect(err).To(MatchError(actionerror.ApplicationInstancesNotFoundError{ApplicationGUID: "some-app-guid"}))
				})
			})

			When("the desired state of the app is STARTED", func() {
				When("the app has not been staged yet", func() {
					When("getting instance stats returns a CF-AppStoppedStatsError", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationApplicationInstanceStatusesReturns(
								nil,
								nil,
								ccerror.ApplicationStoppedStatsError{})
						})

						It("returns an ApplicationInstancesNotFoundError", func() {
							_, _, err := actor.GetApplicationInstancesWithStatsByApplication("some-app-guid")
							Expect(err).To(MatchError(actionerror.ApplicationInstancesNotFoundError{ApplicationGUID: "some-app-guid"}))
						})
					})
				})

				When("the app is not yet running", func() {
					When("getting instance stats returns a CF-AppStoppedStatsError", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationApplicationInstanceStatusesReturns(
								nil,
								nil,
								ccerror.ApplicationStoppedStatsError{})
						})

						It("returns an ApplicationInstancesNotFoundError", func() {
							_, _, err := actor.GetApplicationInstancesWithStatsByApplication("some-app-guid")
							Expect(err).To(MatchError(actionerror.ApplicationInstancesNotFoundError{ApplicationGUID: "some-app-guid"}))
						})
					})
				})
			})
		})

		When("getting the stats and instances return different number of results", func() {
			When("an instance is missing from stats", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationApplicationInstanceStatusesReturns(
						map[int]ccv2.ApplicationInstanceStatus{
							0: {ID: 0},
						}, nil, nil)

					fakeCloudControllerClient.GetApplicationApplicationInstancesReturns(
						map[int]ccv2.ApplicationInstance{
							0: {ID: 0},
							1: {ID: 1, Details: "backend details"},
						}, nil, nil)
				})

				It("sets the detail field to incomplete", func() {
					instances, _, err := actor.GetApplicationInstancesWithStatsByApplication("some-app-guid")
					Expect(err).ToNot(HaveOccurred())
					Expect(instances).To(ConsistOf(
						ApplicationInstanceWithStats{ID: 0},
						ApplicationInstanceWithStats{ID: 1, Details: "backend details (Unable to retrieve information)"},
					))
				})
			})

			When("an instance is missing from instances", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationApplicationInstanceStatusesReturns(
						map[int]ccv2.ApplicationInstanceStatus{
							0: {ID: 0},
							1: {ID: 1},
						}, nil, nil)

					fakeCloudControllerClient.GetApplicationApplicationInstancesReturns(
						map[int]ccv2.ApplicationInstance{
							0: {ID: 0},
						}, nil, nil)
				})

				It("sets the detail field to incomplete", func() {
					instances, _, err := actor.GetApplicationInstancesWithStatsByApplication("some-app-guid")
					Expect(err).ToNot(HaveOccurred())
					Expect(instances).To(ConsistOf(
						ApplicationInstanceWithStats{ID: 0},
						ApplicationInstanceWithStats{ID: 1, Details: "(Unable to retrieve information)"},
					))
				})
			})
		})
	})
})
