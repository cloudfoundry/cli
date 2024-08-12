package shared_test

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var _ = Describe("app summary displayer", func() {

	const instanceStatsTitles = `state\s+since\s+cpu\s+memory\s+disk\s+logging\s+cpu entitlement\s+details`

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
			summary             v7action.DetailedApplicationSummary
			displayStartCommand bool
		)

		JustBeforeEach(func() {
			appSummaryDisplayer.AppDisplay(summary, displayStartCommand)
		})

		When("the app has instances", func() {
			When("the process instances are running", func() {
				var uptime time.Duration

				BeforeEach(func() {
					uptime = time.Since(time.Unix(267321600, 0))
					summary = v7action.DetailedApplicationSummary{
						ApplicationSummary: v7action.ApplicationSummary{
							Application: resources.Application{
								GUID:  "some-app-guid",
								State: constant.ApplicationStarted,
							},
							ProcessSummaries: v7action.ProcessSummaries{
								{
									Process: resources.Process{
										Type:              constant.ProcessTypeWeb,
										MemoryInMB:        types.NullUint64{Value: 32, IsSet: true},
										DiskInMB:          types.NullUint64{Value: 1024, IsSet: true},
										LogRateLimitInBPS: types.NullInt{Value: 1024 * 5, IsSet: true},
									},
									Sidecars: []resources.Sidecar{},
									InstanceDetails: []v7action.ProcessInstance{
										v7action.ProcessInstance{
											Index:          0,
											State:          constant.ProcessInstanceRunning,
											CPUEntitlement: types.NullFloat64{Value: 0, IsSet: true},
											MemoryUsage:    1000000,
											DiskUsage:      1000000,
											LogRate:        1024,
											MemoryQuota:    33554432,
											DiskQuota:      2000000,
											LogRateLimit:   1024 * 5,
											Uptime:         uptime,
											Details:        "Some Details 1",
										},
										v7action.ProcessInstance{
											Index:          1,
											State:          constant.ProcessInstanceRunning,
											CPUEntitlement: types.NullFloat64{Value: 0, IsSet: false},
											MemoryUsage:    2000000,
											DiskUsage:      2000000,
											LogRate:        1024 * 2,
											MemoryQuota:    33554432,
											DiskQuota:      4000000,
											LogRateLimit:   1024 * 5,
											Uptime:         time.Since(time.Unix(330480000, 0)),
											Details:        "Some Details 2",
										},
										v7action.ProcessInstance{
											Index:          2,
											State:          constant.ProcessInstanceRunning,
											CPUEntitlement: types.NullFloat64{Value: 0.03, IsSet: true},
											MemoryUsage:    3000000,
											DiskUsage:      3000000,
											LogRate:        1024 * 3,
											MemoryQuota:    33554432,
											DiskQuota:      6000000,
											LogRateLimit:   1024 * 5,
											Uptime:         time.Since(time.Unix(1277164800, 0)),
										},
									},
								},
								{
									Process: resources.Process{
										Type:              "console",
										MemoryInMB:        types.NullUint64{Value: 16, IsSet: true},
										DiskInMB:          types.NullUint64{Value: 512, IsSet: true},
										LogRateLimitInBPS: types.NullInt{Value: 256, IsSet: true},
									},
									Sidecars: []resources.Sidecar{},
									InstanceDetails: []v7action.ProcessInstance{
										v7action.ProcessInstance{
											Index:        0,
											State:        constant.ProcessInstanceRunning,
											MemoryUsage:  1000000,
											DiskUsage:    1000000,
											LogRate:      128,
											MemoryQuota:  33554432,
											DiskQuota:    8000000,
											LogRateLimit: 256,
											Uptime:       time.Since(time.Unix(167572800, 0)),
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
					Expect(webProcessSummary.Sidecars).To(Equal(""))
					Expect(webProcessSummary.InstanceCount).To(Equal("3/3"))
					Expect(webProcessSummary.MemUsage).To(Equal("32M"))

					Expect(webProcessSummary.Instances[0].Memory).To(Equal("976.6K of 32M"))
					Expect(webProcessSummary.Instances[0].Since).To(MatchRegexp(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`))
					Expect(time.Parse(time.RFC3339, webProcessSummary.Instances[0].Since)).To(BeTemporally("~", time.Now().Add(-uptime), 2*time.Second))
					Expect(webProcessSummary.Instances[0].Disk).To(Equal("976.6K of 1.9M"))
					Expect(webProcessSummary.Instances[0].CPU).To(Equal("0.0%"))
					Expect(webProcessSummary.Instances[0].CPUEntitlement).To(Equal("0.0%"))
					Expect(webProcessSummary.Instances[0].LogRate).To(Equal("1K/s of 5K/s"))
					Expect(webProcessSummary.Instances[0].Details).To(Equal("Some Details 1"))

					Expect(webProcessSummary.Instances[1].Memory).To(Equal("1.9M of 32M"))
					Expect(webProcessSummary.Instances[1].Disk).To(Equal("1.9M of 3.8M"))
					Expect(webProcessSummary.Instances[1].CPU).To(Equal("0.0%"))
					Expect(webProcessSummary.Instances[1].CPUEntitlement).To(Equal(""))
					Expect(webProcessSummary.Instances[1].LogRate).To(Equal("2K/s of 5K/s"))
					Expect(webProcessSummary.Instances[1].Details).To(Equal("Some Details 2"))

					Expect(webProcessSummary.Instances[2].Memory).To(Equal("2.9M of 32M"))
					Expect(webProcessSummary.Instances[2].Disk).To(Equal("2.9M of 5.7M"))
					Expect(webProcessSummary.Instances[2].CPU).To(Equal("0.0%"))
					Expect(webProcessSummary.Instances[2].CPUEntitlement).To(Equal("3.0%"))
					Expect(webProcessSummary.Instances[2].LogRate).To(Equal("3K/s of 5K/s"))

					consoleProcessSummary := processTable.Processes[1]
					Expect(consoleProcessSummary.Type).To(Equal("console"))
					Expect(consoleProcessSummary.Sidecars).To(Equal(""))
					Expect(consoleProcessSummary.InstanceCount).To(Equal("1/1"))
					Expect(consoleProcessSummary.MemUsage).To(Equal("16M"))

					Expect(consoleProcessSummary.Instances[0].Memory).To(Equal("976.6K of 32M"))
					Expect(consoleProcessSummary.Instances[0].Disk).To(Equal("976.6K of 7.6M"))
					Expect(consoleProcessSummary.Instances[0].CPU).To(Equal("0.0%"))
					Expect(consoleProcessSummary.Instances[0].CPUEntitlement).To(Equal(""))
					Expect(consoleProcessSummary.Instances[0].LogRate).To(Equal("128B/s of 256B/s"))
				})
			})

			When("the log rate is unlimited", func() {
				BeforeEach(func() {
					summary = v7action.DetailedApplicationSummary{
						ApplicationSummary: v7action.ApplicationSummary{
							Application: resources.Application{
								GUID:  "some-app-guid",
								State: constant.ApplicationStarted,
							},
							ProcessSummaries: v7action.ProcessSummaries{
								{
									Process: resources.Process{
										Type:       constant.ProcessTypeWeb,
										MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
										DiskInMB:   types.NullUint64{Value: 1024, IsSet: true},
									},
									Sidecars: []resources.Sidecar{},
									InstanceDetails: []v7action.ProcessInstance{
										v7action.ProcessInstance{
											Index:        0,
											State:        constant.ProcessInstanceRunning,
											LogRate:      1024,
											LogRateLimit: -1,
										},
									},
								},
							},
						},
					}
				})

				It("renders unlimited log rate limits correctly", func() {
					processTable := helpers.ParseV3AppProcessTable(output.Contents())
					webProcessSummary := processTable.Processes[0]

					Expect(webProcessSummary.Instances[0].LogRate).To(Equal("1K/s of unlimited"))
				})
			})

			When("some processes have > 0 instances and others have 0 instances", func() {
				BeforeEach(func() {
					summary = v7action.DetailedApplicationSummary{
						ApplicationSummary: v7action.ApplicationSummary{
							Application: resources.Application{
								GUID:  "some-app-guid",
								State: constant.ApplicationStarted,
							},
							ProcessSummaries: v7action.ProcessSummaries{
								{
									Process: resources.Process{
										Type:       constant.ProcessTypeWeb,
										MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
										DiskInMB:   types.NullUint64{Value: 1024, IsSet: true},
									},
									Sidecars: []resources.Sidecar{},
									InstanceDetails: []v7action.ProcessInstance{
										v7action.ProcessInstance{
											Index:       0,
											State:       constant.ProcessInstanceRunning,
											MemoryUsage: 1000000,
											DiskUsage:   1000000,
											MemoryQuota: 33554432,
											DiskQuota:   2000000,
											Uptime:      time.Since(time.Unix(267321600, 0)),
										},
									},
								},
								{
									Process: resources.Process{
										Type:       "console",
										MemoryInMB: types.NullUint64{Value: 16, IsSet: true},
										DiskInMB:   types.NullUint64{Value: 512, IsSet: true},
									},
									Sidecars: []resources.Sidecar{},
								},
							},
						},
					}
				})

				It("lists instance stats for process types that have > 0 instances", func() {
					Expect(testUI.Out).To(Say(`type:\s+web`))
					Expect(testUI.Out).To(Say(`sidecars: `))
					Expect(testUI.Out).To(Say(instanceStatsTitles))
				})

				It("does not show the instance stats table for process types with 0 instances", func() {
					Expect(testUI.Out).To(Say(`type:\s+console`))
					Expect(testUI.Out).To(Say(`sidecars: `))
					Expect(testUI.Out).To(Say("There are no running instances of this process."))
				})
			})

			When("all the instances for a process are down (but scaled to > 0 instances)", func() {
				BeforeEach(func() {
					summary = v7action.DetailedApplicationSummary{
						ApplicationSummary: v7action.ApplicationSummary{
							ProcessSummaries: []v7action.ProcessSummary{
								{
									Process: resources.Process{
										Type:       constant.ProcessTypeWeb,
										MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
									},
									Sidecars:        []resources.Sidecar{},
									InstanceDetails: []v7action.ProcessInstance{{State: constant.ProcessInstanceDown}},
								}},
						},
					}
				})

				It("displays the instances table", func() {
					Expect(testUI.Out).To(Say(`type:\s+web`))
					Expect(testUI.Out).To(Say(`sidecars: `))
					Expect(testUI.Out).To(Say(instanceStatsTitles))
				})
			})

			Describe("start command", func() {
				BeforeEach(func() {
					summary = v7action.DetailedApplicationSummary{
						ApplicationSummary: v7action.ApplicationSummary{
							Application: resources.Application{
								GUID:  "some-app-guid",
								State: constant.ApplicationStarted,
							},
							ProcessSummaries: v7action.ProcessSummaries{
								{
									Process: resources.Process{
										Type:    constant.ProcessTypeWeb,
										Command: *types.NewFilteredString("some-command-1"),
									},
									Sidecars: []resources.Sidecar{},
								},
								{
									Process: resources.Process{
										Type:    "console",
										Command: *types.NewFilteredString("some-command-2"),
									},
									Sidecars: []resources.Sidecar{},
								},
								{
									Process: resources.Process{
										Type: "random",
									},
									Sidecars: []resources.Sidecar{},
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
						Expect(testUI.Out).To(Say(`sidecars: `))
						Expect(testUI.Out).To(Say(`start command:\s+some-command-1`))

						Expect(testUI.Out).To(Say(`type:\s+console`))
						Expect(testUI.Out).To(Say(`sidecars: `))
						Expect(testUI.Out).To(Say(`start command:\s+some-command-2`))

						Expect(testUI.Out).To(Say(`type:\s+random`))
						Expect(testUI.Out).To(Say(`sidecars: `))
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
				summary = v7action.DetailedApplicationSummary{
					ApplicationSummary: v7action.ApplicationSummary{
						Application: resources.Application{
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
						ProcessSummaries: v7action.ProcessSummaries{
							{
								Process: resources.Process{
									Type:       constant.ProcessTypeWeb,
									MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
									DiskInMB:   types.NullUint64{Value: 1024, IsSet: true},
								},
								Sidecars: []resources.Sidecar{},
							},
							{
								Process: resources.Process{
									Type:       "console",
									MemoryInMB: types.NullUint64{Value: 16, IsSet: true},
									DiskInMB:   types.NullUint64{Value: 512, IsSet: true},
								},
								Sidecars: []resources.Sidecar{},
							},
						},
					},
				}
			})

			It("lists information for each of the processes", func() {
				Expect(testUI.Out).To(Say(`type:\s+web`))
				Expect(testUI.Out).To(Say(`sidecars: `))
				Expect(testUI.Out).To(Say(`instances:\s+0/0`))
				Expect(testUI.Out).To(Say(`memory usage:\s+32M`))
				Expect(testUI.Out).To(Say("There are no running instances of this process."))

				Expect(testUI.Out).To(Say(`type:\s+console`))
				Expect(testUI.Out).To(Say(`sidecars: `))
				Expect(testUI.Out).To(Say(`instances:\s+0/0`))
				Expect(testUI.Out).To(Say(`memory usage:\s+16M`))
				Expect(testUI.Out).To(Say("There are no running instances of this process."))
			})

			It("does not display the instance table", func() {
				Expect(testUI.Out).NotTo(Say(instanceStatsTitles))
			})
		})

		When("the app has sidecars", func() {
			BeforeEach(func() {
				summary = v7action.DetailedApplicationSummary{
					ApplicationSummary: v7action.ApplicationSummary{
						Application: resources.Application{
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
						ProcessSummaries: v7action.ProcessSummaries{
							{
								Process: resources.Process{
									Type:       constant.ProcessTypeWeb,
									MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
									DiskInMB:   types.NullUint64{Value: 1024, IsSet: true},
								},
								Sidecars: []resources.Sidecar{
									{Name: "authenticator"},
									{Name: "clock"},
								},
							},
						},
					},
				}
			})

			It("lists information for each of the processes", func() {
				Expect(testUI.Out).To(Say(`type:\s+web`))
				Expect(testUI.Out).To(Say(`sidecars:\s+authenticator, clock`))
				Expect(testUI.Out).To(Say(`instances:\s+0/0`))
				Expect(testUI.Out).To(Say(`memory usage:\s+32M`))
				Expect(testUI.Out).To(Say("There are no running instances of this process."))
			})

			It("does not display the instance table", func() {
				Expect(testUI.Out).NotTo(Say(instanceStatsTitles))
			})
		})

		When("the app is stopped", func() {
			BeforeEach(func() {
				summary = v7action.DetailedApplicationSummary{
					ApplicationSummary: v7action.ApplicationSummary{
						Application: resources.Application{
							GUID:  "some-app-guid",
							State: constant.ApplicationStopped,
						},
						ProcessSummaries: v7action.ProcessSummaries{
							{
								Process: resources.Process{
									Type: constant.ProcessTypeWeb,
								},
								Sidecars: []resources.Sidecar{},
							},
							{
								Process: resources.Process{
									Type: "console",
								},
								Sidecars: []resources.Sidecar{},
							},
						},
					},
				}
			})

			It("lists information for each of the processes", func() {
				Expect(testUI.Out).To(Say(`type:\s+web`))
				Expect(testUI.Out).To(Say(`sidecars: `))
				Expect(testUI.Out).To(Say("There are no running instances of this process."))

				Expect(testUI.Out).To(Say(`type:\s+console`))
				Expect(testUI.Out).To(Say(`sidecars: `))
				Expect(testUI.Out).To(Say("There are no running instances of this process."))
			})

			It("does not display the instance table", func() {
				Expect(testUI.Out).NotTo(Say(instanceStatsTitles))
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
					summary = v7action.DetailedApplicationSummary{
						ApplicationSummary: v7action.ApplicationSummary{
							Application: resources.Application{
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
				summary.Routes = []resources.Route{
					{Host: "route1", URL: "route1.example.com"},
					{Host: "route2", URL: "route2.example.com"},
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
				summary = v7action.DetailedApplicationSummary{
					ApplicationSummary: v7action.ApplicationSummary{
						Application: resources.Application{
							GUID:          "some-guid",
							Name:          "some-app",
							State:         constant.ApplicationStarted,
							LifecycleType: constant.AppLifecycleTypeDocker,
						},
					},
					CurrentDroplet: resources.Droplet{
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

			It("does not display the buildpack info for docker apps", func() {
				Expect(testUI.Out).ToNot(Say("buildpacks:"))
			})
		})

		When("the application is a buildpack app", func() {
			BeforeEach(func() {
				summary = v7action.DetailedApplicationSummary{
					ApplicationSummary: v7action.ApplicationSummary{
						Application: resources.Application{
							LifecycleType: constant.AppLifecycleTypeBuildpack,
						},
					},
					CurrentDroplet: resources.Droplet{
						Stack: "cflinuxfs2",
						Buildpacks: []resources.DropletBuildpack{
							{
								Name:          "ruby_buildpack",
								BuildpackName: "ruby_buildpack_name",
								DetectOutput:  "some-detect-output",
								Version:       "0.0.1",
							},
							{
								Name:          "go_buildpack_without_detect_output",
								BuildpackName: "go_buildpack_name",
								DetectOutput:  "",
								Version:       "0.0.2",
							},
							{
								Name:          "go_buildpack_without_version",
								BuildpackName: "go_buildpack_name",
								DetectOutput:  "",
								Version:       "",
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
				Expect(testUI.Out).To(Say(`stack:\s+cflinuxfs2\n`))
				Expect(testUI.Out).To(Say(`buildpacks:\s+\n`))
				Expect(testUI.Out).To(Say(`name\s+version\s+detect output\s+buildpack name\n`))
				Expect(testUI.Out).To(Say(`ruby_buildpack\s+0.0.1\s+some-detect-output\s+ruby_buildpack_name\n`))
				Expect(testUI.Out).To(Say(`some-buildpack`))
			})
		})

		When("there is an active deployment", func() {
			var LastStatusChangeTimeString = "2024-07-29T17:32:29Z"
			var dateTimeRegexPattern = `[a-zA-Z]{3}\s\d{2}\s[a-zA-Z]{3}\s\d{2}\:\d{2}\:\d{2}\s[A-Z]{3}\s\d{4}`
			var maxInFlightDefaultValue = 1

			When("the deployment strategy is rolling", func() {
				When("the deployment is in progress", func() {
					When("last status change has a timestamp and max-in-flight is non-default", func() {
						BeforeEach(func() {
							summary = v7action.DetailedApplicationSummary{
								Deployment: resources.Deployment{
									Strategy:         constant.DeploymentStrategyRolling,
									StatusValue:      constant.DeploymentStatusValueActive,
									StatusReason:     constant.DeploymentStatusReasonDeploying,
									LastStatusChange: LastStatusChangeTimeString,
									Options: resources.DeploymentOpts{
										MaxInFlight: 2,
									},
								},
							}
						})

						It("displays the message", func() {
							var actualOut = fmt.Sprintf("%s", testUI.Out)
							Expect(actualOut).To(MatchRegexp(`Rolling deployment currently DEPLOYING \(since %s\)`, dateTimeRegexPattern))
						})
						It("displays max-in-flight value", func() {
							Expect(testUI.Out).To(Say(`max-in-flight: 2`))
						})
					})
					When("last status change has a timestamp and max-in-flight is default", func() {
						BeforeEach(func() {
							summary = v7action.DetailedApplicationSummary{
								Deployment: resources.Deployment{
									Strategy:         constant.DeploymentStrategyRolling,
									StatusValue:      constant.DeploymentStatusValueActive,
									StatusReason:     constant.DeploymentStatusReasonDeploying,
									LastStatusChange: LastStatusChangeTimeString,
									Options: resources.DeploymentOpts{
										MaxInFlight: maxInFlightDefaultValue,
									},
								},
							}
						})

						It("displays the message", func() {
							var actualOut = fmt.Sprintf("%s", testUI.Out)
							Expect(actualOut).To(MatchRegexp(`Rolling deployment currently DEPLOYING \(since %s\)`, dateTimeRegexPattern))
						})
						It("does not display max-in-flight", func() {
							Expect(testUI.Out).NotTo(Say(`max-in-flight`))
						})
					})
					// 'unset' is important for the newer-CLI-than-CAPI scenario
					When("last status change has a timestamp and max-in-flight is unset", func() {
						BeforeEach(func() {
							summary = v7action.DetailedApplicationSummary{
								Deployment: resources.Deployment{
									Strategy:         constant.DeploymentStrategyRolling,
									StatusValue:      constant.DeploymentStatusValueActive,
									StatusReason:     constant.DeploymentStatusReasonDeploying,
									LastStatusChange: LastStatusChangeTimeString,
								},
							}
						})

						It("displays the message", func() {
							var actualOut = fmt.Sprintf("%s", testUI.Out)
							Expect(actualOut).To(MatchRegexp(`Rolling deployment currently DEPLOYING \(since %s\)`, dateTimeRegexPattern))
						})
						It("does not display max-in-flight", func() {
							Expect(testUI.Out).NotTo(Say(`max-in-flight`))
						})
					})

					When("last status change is an empty string and max-in-flight is non-default", func() {
						BeforeEach(func() {
							summary = v7action.DetailedApplicationSummary{
								Deployment: resources.Deployment{
									Strategy:         constant.DeploymentStrategyRolling,
									StatusValue:      constant.DeploymentStatusValueActive,
									StatusReason:     constant.DeploymentStatusReasonDeploying,
									LastStatusChange: "",
									Options: resources.DeploymentOpts{
										MaxInFlight: 2,
									},
								},
							}
						})

						It("displays the message", func() {
							Expect(testUI.Out).To(Say(`Rolling deployment currently DEPLOYING`))
							Expect(testUI.Out).NotTo(Say(`\(since`))
						})
						It("displays max-in-flight value", func() {
							Expect(testUI.Out).To(Say(`max-in-flight: 2`))
						})
					})
					When("last status change is an empty string and max-in-flight is default", func() {
						BeforeEach(func() {
							summary = v7action.DetailedApplicationSummary{
								Deployment: resources.Deployment{
									Strategy:         constant.DeploymentStrategyRolling,
									StatusValue:      constant.DeploymentStatusValueActive,
									StatusReason:     constant.DeploymentStatusReasonDeploying,
									LastStatusChange: "",
									Options: resources.DeploymentOpts{
										MaxInFlight: maxInFlightDefaultValue,
									},
								},
							}
						})

						It("displays the message", func() {
							Expect(testUI.Out).To(Say(`Rolling deployment currently DEPLOYING`))
							Expect(testUI.Out).NotTo(Say(`\(since`))
						})
						It("does not display max-in-flight", func() {
							Expect(testUI.Out).NotTo(Say(`max-in-flight`))
						})
					})
				})

				When("the deployment is cancelled", func() {
					When("max-in-flight value is non-default", func() {
						BeforeEach(func() {
							summary = v7action.DetailedApplicationSummary{
								Deployment: resources.Deployment{
									Strategy:         constant.DeploymentStrategyRolling,
									StatusValue:      constant.DeploymentStatusValueActive,
									StatusReason:     constant.DeploymentStatusReasonCanceling,
									LastStatusChange: LastStatusChangeTimeString,
									Options: resources.DeploymentOpts{
										MaxInFlight: 2,
									},
								},
							}
						})

						It("displays the message", func() {
							var actualOut = fmt.Sprintf("%s", testUI.Out)
							Expect(actualOut).To(MatchRegexp(`Rolling deployment currently CANCELING \(since %s\)`, dateTimeRegexPattern))
						})
						It("displays max-in-flight value", func() {
							Expect(testUI.Out).To(Say(`max-in-flight: 2`))
						})
					})
					When("max-in-flight value is default", func() {
						BeforeEach(func() {
							summary = v7action.DetailedApplicationSummary{
								Deployment: resources.Deployment{
									Strategy:         constant.DeploymentStrategyRolling,
									StatusValue:      constant.DeploymentStatusValueActive,
									StatusReason:     constant.DeploymentStatusReasonCanceling,
									LastStatusChange: LastStatusChangeTimeString,
									Options: resources.DeploymentOpts{
										MaxInFlight: maxInFlightDefaultValue,
									},
								},
							}
						})

						It("displays the message", func() {
							var actualOut = fmt.Sprintf("%s", testUI.Out)
							Expect(actualOut).To(MatchRegexp(`Rolling deployment currently CANCELING \(since %s\)`, dateTimeRegexPattern))
						})
						It("does not display max-in-flight", func() {
							Expect(testUI.Out).NotTo(Say(`max-in-flight`))
						})
					})
				})
			})
			When("the deployment strategy is canary", func() {
				When("the deployment is in progress", func() {
					When("max-in-flight value is non-default", func() {
						BeforeEach(func() {
							summary = v7action.DetailedApplicationSummary{
								Deployment: resources.Deployment{
									Strategy:         constant.DeploymentStrategyCanary,
									StatusValue:      constant.DeploymentStatusValueActive,
									StatusReason:     constant.DeploymentStatusReasonDeploying,
									LastStatusChange: LastStatusChangeTimeString,
									Options: resources.DeploymentOpts{
										MaxInFlight: 2,
									},
								},
							}
						})

						It("displays the message", func() {
							var actualOut = fmt.Sprintf("%s", testUI.Out)
							Expect(actualOut).To(MatchRegexp(`Canary deployment currently DEPLOYING \(since %s\)`, dateTimeRegexPattern))
							Expect(testUI.Out).NotTo(Say(`promote the canary deployment`))
						})
						It("displays max-in-flight value", func() {
							Expect(testUI.Out).To(Say(`max-in-flight: 2`))
						})
					})
					When("max-in-flight value is default", func() {
						BeforeEach(func() {
							summary = v7action.DetailedApplicationSummary{
								Deployment: resources.Deployment{
									Strategy:         constant.DeploymentStrategyCanary,
									StatusValue:      constant.DeploymentStatusValueActive,
									StatusReason:     constant.DeploymentStatusReasonDeploying,
									LastStatusChange: LastStatusChangeTimeString,
									Options: resources.DeploymentOpts{
										MaxInFlight: maxInFlightDefaultValue,
									},
								},
							}
						})

						It("displays the message", func() {
							var actualOut = fmt.Sprintf("%s", testUI.Out)
							Expect(actualOut).To(MatchRegexp(`Canary deployment currently DEPLOYING \(since %s\)`, dateTimeRegexPattern))
							Expect(testUI.Out).NotTo(Say(`promote the canary deployment`))
						})
						It("does not display max-in-flight", func() {
							Expect(testUI.Out).NotTo(Say(`max-in-flight`))
						})
					})
				})

				When("the deployment is paused", func() {
					When("max-in-flight value is non-default", func() {
						BeforeEach(func() {
							summary = v7action.DetailedApplicationSummary{
								ApplicationSummary: v7action.ApplicationSummary{
									Application: resources.Application{
										Name: "foobar",
									},
								},
								Deployment: resources.Deployment{
									Strategy:         constant.DeploymentStrategyCanary,
									StatusValue:      constant.DeploymentStatusValueActive,
									StatusReason:     constant.DeploymentStatusReasonPaused,
									LastStatusChange: LastStatusChangeTimeString,
									Options: resources.DeploymentOpts{
										MaxInFlight: 2,
									},
								},
							}
						})

						It("displays the message", func() {
							var actualOut = fmt.Sprintf("%s", testUI.Out)
							Expect(actualOut).To(MatchRegexp(`Canary deployment currently PAUSED \(since %s\)`, dateTimeRegexPattern))
							Expect(testUI.Out).To(Say("Please run `cf continue-deployment foobar` to promote the canary deployment, or `cf cancel-deployment foobar` to rollback to the previous version."))
						})
						It("displays max-in-flight value", func() {
							Expect(testUI.Out).To(Say(`max-in-flight: 2`))
						})
					})
					When("max-in-flight value is default", func() {
						BeforeEach(func() {
							summary = v7action.DetailedApplicationSummary{
								ApplicationSummary: v7action.ApplicationSummary{
									Application: resources.Application{
										Name: "foobar",
									},
								},
								Deployment: resources.Deployment{
									Strategy:         constant.DeploymentStrategyCanary,
									StatusValue:      constant.DeploymentStatusValueActive,
									StatusReason:     constant.DeploymentStatusReasonPaused,
									LastStatusChange: LastStatusChangeTimeString,
									Options: resources.DeploymentOpts{
										MaxInFlight: maxInFlightDefaultValue,
									},
								},
							}
						})

						It("displays the message", func() {
							var actualOut = fmt.Sprintf("%s", testUI.Out)
							Expect(actualOut).To(MatchRegexp(`Canary deployment currently PAUSED \(since %s\)`, dateTimeRegexPattern))
							Expect(testUI.Out).To(Say("Please run `cf continue-deployment foobar` to promote the canary deployment, or `cf cancel-deployment foobar` to rollback to the previous version."))
						})
						It("does not display max-in-flight", func() {
							Expect(testUI.Out).NotTo(Say(`max-in-flight`))
						})
					})
				})

				When("the deployment is canceling", func() {
					When("max-in-flight value is non-default", func() {
						BeforeEach(func() {
							summary = v7action.DetailedApplicationSummary{
								Deployment: resources.Deployment{
									Strategy:         constant.DeploymentStrategyCanary,
									StatusValue:      constant.DeploymentStatusValueActive,
									StatusReason:     constant.DeploymentStatusReasonCanceling,
									LastStatusChange: LastStatusChangeTimeString,
									Options: resources.DeploymentOpts{
										MaxInFlight: 2,
									},
								},
							}
						})

						It("displays the message", func() {
							var actualOut = fmt.Sprintf("%s", testUI.Out)
							Expect(actualOut).To(MatchRegexp(`Canary deployment currently CANCELING \(since %s\)`, dateTimeRegexPattern))
							Expect(testUI.Out).NotTo(Say(`promote the canary deployment`))
						})
						It("displays max-in-flight value", func() {
							Expect(testUI.Out).To(Say(`max-in-flight: 2`))
						})
					})
					When("max-in-flight value is default", func() {
						BeforeEach(func() {
							summary = v7action.DetailedApplicationSummary{
								Deployment: resources.Deployment{
									Strategy:         constant.DeploymentStrategyCanary,
									StatusValue:      constant.DeploymentStatusValueActive,
									StatusReason:     constant.DeploymentStatusReasonCanceling,
									LastStatusChange: LastStatusChangeTimeString,
									Options: resources.DeploymentOpts{
										MaxInFlight: maxInFlightDefaultValue,
									},
								},
							}
						})

						It("displays the message", func() {
							var actualOut = fmt.Sprintf("%s", testUI.Out)
							Expect(actualOut).To(MatchRegexp(`Canary deployment currently CANCELING \(since %s\)`, dateTimeRegexPattern))
							Expect(testUI.Out).NotTo(Say(`promote the canary deployment`))
						})
						It("does not display max-in-flight", func() {
							Expect(testUI.Out).NotTo(Say(`max-in-flight`))
						})
					})
				})
			})
		})

		When("there is no active deployment", func() {
			BeforeEach(func() {
				summary = v7action.DetailedApplicationSummary{
					Deployment: resources.Deployment{
						Strategy:     "",
						StatusValue:  "",
						StatusReason: "",
					},
				}
			})

			It("does not display deployment info", func() {
				Expect(testUI.Out).NotTo(Say(fmt.Sprintf("%s deployment currently %s",
					cases.Title(language.English, cases.NoLower).String(string(summary.Deployment.Strategy)),
					summary.Deployment.StatusReason)))
			})
			It("does not display max-in-flight", func() {
				Expect(testUI.Out).NotTo(Say(`max-in-flight`))
			})
		})
	})
})
