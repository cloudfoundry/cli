package shared_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/command/v6/shared/sharedfakes"
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
		output              *Buffer
		testUI              *ui.UI
		fakeConfig          *commandfakes.FakeConfig
		fakeV2Actor         *sharedfakes.FakeV2AppActor
		fakeActor           *sharedfakes.FakeV3AppSummaryActor
		appName             string
		executeErr          error
	)

	BeforeEach(func() {
		output = NewBuffer()
		testUI = ui.NewTestUI(nil, output, NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(sharedfakes.FakeV3AppSummaryActor)
		fakeV2Actor = new(sharedfakes.FakeV2AppActor)
		appName = "some-app"

		appSummaryDisplayer = AppSummaryDisplayer{
			UI:         testUI,
			Config:     fakeConfig,
			Actor:      fakeActor,
			V2AppActor: fakeV2Actor,
			AppName:    appName,
		}

		fakeConfig.TargetedSpaceReturns(configv3.Space{
			GUID: "some-space-guid",
			Name: "some-space"})
	})

	Describe("DisplayAppInfo", func() {
		JustBeforeEach(func() {
			executeErr = appSummaryDisplayer.DisplayAppInfo()
		})

		When("getting the app summary succeeds", func() {
			When("the app has instances", func() {
				BeforeEach(func() {
					appSummary := v3action.ApplicationSummary{
						Application: v3action.Application{
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
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
										Uptime:      time.Now().Sub(time.Unix(167572800, 0)),
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
										Uptime:      time.Now().Sub(time.Unix(267321600, 0)),
									},
									v3action.ProcessInstance{
										Index:       1,
										State:       constant.ProcessInstanceRunning,
										MemoryUsage: 2000000,
										DiskUsage:   2000000,
										MemoryQuota: 33554432,
										DiskQuota:   4000000,
										Uptime:      time.Now().Sub(time.Unix(330480000, 0)),
									},
									v3action.ProcessInstance{
										Index:       2,
										State:       constant.ProcessInstanceRunning,
										MemoryUsage: 3000000,
										DiskUsage:   3000000,
										MemoryQuota: 33554432,
										DiskQuota:   6000000,
										Uptime:      time.Now().Sub(time.Unix(1277164800, 0)),
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

				When("the isolation segment name is provided", func() {
					var isolationSegmentName string
					BeforeEach(func() {
						isolationSegmentName = "potato beans"
						fakeV2Actor.GetApplicationInstancesWithStatsByApplicationReturns(
							[]v2action.ApplicationInstanceWithStats{
								{IsolationSegment: isolationSegmentName},
							},
							v2action.Warnings{"some-instance-stats-warning"}, nil)
					})

					It("should output the isolation segment name", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say(`isolation segment:\s+%s`, isolationSegmentName))

						Expect(testUI.Err).To(Say("get-app-summary-warning"))
						Expect(testUI.Err).To(Say("some-instance-stats-warning"))

						Expect(fakeV2Actor.GetApplicationInstancesWithStatsByApplicationCallCount()).To(Equal(1))
						Expect(fakeV2Actor.GetApplicationInstancesWithStatsByApplicationArgsForCall(0)).To(Equal("some-app-guid"))
					})
				})

				When("the isolation segment name is missing", func() {
					BeforeEach(func() {
						fakeV2Actor.GetApplicationInstancesWithStatsByApplicationReturns(
							[]v2action.ApplicationInstanceWithStats{{}},
							v2action.Warnings{"some-instance-stats-warning"}, nil)
					})

					It("should not output the isolation segment row", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).ToNot(Say("isolation segment:"))

						Expect(testUI.Err).To(Say("get-app-summary-warning"))
						Expect(testUI.Err).To(Say("some-instance-stats-warning"))
					})
				})

				When("getting the isolation segment information errors", func() {
					When("a random error is returned", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("knaslfnasldfnasdfnasdkj")
							fakeV2Actor.GetApplicationInstancesWithStatsByApplicationReturns(nil, v2action.Warnings{"some-instance-stats-warning"}, expectedErr)
						})

						It("displays the warnings and returns an error", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("get-app-summary-warning"))
							Expect(testUI.Err).To(Say("some-instance-stats-warning"))
						})
					})

					When("a random error is returned", func() {
						BeforeEach(func() {
							fakeV2Actor.GetApplicationInstancesWithStatsByApplicationReturns(nil, v2action.Warnings{"some-instance-stats-warning"}, ccerror.ResourceNotFoundError{})
						})

						It("displays the warnings and continues", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Out).ToNot(Say("isolation segment:"))

							Expect(testUI.Err).To(Say("get-app-summary-warning"))
							Expect(testUI.Err).To(Say("some-instance-stats-warning"))
						})
					})
				})

				It("lists information for each of the processes", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					processTable := helpers.ParseV3AppProcessTable(output.Contents())
					Expect(len(processTable.Processes)).To(Equal(2))

					webProcessSummary := processTable.Processes[0]
					Expect(webProcessSummary.Type).To(Equal("web"))
					Expect(webProcessSummary.InstanceCount).To(Equal("3/3"))
					Expect(webProcessSummary.MemUsage).To(Equal("32M"))

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
					Expect(consoleProcessSummary.Type).To(Equal("console"))
					Expect(consoleProcessSummary.InstanceCount).To(Equal("1/1"))
					Expect(consoleProcessSummary.MemUsage).To(Equal("16M"))

					Expect(consoleProcessSummary.Instances[0].Memory).To(Equal("976.6K of 32M"))
					Expect(consoleProcessSummary.Instances[0].Disk).To(Equal("976.6K of 7.6M"))
					Expect(consoleProcessSummary.Instances[0].CPU).To(Equal("0.0%"))

					Expect(testUI.Err).To(Say("get-app-summary-warning"))

					Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
					passedAppName, spaceName, withObfuscatedValues := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
					Expect(passedAppName).To(Equal("some-app"))
					Expect(spaceName).To(Equal("some-space-guid"))
					Expect(withObfuscatedValues).To(BeFalse())
				})
			})

			When("the app has no instances", func() {
				BeforeEach(func() {
					appSummary := v3action.ApplicationSummary{
						Application: v3action.Application{
							GUID: "some-app-guid",
						},
					}

					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
						appSummary,
						v3action.Warnings{"get-app-summary-warning"},
						nil)
					fakeV2Actor.GetApplicationInstancesWithStatsByApplicationReturns(nil, v2action.Warnings{"some-instance-stats-warning"}, nil)
				})

				It("should not output isolation segment header", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).ToNot(Say("isolation segment:"))

					Expect(fakeV2Actor.GetApplicationInstancesWithStatsByApplicationCallCount()).To(Equal(0))
				})
			})

			When("the app is stopped", func() {
				BeforeEach(func() {
					appSummary := v3action.ApplicationSummary{
						Application: v3action.Application{
							GUID:  "some-app-guid",
							State: constant.ApplicationStopped,
						},
						ProcessSummaries: v3action.ProcessSummaries{{}},
					}

					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
						appSummary,
						v3action.Warnings{"get-app-summary-warning"},
						nil)
					fakeV2Actor.GetApplicationInstancesWithStatsByApplicationReturns(nil, v2action.Warnings{"some-instance-stats-warning"}, nil)
				})

				It("should not output isolation segment header", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).ToNot(Say("isolation segment:"))

					Expect(fakeV2Actor.GetApplicationInstancesWithStatsByApplicationCallCount()).To(Equal(0))
				})
			})
		})

		When("getting the app summary fails", func() {
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
