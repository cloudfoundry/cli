package v3action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application with ProcessSummary Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("GetApplicationsWithProcessesBySpace", func() {
		Context("when there are apps", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{
							Name:  "some-app-name-1",
							GUID:  "some-app-guid-1",
							State: constant.ApplicationStarted,
						},
						{
							Name:  "some-app-name-2",
							GUID:  "some-app-guid-2",
							State: constant.ApplicationStopped,
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)

				fakeCloudControllerClient.GetApplicationProcessesReturnsOnCall(
					0,
					[]ccv3.Process{
						{
							GUID: "some-process-guid-1",
							Type: "some-process-type-1",
						},
						{
							GUID: "some-process-guid-2",
							Type: "some-process-type-2",
						},
					},
					ccv3.Warnings{"some-process-warning-1"},
					nil,
				)
				fakeCloudControllerClient.GetApplicationProcessesReturnsOnCall(
					1,
					[]ccv3.Process{
						{
							GUID: "some-process-guid-3",
							Type: "some-process-type-3",
						},
					},
					ccv3.Warnings{"some-process-warning-2"},
					nil,
				)

				fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(
					0,
					[]ccv3.ProcessInstance{{State: constant.ProcessInstanceRunning}, {State: constant.ProcessInstanceDown}, {State: constant.ProcessInstanceRunning}},
					ccv3.Warnings{"some-process-stats-warning-1"},
					nil,
				)
				fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(
					1,
					[]ccv3.ProcessInstance{{State: constant.ProcessInstanceRunning}, {State: constant.ProcessInstanceRunning}},
					ccv3.Warnings{"some-process-stats-warning-2"},
					nil,
				)
				fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(
					2,
					[]ccv3.ProcessInstance{{State: constant.ProcessInstanceDown}},
					ccv3.Warnings{"some-process-stats-warning-3"},
					nil,
				)
			})

			It("returns app summaries and warnings", func() {
				summaries, warnings, err := actor.GetApplicationsWithProcessesBySpace("some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(summaries).To(ConsistOf(
					ApplicationWithProcessSummary{
						Application: Application{
							Name:  "some-app-name-1",
							GUID:  "some-app-guid-1",
							State: constant.ApplicationStarted,
						},
						ProcessSummaries: []ProcessSummary{
							{
								Process:         Process{GUID: "some-process-guid-1", Type: "some-process-type-1"},
								InstanceDetails: []ProcessInstance{{State: constant.ProcessInstanceRunning}, {State: constant.ProcessInstanceDown}, {State: constant.ProcessInstanceRunning}},
							},
							{
								Process:         Process{GUID: "some-process-guid-2", Type: "some-process-type-2"},
								InstanceDetails: []ProcessInstance{{State: constant.ProcessInstanceRunning}, {State: constant.ProcessInstanceRunning}},
							},
						},
					},
					ApplicationWithProcessSummary{
						Application: Application{
							Name:  "some-app-name-2",
							GUID:  "some-app-guid-2",
							State: constant.ApplicationStopped,
						},
						ProcessSummaries: []ProcessSummary{
							{
								Process:         Process{GUID: "some-process-guid-3", Type: "some-process-type-3"},
								InstanceDetails: []ProcessInstance{{State: constant.ProcessInstanceDown}},
							},
						},
					}))
				Expect(warnings).To(ConsistOf("some-warning", "some-process-warning-1", "some-process-stats-warning-1", "some-process-stats-warning-2", "some-process-warning-2", "some-process-stats-warning-3"))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
					ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.NameOrder}},
				))

				Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(2))
				Expect(fakeCloudControllerClient.GetApplicationProcessesArgsForCall(0)).To(Equal("some-app-guid-1"))
				Expect(fakeCloudControllerClient.GetApplicationProcessesArgsForCall(1)).To(Equal("some-app-guid-2"))

				Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(3))
				Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("some-process-guid-1"))
				Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(1)).To(Equal("some-process-guid-2"))
				Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(2)).To(Equal("some-process-guid-3"))
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
					[]ccv3.Process{},
					ccv3.Warnings{"some-process-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, warnings, err := actor.GetApplicationsWithProcessesBySpace("some-space-guid")
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
				_, warnings, err := actor.GetApplicationsWithProcessesBySpace("some-space-guid")
				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(Equal(Warnings{"some-warning", "some-process-warning", "some-process-stats-warning"}))
			})
		})
	})
})
