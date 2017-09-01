package v3action_test

import (
	"errors"
	"time"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Process Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil)
	})

	Describe("Instance", func() {
		Describe("StartTime", func() {
			It("returns the time that the instance started", func() {
				instance := Instance{Uptime: 86400}
				Expect(instance.StartTime()).To(BeTemporally("~", time.Now().Add(-24*time.Hour), 10*time.Second))
			})
		})
	})

	Describe("ProcessSummary", func() {
		var summary ProcessSummary
		BeforeEach(func() {
			summary = ProcessSummary{
				InstanceDetails: []Instance{
					Instance{State: "RUNNING"},
					Instance{State: "RUNNING"},
					Instance{State: "STOPPED"},
				},
			}
		})

		Describe("TotalInstanceCount", func() {
			It("returns the total number of instances", func() {
				Expect(summary.TotalInstanceCount()).To(Equal(3))
			})
		})

		Describe("HealthyInstanceCount", func() {
			It("returns the total number of RUNNING instances", func() {
				Expect(summary.HealthyInstanceCount()).To(Equal(2))
			})
		})
	})

	Describe("ProcessSummaries", func() {
		var summaries ProcessSummaries

		BeforeEach(func() {
			summaries = ProcessSummaries{
				{
					Process: Process{
						Type: "worker",
					},
					InstanceDetails: []Instance{
						{State: "RUNNING"},
						{State: "STOPPED"},
					},
				},
				{
					Process: Process{
						Type: "console",
					},
					InstanceDetails: []Instance{
						{State: "RUNNING"},
					},
				},
				{
					Process: Process{
						Type: "web",
					},
					InstanceDetails: []Instance{
						{State: "RUNNING"},
						{State: "RUNNING"},
						{State: "STOPPED"},
					},
				},
			}
		})

		Describe("Sort", func() {
			It("sorts processes with web first and then alphabetically sorted", func() {
				summaries.Sort()
				Expect(summaries[0].Type).To(Equal("web"))
				Expect(summaries[1].Type).To(Equal("console"))
				Expect(summaries[2].Type).To(Equal("worker"))
			})
		})

		Describe("String", func() {
			It("returns all processes and their instance count ", func() {
				Expect(summaries.String()).To(Equal("web:2/3, console:1/1, worker:1/2"))
			})
		})
	})

	Describe("ScaleProcessByApplication", func() {
		var passedProcess Process

		BeforeEach(func() {
			passedProcess = Process{
				Type:       "web",
				Instances:  types.NullInt{Value: 2, IsSet: true},
				MemoryInMB: types.NullUint64{Value: 100, IsSet: true},
				DiskInMB:   types.NullUint64{Value: 200, IsSet: true},
			}
		})

		Context("when no errors are encountered scaling the application process", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationProcessScaleReturns(
					ccv3.Warnings{"scale-process-warning"},
					nil)
			})

			It("scales correct process", func() {
				warnings, err := actor.ScaleProcessByApplication("some-app-guid", passedProcess)

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("scale-process-warning"))

				Expect(fakeCloudControllerClient.CreateApplicationProcessScaleCallCount()).To(Equal(1))
				appGUIDArg, processArg := fakeCloudControllerClient.CreateApplicationProcessScaleArgsForCall(0)
				Expect(appGUIDArg).To(Equal("some-app-guid"))
				Expect(processArg).To(Equal(ccv3.Process{
					Type:       "web",
					Instances:  passedProcess.Instances,
					MemoryInMB: passedProcess.MemoryInMB,
					DiskInMB:   passedProcess.DiskInMB,
				}))
			})
		})

		Context("when an error is encountered scaling the application process", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("scale process error")
				fakeCloudControllerClient.CreateApplicationProcessScaleReturns(
					ccv3.Warnings{"scale-process-warning"},
					expectedErr)
			})

			It("returns the error and all warnings", func() {
				warnings, err := actor.ScaleProcessByApplication("some-app-guid", passedProcess)
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("scale-process-warning"))
			})
		})

		Context("when a ProcessNotFoundError error is encountered scaling the application process", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationProcessScaleReturns(
					ccv3.Warnings{"scale-process-warning"},
					ccerror.ProcessNotFoundError{},
				)
			})

			It("returns the error and all warnings", func() {
				warnings, err := actor.ScaleProcessByApplication("some-app-guid", passedProcess)
				Expect(err).To(Equal(ProcessNotFoundError{ProcessType: "web"}))
				Expect(warnings).To(ConsistOf("scale-process-warning"))
			})
		})
	})

	Describe("GetProcessByApplicationAndProcessType", func() {
		Context("when CC returns a process", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
					ccv3.Process{
						Type:       "web",
						Instances:  types.NullInt{Value: 2, IsSet: true},
						MemoryInMB: types.NullUint64{Value: 100, IsSet: true},
						DiskInMB:   types.NullUint64{Value: 200, IsSet: true},
					},
					ccv3.Warnings{"get-process-warning"},
					nil,
				)
			})

			It("returns the process", func() {
				process, warnings, err := actor.GetProcessByApplicationAndProcessType("some-app-guid", "web")

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-process-warning"))
				Expect(process).To(Equal(Process{
					Type:       "web",
					Instances:  types.NullInt{Value: 2, IsSet: true},
					MemoryInMB: types.NullUint64{Value: 100, IsSet: true},
					DiskInMB:   types.NullUint64{Value: 200, IsSet: true},
				}))

				Expect(fakeCloudControllerClient.GetApplicationProcessByTypeCallCount()).To(Equal(1))
				appGUIDArg, processTypeArg := fakeCloudControllerClient.GetApplicationProcessByTypeArgsForCall(0)
				Expect(appGUIDArg).To(Equal("some-app-guid"))
				Expect(processTypeArg).To(Equal("web"))
			})
		})

		Context("when CC returns an error", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("some-error")
				fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
					ccv3.Process{},
					ccv3.Warnings{"get-process-warning"},
					expectedErr,
				)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := actor.GetProcessByApplicationAndProcessType("some-app-guid", "web")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("get-process-warning"))
			})
		})

		Context("when CC returns a ProcessNotFoundError", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
					ccv3.Process{},
					ccv3.Warnings{"get-process-warning"},
					ccerror.ProcessNotFoundError{},
				)
			})

			It("returns the ProcessNotFoundError and all warnings", func() {
				_, warnings, err := actor.GetProcessByApplicationAndProcessType("some-app-guid", "web")
				Expect(err).To(Equal(ProcessNotFoundError{ProcessType: "web"}))
				Expect(warnings).To(ConsistOf("get-process-warning"))
			})
		})
	})
})
