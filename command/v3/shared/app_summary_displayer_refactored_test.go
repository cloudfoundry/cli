package shared_test

import (
	"time"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("app summary displayer", func() {
	var (
		appSummaryDisplayer *AppSummaryDisplayer2
		output              *Buffer
		testUI              *ui.UI
	)

	BeforeEach(func() {
		output = NewBuffer()
		testUI = ui.NewTestUI(nil, output, NewBuffer())

		appSummaryDisplayer = NewAppSummaryDisplayer2(testUI)
	})

	Describe("AppDisplay", func() {
		var (
			summary v2v3action.ApplicationSummary
		)

		JustBeforeEach(func() {
			appSummaryDisplayer.AppDisplay(summary)
		})

		Context("when the app has instances", func() {
			Context("when the process instances are running", func() {
				BeforeEach(func() {
					summary = v2v3action.ApplicationSummary{
						ApplicationSummary: v3action.ApplicationSummary{
							Application: v3action.Application{
								GUID:  "some-app-guid",
								State: constant.ApplicationStarted,
							},
							ProcessSummaries: v3action.ProcessSummaries{
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
							},
						},
					}
				})

				It("lists information for each of the processes", func() {
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
				})
			})

			Context("when only one process instance is running", func() {
				BeforeEach(func() {
					summary = v2v3action.ApplicationSummary{
						ApplicationSummary: v3action.ApplicationSummary{
							Application: v3action.Application{
								GUID:  "some-app-guid",
								State: constant.ApplicationStarted,
							},
							ProcessSummaries: v3action.ProcessSummaries{
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
									},
								},
								{
									Process: v3action.Process{
										Type:       "console",
										MemoryInMB: types.NullUint64{Value: 16, IsSet: true},
										DiskInMB:   types.NullUint64{Value: 512, IsSet: true},
									},
								},
							},
						},
					}
				})

				It("lists information for each of the processes", func() {
					Expect(testUI.Out).To(Say("type:\\s+web"))
					Expect(testUI.Out).To(Say(`state\s+since\s+cpu\s+memory\s+disk`))
					Expect(testUI.Out).To(Say("type:\\s+console"))
					Expect(testUI.Out).To(Say("There are no running instances of this process."))
				})
			})

			Context("when all the instances in all processes are down", func() {
				BeforeEach(func() {
					summary = v2v3action.ApplicationSummary{
						ApplicationSummary: v3action.ApplicationSummary{
							ProcessSummaries: []v3action.ProcessSummary{
								{
									Process: v3action.Process{
										Type:       constant.ProcessTypeWeb,
										MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
									},
									InstanceDetails: []v3action.ProcessInstance{{State: constant.ProcessInstanceDown}},
								},
								{
									Process: v3action.Process{
										Type:       "console",
										MemoryInMB: types.NullUint64{Value: 128, IsSet: true},
									},
									InstanceDetails: []v3action.ProcessInstance{{State: constant.ProcessInstanceDown}},
								},
								{
									Process: v3action.Process{
										Type:       "worker",
										MemoryInMB: types.NullUint64{Value: 64, IsSet: true},
									},
									InstanceDetails: []v3action.ProcessInstance{{State: constant.ProcessInstanceDown}},
								},
							},
						},
					}
				})

				It("says no instances are running", func() {
					Expect(testUI.Out).To(Say("type:\\s+web"))
					Expect(testUI.Out).To(Say("There are no running instances of this process."))
					Expect(testUI.Out).To(Say("type:\\s+console"))
					Expect(testUI.Out).To(Say("There are no running instances of this process."))
					Expect(testUI.Out).To(Say("type:\\s+worker"))
					Expect(testUI.Out).To(Say("There are no running instances of this process."))
				})

				It("does not display the instance table", func() {
					Expect(testUI.Out).NotTo(Say(`state\s+since\s+cpu\s+memory\s+disk`))
				})
			})
		})

		Context("when the app has no instances", func() {
			BeforeEach(func() {
				summary = v2v3action.ApplicationSummary{
					ApplicationSummary: v3action.ApplicationSummary{
						Application: v3action.Application{
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
						ProcessSummaries: v3action.ProcessSummaries{
							{
								Process: v3action.Process{
									Type:       constant.ProcessTypeWeb,
									MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
									DiskInMB:   types.NullUint64{Value: 1024, IsSet: true},
								},
							},
							{
								Process: v3action.Process{
									Type:       "console",
									MemoryInMB: types.NullUint64{Value: 16, IsSet: true},
									DiskInMB:   types.NullUint64{Value: 512, IsSet: true},
								},
							},
						},
					},
				}
			})

			It("lists information for each of the processes", func() {
				Expect(testUI.Out).To(Say("type:\\s+web"))
				Expect(testUI.Out).To(Say("instances:\\s+0/0"))
				Expect(testUI.Out).To(Say("memory usage:\\s+32M"))
				Expect(testUI.Out).To(Say("There are no running instances of this process."))

				Expect(testUI.Out).To(Say("type:\\s+console"))
				Expect(testUI.Out).To(Say("instances:\\s+0/0"))
				Expect(testUI.Out).To(Say("memory usage:\\s+16M"))
				Expect(testUI.Out).To(Say("There are no running instances of this process."))
			})

			It("does not display the instance table", func() {
				Expect(testUI.Out).NotTo(Say(`state\s+since\s+cpu\s+memory\s+disk`))
			})
		})

		Describe("isolation segments", func() {
			Context("when the isolation segment name is provided", func() {
				var isolationSegmentName string
				BeforeEach(func() {
					isolationSegmentName = "potato beans"
					summary.ApplicationInstanceWithStats =
						[]v2action.ApplicationInstanceWithStats{
							{IsolationSegment: isolationSegmentName},
						}
				})

				It("should output the isolation segment name", func() {
					Expect(testUI.Out).To(Say("isolation segment:\\s+%s", isolationSegmentName))
				})
			})

			Context("when the application summary has no isolation segment information", func() {
				BeforeEach(func() {
					summary = v2v3action.ApplicationSummary{
						ApplicationSummary: v3action.ApplicationSummary{
							Application: v3action.Application{
								GUID:  "some-app-guid",
								State: constant.ApplicationStopped,
							},
						},
					}
				})

				It("should not output isolation segment header", func() {
					Expect(testUI.Out).ToNot(Say("isolation segment:"))
				})
			})
		})

		Describe("last upload time", func() {
			Context("when the application has a last uploaded time", func() {
				BeforeEach(func() {
					summary.CurrentDroplet.CreatedAt = "2006-01-02T15:04:05Z07:00"
				})

				It("displays the uploaded time", func() {
					Expect(testUI.Out).To(Say("last uploaded:\\s+%s", "Sun 31 Dec 16:07:02 LMT 0000"))
				})
			})

			Context("when the application does not have a last uploaded time", func() {
				BeforeEach(func() {
					summary.CurrentDroplet.CreatedAt = ""
				})

				It("leaves last uploaded blank", func() {
					Expect(testUI.Out).To(Say("(?m)last uploaded:\\s*\n"))
				})
			})
		})

		Context("when the application has routes", func() {
			BeforeEach(func() {
				summary.Routes = []v2action.Route{
					{Host: "route1", Domain: v2action.Domain{Name: "example.com"}},
					{Host: "route2", Domain: v2action.Domain{Name: "example.com"}},
				}
			})

			It("displays routes", func() {
				Expect(testUI.Out).To(Say("routes:\\s+%s, %s", "route1.example.com", "route2.example.com"))
			})
		})

		Context("when the application has a stack", func() {
			BeforeEach(func() {
				summary.CurrentDroplet.Stack = "some-stack"
			})

			It("displays stack", func() {
				Expect(testUI.Out).To(Say("stack:\\s+some-stack"))
			})
		})

		Context("when the application is a docker app", func() {
			BeforeEach(func() {
				summary = v2v3action.ApplicationSummary{
					ApplicationSummary: v3action.ApplicationSummary{
						Application: v3action.Application{
							GUID:          "some-guid",
							Name:          "some-app",
							State:         constant.ApplicationStarted,
							LifecycleType: constant.AppLifecycleTypeDocker,
						},
						CurrentDroplet: v3action.Droplet{
							Image: "docker/some-image",
						},
					},
				}
			})

			It("displays the app information", func() {
				Expect(testUI.Out).To(Say("name:\\s+some-app"))
				Expect(testUI.Out).To(Say("requested state:\\s+started"))
				Expect(testUI.Out).To(Say("routes:\\s+\\n"))
				Expect(testUI.Out).To(Say("stack:\\s+\\n"))
				Expect(testUI.Out).To(Say("(?m)docker image:\\s+docker/some-image$\\n"))
			})
		})

		Context("when the application is a buildpack app", func() {
			BeforeEach(func() {
				summary = v2v3action.ApplicationSummary{
					ApplicationSummary: v3action.ApplicationSummary{
						CurrentDroplet: v3action.Droplet{
							Stack: "cflinuxfs2",
							Buildpacks: []v3action.Buildpack{
								{
									Name:         "ruby_buildpack",
									DetectOutput: "some-detect-output",
								},
								{
									Name:         "some-buildpack",
									DetectOutput: "",
								},
							},
						},
					},
				}
			})

			It("displays stack and buildpacks", func() {
				Expect(testUI.Out).To(Say("stack:\\s+cflinuxfs2"))
				Expect(testUI.Out).To(Say("buildpacks:\\s+some-detect-output, some-buildpack"))
			})
		})

		Context("when app has no processes", func() {
		})

	})
})
