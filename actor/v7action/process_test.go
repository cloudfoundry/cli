package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
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
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("GetProcessByTypeAndApplication", func() {
		var (
			processType string
			appGUID     string

			process  Process
			warnings Warnings
			err      error
		)

		BeforeEach(func() {
			processType = "web"
			appGUID = "some-app-guid"
		})

		JustBeforeEach(func() {
			process, warnings, err = actor.GetProcessByTypeAndApplication(processType, appGUID)
		})

		When("getting the application process is succesful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
					ccv3.Process{
						GUID: "some-process-guid",
					},
					ccv3.Warnings{"some-process-warning"},
					nil,
				)
			})

			It("returns the process and warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-process-warning"))
				Expect(process).To(Equal(Process{
					GUID: "some-process-guid",
				}))

				Expect(fakeCloudControllerClient.GetApplicationProcessByTypeCallCount()).To(Equal(1))
				passedAppGUID, passedProcessType := fakeCloudControllerClient.GetApplicationProcessByTypeArgsForCall(0)
				Expect(passedAppGUID).To(Equal(appGUID))
				Expect(passedProcessType).To(Equal(processType))
			})
		})

		When("getting application process by type returns an error", func() {
			var expectedErr error

			When("the api returns a ProcessNotFoundError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
						ccv3.Process{},
						ccv3.Warnings{"some-process-warning"},
						ccerror.ProcessNotFoundError{},
					)
				})

				It("returns a ProcessNotFoundError and all warnings", func() {
					Expect(err).To(Equal(actionerror.ProcessNotFoundError{ProcessType: "web"}))
					Expect(warnings).To(Equal(Warnings{"some-process-warning"}))
				})
			})

			Context("generic error", func() {
				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
						ccv3.Process{},
						ccv3.Warnings{"some-process-warning"},
						expectedErr,
					)
				})

				It("returns the error and warnings", func() {
					Expect(err).To(Equal(expectedErr))
					Expect(warnings).To(Equal(Warnings{"some-process-warning"}))
				})
			})
		})
	})

	Describe("ScaleProcessByApplication", func() {
		var (
			passedProcess Process
			warnings      Warnings
			executeErr    error
		)

		BeforeEach(func() {
			passedProcess = Process{
				Type:       constant.ProcessTypeWeb,
				Instances:  types.NullInt{Value: 2, IsSet: true},
				MemoryInMB: types.NullUint64{Value: 100, IsSet: true},
				DiskInMB:   types.NullUint64{Value: 200, IsSet: true},
			}
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.ScaleProcessByApplication("some-app-guid", passedProcess)
		})

		When("no errors are encountered scaling the application process", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationProcessScaleReturns(
					ccv3.Process{GUID: "some-process-guid"},
					ccv3.Warnings{"scale-process-warning"},
					nil)
			})

			It("scales correct process", func() {
				Expect(executeErr).ToNot(HaveOccurred())
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

		When("an error is encountered scaling the application process", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("scale process error")
				fakeCloudControllerClient.CreateApplicationProcessScaleReturns(
					ccv3.Process{GUID: "some-process-guid"},
					ccv3.Warnings{"scale-process-warning"},
					expectedErr)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("scale-process-warning"))
			})
		})

		When("a ProcessNotFoundError error is encountered scaling the application process", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationProcessScaleReturns(
					ccv3.Process{GUID: "some-process-guid"},
					ccv3.Warnings{"scale-process-warning"},
					ccerror.ProcessNotFoundError{},
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(Equal(actionerror.ProcessNotFoundError{ProcessType: constant.ProcessTypeWeb}))
				Expect(warnings).To(ConsistOf("scale-process-warning"))
			})
		})
	})
})
