package shared_test

import (
	"time"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("app summary displayer", func() {
	var (
		appSummaryDisplayer *AppSummaryDisplayer
		output              *Buffer
		testUI              *ui.UI
	)

	BeforeEach(func() {
		output = NewBuffer()
		testUI = ui.NewTestUI(nil, output, NewBuffer())

		appSummaryDisplayer = NewAppSummaryDisplayer(testUI)
	})

	Describe("AppDisplay", func() {
		var (
			summary             v7action.ApplicationSummary
			displayStartCommand bool
		)

		JustBeforeEach(func() {
			appSummaryDisplayer.AppDisplay(summary, displayStartCommand)
		})

		When("the app has instances", func() {
			When("the process instances are running", func() {
				var uptime time.Duration

				BeforeEach(func() {
					uptime = time.Now().Sub(time.Unix(267321600, 0))
					summary = v7action.ApplicationSummary{
						Application: v7action.Application{
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
						ProcessSummaries: v7action.ProcessSummaries{
							{
								Process: v7action.Process{
									Type:       constant.ProcessTypeWeb,
									MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
									DiskInMB:   types.NullUint64{Value: 1024, IsSet: true},
								},
								InstanceDetails: []v7action.ProcessInstance{
									v7action.ProcessInstance{
										Index:       0,
										State:       constant.ProcessInstanceRunning,
										MemoryUsage: 1000000,
										DiskUsage:   1000000,
										MemoryQuota: 33554432,
										DiskQuota:   2000000,
										Uptime:      int(uptime.Seconds()),
										Details:     "Some Details 1",
									},
									v7action.ProcessInstance{
										Index:       1,
										State:       constant.ProcessInstanceRunning,
										MemoryUsage: 2000000,
										DiskUsage:   2000000,
										MemoryQuota: 33554432,
										DiskQuota:   4000000,
										Uptime:      int(time.Now().Sub(time.Unix(330480000, 0)).Seconds()),
										Details:     "Some Details 2",
									},
									v7action.ProcessInstance{
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
								Process: v7action.Process{
									Type:       "console",
									MemoryInMB: types.NullUint64{Value: 16, IsSet: true},
									DiskInMB:   types.NullUint64{Value: 512, IsSet: true},
								},
								InstanceDetails: []v7action.ProcessInstance{
									v7action.ProcessInstance{
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
					Expect(webProcessSummary.Instances[0].Since).To(MatchRegexp(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`))
					Expect(time.Parse(time.RFC3339, webProcessSummary.Instances[0].Since)).To(BeTemporally("~", time.Now().Add(-uptime), 2*time.Second))
					Expect(webProcessSummary.Instances[0].Disk).To(Equal("976.6K of 1.9M"))
					Expect(webProcessSummary.Instances[0].CPU).To(Equal("0.0%"))
					Expect(webProcessSummary.Instances[0].Details).To(Equal("Some Details 1"))

					Expect(webProcessSummary.Instances[1].Memory).To(Equal("1.9M of 32M"))
					Expect(webProcessSummary.Instances[1].Disk).To(Equal("1.9M of 3.8M"))
					Expect(webProcessSummary.Instances[1].CPU).To(Equal("0.0%"))
					Expect(webProcessSummary.Instances[1].Details).To(Equal("Some Details 2"))

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

			When("some processes have > 0 instances and others have 0 instances", func() {
				BeforeEach(func() {
					summary = v7action.ApplicationSummary{
						Application: v7action.Application{
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
						ProcessSummaries: v7action.ProcessSummaries{
							{
								Process: v7action.Process{
									Type:       constant.ProcessTypeWeb,
									MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
									DiskInMB:   types.NullUint64{Value: 1024, IsSet: true},
								},
								InstanceDetails: []v7action.ProcessInstance{
									v7action.ProcessInstance{
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
								Process: v7action.Process{
									Type:       "console",
									MemoryInMB: types.NullUint64{Value: 16, IsSet: true},
									DiskInMB:   types.NullUint64{Value: 512, IsSet: true},
								},
							},
						},
					}
				})

				It("lists instance stats for process types that have > 0 instances", func() {
					Expect(testUI.Out).To(Say(`type:\s+web`))
					Expect(testUI.Out).To(Say(`state\s+since\s+cpu\s+memory\s+disk\s+details`))
				})

				It("does not show the instance stats table for process types with 0 instances", func() {
					Expect(testUI.Out).To(Say(`type:\s+console`))
					Expect(testUI.Out).To(Say("There are no running instances of this process."))
				})
			})

			When("all the instances for a process are down (but scaled to > 0 instances)", func() {
				BeforeEach(func() {
					summary = v7action.ApplicationSummary{

						ProcessSummaries: []v7action.ProcessSummary{
							{
								Process: v7action.Process{
									Type:       constant.ProcessTypeWeb,
									MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
								},
								InstanceDetails: []v7action.ProcessInstance{{State: constant.ProcessInstanceDown}},
							},
						},
					}
				})

				It("displays the instances table", func() {
					Expect(testUI.Out).To(Say(`type:\s+web`))
					Expect(testUI.Out).To(Say(`state\s+since\s+cpu\s+memory\s+disk\s+details`))
				})
			})

			Describe("start command", func() {
				BeforeEach(func() {
					summary = v7action.ApplicationSummary{
						Application: v7action.Application{
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
						ProcessSummaries: v7action.ProcessSummaries{
							{
								Process: v7action.Process{
									Type:    constant.ProcessTypeWeb,
									Command: *types.NewFilteredString("some-command-1"),
								},
							},
							{
								Process: v7action.Process{
									Type:    "console",
									Command: *types.NewFilteredString("some-command-2"),
								},
							},
							{
								Process: v7action.Process{
									Type: "random",
								},
							},
						},
					}
				})

				When("displayStartCommand is true", func() {
					BeforeEach(func() {
						displayStartCommand = true
					})

					It("displays the non-empty start command for each process", func() {
						Expect(testUI.Out).To(Say(`type:\s+web`))
						Expect(testUI.Out).To(Say(`start command:\s+some-command-1`))

						Expect(testUI.Out).To(Say(`type:\s+console`))
						Expect(testUI.Out).To(Say(`start command:\s+some-command-2`))

						Expect(testUI.Out).To(Say(`type:\s+random`))
						Expect(testUI.Out).ToNot(Say("start command:"))
					})
				})

				When("displayStartCommand is false", func() {
					BeforeEach(func() {
						displayStartCommand = false
					})

					It("hides the start command", func() {
						Expect(testUI.Out).ToNot(Say("start command:"))
					})
				})
			})
		})

		When("the app has no instances", func() {
			BeforeEach(func() {
				summary = v7action.ApplicationSummary{
					Application: v7action.Application{
						GUID:  "some-app-guid",
						State: constant.ApplicationStarted,
					},
					ProcessSummaries: v7action.ProcessSummaries{
						{
							Process: v7action.Process{
								Type:       constant.ProcessTypeWeb,
								MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
								DiskInMB:   types.NullUint64{Value: 1024, IsSet: true},
							},
						},
						{
							Process: v7action.Process{
								Type:       "console",
								MemoryInMB: types.NullUint64{Value: 16, IsSet: true},
								DiskInMB:   types.NullUint64{Value: 512, IsSet: true},
							},
						},
					},
				}
			})

			It("lists information for each of the processes", func() {
				Expect(testUI.Out).To(Say(`type:\s+web`))
				Expect(testUI.Out).To(Say(`instances:\s+0/0`))
				Expect(testUI.Out).To(Say(`memory usage:\s+32M`))
				Expect(testUI.Out).To(Say("There are no running instances of this process."))

				Expect(testUI.Out).To(Say(`type:\s+console`))
				Expect(testUI.Out).To(Say(`instances:\s+0/0`))
				Expect(testUI.Out).To(Say(`memory usage:\s+16M`))
				Expect(testUI.Out).To(Say("There are no running instances of this process."))
			})

			It("does not display the instance table", func() {
				Expect(testUI.Out).NotTo(Say(`state\s+since\s+cpu\s+memory\s+disk\s+details`))
			})
		})

		When("the app is stopped", func() {
			BeforeEach(func() {
				summary = v7action.ApplicationSummary{
					Application: v7action.Application{
						GUID:  "some-app-guid",
						State: constant.ApplicationStopped,
					},
					ProcessSummaries: v7action.ProcessSummaries{
						{
							Process: v7action.Process{
								Type: constant.ProcessTypeWeb,
							},
						},
						{
							Process: v7action.Process{
								Type: "console",
							},
						},
					},
				}
			})

			It("lists information for each of the processes", func() {
				Expect(testUI.Out).To(Say(`type:\s+web`))
				Expect(testUI.Out).To(Say("There are no running instances of this process."))

				Expect(testUI.Out).To(Say(`type:\s+console`))
				Expect(testUI.Out).To(Say("There are no running instances of this process."))
			})

			It("does not display the instance table", func() {
				Expect(testUI.Out).NotTo(Say(`state\s+since\s+cpu\s+memory\s+disk\s+details`))
			})
		})

		Describe("isolation segments", func() {
			When("the isolation segment name is provided", func() {
				var isolationSegmentName string
				BeforeEach(func() {
					isolationSegmentName = "potato beans"
					summary.ProcessSummaries = v7action.ProcessSummaries{
						v7action.ProcessSummary{
							InstanceDetails: []v7action.ProcessInstance{
								{IsolationSegment: isolationSegmentName},
							},
						},
					}
				})

				It("should output the isolation segment name", func() {
					Expect(testUI.Out).To(Say(`isolation segment:\s+%s`, isolationSegmentName))
				})
			})

			When("the application summary has no isolation segment information", func() {
				BeforeEach(func() {
					summary = v7action.ApplicationSummary{
						Application: v7action.Application{
							GUID:  "some-app-guid",
							State: constant.ApplicationStopped,
						},
					}
				})

				It("should not output isolation segment header", func() {
					Expect(testUI.Out).ToNot(Say("isolation segment:"))
				})
			})
		})

		Describe("last upload time", func() {
			When("the application has a last uploaded time", func() {
				var createdTime string

				BeforeEach(func() {
					createdTime = "2006-01-02T15:04:05-07:00"
					summary.CurrentDroplet.CreatedAt = createdTime
				})

				It("displays the uploaded time", func() {
					t, err := time.Parse(time.RFC3339, createdTime)
					Expect(err).To(Not(HaveOccurred()))

					time := t.Local().Format("Mon 02 Jan 15:04:05 MST 2006")
					Expect(testUI.Out).To(Say(`last uploaded:\s+%s`, time))
				})
			})

			When("the application does not have a last uploaded time", func() {
				BeforeEach(func() {
					summary.CurrentDroplet.CreatedAt = ""
				})

				It("leaves last uploaded blank", func() {
					Expect(testUI.Out).To(Say(`(?m)last uploaded:\s*\n`))
				})
			})
		})

		When("the application has routes", func() {
			BeforeEach(func() {
				summary.Routes = []v2action.Route{
					{Host: "route1", Domain: v2action.Domain{Name: "example.com"}},
					{Host: "route2", Domain: v2action.Domain{Name: "example.com"}},
				}
			})

			It("displays routes", func() {
				Expect(testUI.Out).To(Say(`routes:\s+%s, %s`, "route1.example.com", "route2.example.com"))
			})
		})

		When("the application has a stack", func() {
			BeforeEach(func() {
				summary.CurrentDroplet.Stack = "some-stack"
			})

			It("displays stack", func() {
				Expect(testUI.Out).To(Say(`stack:\s+some-stack`))
			})
		})

		When("the application is a docker app", func() {
			BeforeEach(func() {
				summary = v7action.ApplicationSummary{
					Application: v7action.Application{
						GUID:          "some-guid",
						Name:          "some-app",
						State:         constant.ApplicationStarted,
						LifecycleType: constant.AppLifecycleTypeDocker,
					},
					CurrentDroplet: v7action.Droplet{
						Image: "docker/some-image",
					},
				}
			})

			It("displays the app information", func() {
				Expect(testUI.Out).To(Say(`name:\s+some-app`))
				Expect(testUI.Out).To(Say(`requested state:\s+started`))
				Expect(testUI.Out).To(Say(`routes:\s+\n`))
				Expect(testUI.Out).To(Say(`stack:\s+\n`))
				Expect(testUI.Out).To(Say(`(?m)docker image:\s+docker/some-image$\n`))
			})
		})

		When("the application is a buildpack app", func() {
			BeforeEach(func() {
				summary = v7action.ApplicationSummary{
					CurrentDroplet: v7action.Droplet{
						Stack: "cflinuxfs2",
						Buildpacks: []v7action.DropletBuildpack{
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
				}
			})

			It("displays stack and buildpacks", func() {
				Expect(testUI.Out).To(Say(`stack:\s+cflinuxfs2`))
				Expect(testUI.Out).To(Say(`buildpacks:\s+some-detect-output, some-buildpack`))
			})
		})
	})
})
