package shared_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/command/v3/shared/sharedfakes"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("app summary displayer", func() {
	var (
		appSummaryDisplayer AppSummaryDisplayer
		input               *Buffer
		output              *Buffer
		testUI              *ui.UI
		fakeConfig          *commandfakes.FakeConfig
		fakeActor           *sharedfakes.FakeV3AppSummaryActor
		appName             string
		executeErr          error
	)

	BeforeEach(func() {
		input = NewBuffer()
		output = NewBuffer()
		testUI = ui.NewTestUI(input, output, NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(sharedfakes.FakeV3AppSummaryActor)
		appName = "some-app"

		appSummaryDisplayer = AppSummaryDisplayer{
			UI:              testUI,
			Config:          fakeConfig,
			Actor:           fakeActor,
			V2AppRouteActor: nil,
			AppName:         appName,
		}

		fakeConfig.TargetedSpaceReturns(configv3.Space{
			GUID: "some-space-guid",
			Name: "some-space"})
	})

	JustBeforeEach(func() {
		executeErr = appSummaryDisplayer.DisplayAppProcessInfo()
	})

	Describe("DisplayAppProcessInfo", func() {
		Context("when getting the app summary succeeds", func() {
			BeforeEach(func() {
				appSummary := v3action.ApplicationSummary{
					ProcessSummaries: v3action.ProcessSummaries{
						{
							Process: v3action.Process{
								Type:       "console",
								MemoryInMB: types.NullUint64{Value: 16, IsSet: true},
								DiskInMB:   types.NullUint64{Value: 512, IsSet: true},
							},
							InstanceDetails: []v3action.ProcessInstance{
								v3action.ProcessInstance{
									Index:       0,
									State:       constant.ProcessInstanceRunning,
									MemoryUsage: 1000000,
									DiskUsage:   1000000,
									MemoryQuota: 33554432,
									DiskQuota:   8000000,
									Uptime:      int(time.Now().Sub(time.Unix(167572800, 0)).Seconds()),
								},
							},
						},
						{
							Process: v3action.Process{
								Type:       constant.ProcessTypeWeb,
								MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
								DiskInMB:   types.NullUint64{Value: 1024, IsSet: true},
							},
							InstanceDetails: []v3action.ProcessInstance{
								v3action.ProcessInstance{
									Index:       0,
									State:       constant.ProcessInstanceRunning,
									MemoryUsage: 1000000,
									DiskUsage:   1000000,
									MemoryQuota: 33554432,
									DiskQuota:   2000000,
									Uptime:      int(time.Now().Sub(time.Unix(267321600, 0)).Seconds()),
								},
								v3action.ProcessInstance{
									Index:       1,
									State:       constant.ProcessInstanceRunning,
									MemoryUsage: 2000000,
									DiskUsage:   2000000,
									MemoryQuota: 33554432,
									DiskQuota:   4000000,
									Uptime:      int(time.Now().Sub(time.Unix(330480000, 0)).Seconds()),
								},
								v3action.ProcessInstance{
									Index:       2,
									State:       constant.ProcessInstanceRunning,
									MemoryUsage: 3000000,
									DiskUsage:   3000000,
									MemoryQuota: 33554432,
									DiskQuota:   6000000,
									Uptime:      int(time.Now().Sub(time.Unix(1277164800, 0)).Seconds()),
								},
							},
						},
					},
				}

				fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
					appSummary,
					v3action.Warnings{"get-app-summary-warning"},
					nil)
			})

			It("lists information for each of the processes", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				processTable := helpers.ParseV3AppProcessTable(output.Contents())
				Expect(len(processTable.Processes)).To(Equal(2))

				webProcessSummary := processTable.Processes[0]
				Expect(webProcessSummary.Title).To(Equal("web:3/3"))

				Expect(webProcessSummary.Instances[0].Memory).To(Equal("976.6K of 32M"))
				Expect(webProcessSummary.Instances[0].Disk).To(Equal("976.6K of 1.9M"))
				Expect(webProcessSummary.Instances[0].CPU).To(Equal("0.0%"))

				Expect(webProcessSummary.Instances[1].Memory).To(Equal("1.9M of 32M"))
				Expect(webProcessSummary.Instances[1].Disk).To(Equal("1.9M of 3.8M"))
				Expect(webProcessSummary.Instances[1].CPU).To(Equal("0.0%"))

				Expect(webProcessSummary.Instances[2].Memory).To(Equal("2.9M of 32M"))
				Expect(webProcessSummary.Instances[2].Disk).To(Equal("2.9M of 5.7M"))
				Expect(webProcessSummary.Instances[2].CPU).To(Equal("0.0%"))

				consoleProcessSummary := processTable.Processes[1]
				Expect(consoleProcessSummary.Title).To(Equal("console:1/1"))

				Expect(consoleProcessSummary.Instances[0].Memory).To(Equal("976.6K of 32M"))
				Expect(consoleProcessSummary.Instances[0].Disk).To(Equal("976.6K of 7.6M"))
				Expect(consoleProcessSummary.Instances[0].CPU).To(Equal("0.0%"))

				Expect(testUI.Err).To(Say("get-app-summary-warning"))

				Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
				passedAppName, spaceName := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
				Expect(passedAppName).To(Equal("some-app"))
				Expect(spaceName).To(Equal("some-space-guid"))
			})
		})

		Context("when getting the app summary fails", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
					v3action.ApplicationSummary{},
					v3action.Warnings{"get-app-summary-warning"},
					errors.New("some-app-summary-error"),
				)
			})

			It("returns the error and displays all warnings", func() {
				Expect(executeErr).To(MatchError("some-app-summary-error"))
				Expect(testUI.Err).To(Say("get-app-summary-warning"))
				Expect(output.Contents()).To(HaveLen(0))
			})
		})
	})
})
