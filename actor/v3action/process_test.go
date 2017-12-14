package v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
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
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("ScaleProcessByApplication", func() {
		var passedProcess Process

		BeforeEach(func() {
			passedProcess = Process{
				Type:       constant.ProcessTypeWeb,
				Instances:  types.NullInt{Value: 2, IsSet: true},
				MemoryInMB: types.NullUint64{Value: 100, IsSet: true},
				DiskInMB:   types.NullUint64{Value: 200, IsSet: true},
			}
		})

		Context("when no errors are encountered scaling the application process", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationProcessScaleReturns(
					ccv3.Process{GUID: "some-process-guid"},
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
					Type:       constant.ProcessTypeWeb,
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
					ccv3.Process{GUID: "some-process-guid"},
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
					ccv3.Process{GUID: "some-process-guid"},
					ccv3.Warnings{"scale-process-warning"},
					ccerror.ProcessNotFoundError{},
				)
			})

			It("returns the error and all warnings", func() {
				warnings, err := actor.ScaleProcessByApplication("some-app-guid", passedProcess)
				Expect(err).To(Equal(actionerror.ProcessNotFoundError{ProcessType: constant.ProcessTypeWeb}))
				Expect(warnings).To(ConsistOf("scale-process-warning"))
			})
		})
	})
})
