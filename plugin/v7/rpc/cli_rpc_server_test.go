// +build V7

package rpc_test

import (
	"errors"
	"net"
	"net/rpc"
	"os"
	"time"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/command/commandfakes"
	plugin "code.cloudfoundry.org/cli/plugin/v7"
	plugin_models "code.cloudfoundry.org/cli/plugin/v7/models"
	cmdRunner "code.cloudfoundry.org/cli/plugin/v7/rpc"
	"code.cloudfoundry.org/cli/plugin/v7/rpc/rpcfakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {

	var (
		err        error
		client     *rpc.Client
		rpcService *cmdRunner.CliRpcService
	)

	AfterEach(func() {
		if client != nil {
			client.Close()
		}
	})

	BeforeEach(func() {
		rpc.DefaultServer = rpc.NewServer()
	})

	Describe(".NewRpcService", func() {
		BeforeEach(func() {
			rpcService, err = cmdRunner.NewRpcService(nil, nil, nil, rpc.DefaultServer, nil, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an err of another Rpc process is already registered", func() {
			_, err := cmdRunner.NewRpcService(nil, nil, nil, rpc.DefaultServer, nil, nil)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe(".Stop", func() {
		BeforeEach(func() {
			rpcService, err = cmdRunner.NewRpcService(nil, nil, nil, rpc.DefaultServer, nil, nil)
			Expect(err).ToNot(HaveOccurred())

			err := rpcService.Start()
			Expect(err).ToNot(HaveOccurred())

			pingCli(rpcService.Port())
		})

		It("shuts down the rpc server", func() {
			rpcService.Stop()

			//give time for server to stop
			time.Sleep(50 * time.Millisecond)

			client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
			Expect(err).To(HaveOccurred())
		})
	})

	Describe(".Start", func() {
		BeforeEach(func() {
			rpcService, err = cmdRunner.NewRpcService(nil, nil, nil, rpc.DefaultServer, nil, nil)
			Expect(err).ToNot(HaveOccurred())

			err := rpcService.Start()
			Expect(err).ToNot(HaveOccurred())

			pingCli(rpcService.Port())
		})

		AfterEach(func() {
			rpcService.Stop()

			//give time for server to stop
			time.Sleep(50 * time.Millisecond)
		})

		It("Start an Rpc server for communication", func() {
			client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe(".SetPluginMetadata", func() {
		var (
			metadata *plugin.PluginMetadata
		)

		BeforeEach(func() {
			rpcService, err = cmdRunner.NewRpcService(nil, nil, nil, rpc.DefaultServer, nil, nil)
			Expect(err).ToNot(HaveOccurred())

			err := rpcService.Start()
			Expect(err).ToNot(HaveOccurred())

			pingCli(rpcService.Port())

			client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
			Expect(err).ToNot(HaveOccurred())

			metadata = &plugin.PluginMetadata{
				Name: "foo",
				Commands: []plugin.Command{
					{Name: "cmd_1", HelpText: "cm 1 help text"},
					{Name: "cmd_2", HelpText: "cmd 2 help text"},
				},
			}
		})

		AfterEach(func() {
			rpcService.Stop()

			//give time for server to stop
			time.Sleep(50 * time.Millisecond)
		})

		It("set the rpc command's Return Data", func() {
			var success bool
			err = client.Call("CliRpcCmd.SetPluginMetadata", metadata, &success)

			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
			Expect(rpcService.RpcCmd.PluginMetadata).To(Equal(metadata))
		})
	})

	Describe("disabling terminal output", func() {
		var terminalOutputSwitch *rpcfakes.FakeTerminalOutputSwitch

		BeforeEach(func() {
			terminalOutputSwitch = new(rpcfakes.FakeTerminalOutputSwitch)
			rpcService, err = cmdRunner.NewRpcService(nil, terminalOutputSwitch, nil, rpc.DefaultServer, nil, nil)
			Expect(err).ToNot(HaveOccurred())

			err := rpcService.Start()
			Expect(err).ToNot(HaveOccurred())

			pingCli(rpcService.Port())
		})

		It("should disable the terminal output switch", func() {
			client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
			Expect(err).ToNot(HaveOccurred())

			var success bool
			err = client.Call("CliRpcCmd.DisableTerminalOutput", true, &success)

			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
			Expect(terminalOutputSwitch.DisableTerminalOutputCallCount()).To(Equal(1))
			Expect(terminalOutputSwitch.DisableTerminalOutputArgsForCall(0)).To(Equal(true))
		})
	})

	Describe("Plugin API", func() {
		var (
			fakePluginActor *rpcfakes.FakePluginActor
			fakeConfig      *commandfakes.FakeConfig
		)

		BeforeEach(func() {
			fakePluginActor = new(rpcfakes.FakePluginActor)
			fakeConfig = new(commandfakes.FakeConfig)
			outputCapture := terminal.NewTeePrinter(os.Stdout)
			terminalOutputSwitch := terminal.NewTeePrinter(os.Stdout)

			rpcService, err = cmdRunner.NewRpcService(outputCapture, terminalOutputSwitch, nil, rpc.DefaultServer, fakeConfig, fakePluginActor)
			Expect(err).ToNot(HaveOccurred())

			err := rpcService.Start()
			Expect(err).ToNot(HaveOccurred())

			pingCli(rpcService.Port())
			client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			rpcService.Stop()

			//give time for server to stop
			time.Sleep(50 * time.Millisecond)
		})

		Describe("GetApp", func() {
			var (
				summary v7action.DetailedApplicationSummary
			)
			BeforeEach(func() {
				summary = v7action.DetailedApplicationSummary{
					ApplicationSummary: v7action.ApplicationSummary{
						Application: v7action.Application{
							GUID:      "some-app-guid",
							Name:      "some-app",
							StackName: "some-stack",
							State:     constant.ApplicationStarted,
						},
						ProcessSummaries: v7action.ProcessSummaries{
							{
								Process: v7action.Process{
									Type:               constant.ProcessTypeWeb,
									Command:            *types.NewFilteredString("some-command-1"),
									MemoryInMB:         types.NullUint64{IsSet: true, Value: 512},
									DiskInMB:           types.NullUint64{IsSet: true, Value: 64},
									HealthCheckTimeout: 60,
									Instances:          types.NullInt{IsSet: true, Value: 5},
								},
								InstanceDetails: []v7action.ProcessInstance{
									{State: constant.ProcessInstanceRunning},
									{State: constant.ProcessInstanceRunning},
									{State: constant.ProcessInstanceCrashed},
									{State: constant.ProcessInstanceRunning},
									{State: constant.ProcessInstanceRunning},
								},
							},
							{
								Process: v7action.Process{
									Type:               "console",
									Command:            *types.NewFilteredString("some-command-2"),
									MemoryInMB:         types.NullUint64{IsSet: true, Value: 256},
									DiskInMB:           types.NullUint64{IsSet: true, Value: 16},
									HealthCheckTimeout: 120,
									Instances:          types.NullInt{IsSet: true, Value: 1},
								},
								InstanceDetails: []v7action.ProcessInstance{
									{State: constant.ProcessInstanceRunning},
								},
							},
						},
					},
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
				fakePluginActor.GetDetailedAppSummaryReturns(summary, v7action.Warnings{"warning-1", "warning-2"}, nil)

				fakeConfig.TargetedSpaceReturns(configv3.Space{
					Name: "some-space",
					GUID: "some-space-guid",
				})

			})

			It("retrieves the app summary", func() {
				result := plugin_models.DetailedApplicationSummary{}
				err := client.Call("CliRpcCmd.GetApp", "some-app", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePluginActor.GetDetailedAppSummaryCallCount()).To(Equal(1))
				appName, spaceGUID, withObfuscatedValues := fakePluginActor.GetDetailedAppSummaryArgsForCall(0)
				Expect(appName).To(Equal("some-app"))
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(withObfuscatedValues).To(BeTrue())
			})

			It("populates the plugin model with the retrieved app", func() {
				result := plugin_models.DetailedApplicationSummary{}
				err := client.Call("CliRpcCmd.GetApp", "some-app", &result)
				Expect(err).ToNot(HaveOccurred())

				//fmt.Fprintf(os.Stdout, "%+v", result)
				Expect(result).To(BeEquivalentTo(summary))
			})

			Context("when retrieving the app fails", func() {
				BeforeEach(func() {
					fakePluginActor.GetDetailedAppSummaryReturns(v7action.DetailedApplicationSummary{}, nil, errors.New("some-error"))
				})
				It("returns an error", func() {
					result := plugin_models.DetailedApplicationSummary{}
					err := client.Call("CliRpcCmd.GetApp", "some-app", &result)
					Expect(err).To(MatchError("some-error"))
				})
			})
		})

		Describe("GetOrg", func() {
			var (
				org    v7action.Organization
				spaces []v7action.Space
			)

			BeforeEach(func() {
				org = v7action.Organization{
					Name: "org-name",
					GUID: "org-guid",
				}
				spaces = []v7action.Space{
					v7action.Space{
						Name: "space-name-1",
						GUID: "space-guid-1",
					},
					v7action.Space{
						Name: "space-name-2",
						GUID: "space-guid-2",
					},
				}
				fakePluginActor.GetOrganizationByNameReturns(org, nil, nil)
				fakePluginActor.GetOrganizationSpacesReturns(spaces, nil, nil)
			})

			It("retrives the organization", func() {
				result := plugin_models.Organization{}
				err := client.Call("CliRpcCmd.GetOrg", "org-name", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePluginActor.GetOrganizationByNameCallCount()).To(Equal(1))
				orgName := fakePluginActor.GetOrganizationByNameArgsForCall(0)
				Expect(orgName).To(Equal(org.Name))
			})

			It("retrives the spaces for the organization", func() {
				result := plugin_models.Organization{}
				err := client.Call("CliRpcCmd.GetOrg", "org-name", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePluginActor.GetOrganizationSpacesCallCount()).To(Equal(1))
				orgGUID := fakePluginActor.GetOrganizationSpacesArgsForCall(0)
				Expect(orgGUID).To(Equal(org.GUID))
			})

			It("populates the plugin model with the retrieved org and space information", func() {
				result := plugin_models.Organization{}
				err := client.Call("CliRpcCmd.GetOrg", "org-name", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result.Name).To(Equal(org.Name))
				Expect(result.GUID).To(Equal(org.GUID))
				Expect(len(result.Spaces)).To(Equal(2))
				Expect(result.Spaces[1].Name).To(Equal(spaces[1].Name))
			})

			Context("when retrieving the org fails", func() {
				BeforeEach(func() {
					fakePluginActor.GetOrganizationByNameReturns(v7action.Organization{}, nil, errors.New("org-error"))
				})

				It("returns an error", func() {
					result := plugin_models.Organization{}
					err := client.Call("CliRpcCmd.GetOrg", "some-org", &result)
					Expect(err).To(MatchError("org-error"))
				})
			})

			Context("when retrieving the space fails", func() {
				BeforeEach(func() {
					fakePluginActor.GetOrganizationSpacesReturns([]v7action.Space{}, nil, errors.New("space-error"))
				})

				It("returns an error", func() {
					result := plugin_models.Organization{}
					err := client.Call("CliRpcCmd.GetOrg", "some-org", &result)
					Expect(err).To(MatchError("space-error"))
				})
			})
		})

		Describe("GetCurrentSpace", func() {
			BeforeEach(func() {
				fakeConfig.TargetedSpaceReturns(configv3.Space{
					Name: "the-charlatans",
					GUID: "united-travel-service",
				})
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{
					Name: "the-actress",
					GUID: "family",
				})
				expectedSpace := v7action.Space{
					GUID: "united-travel-service",
					Name: "the-charlatans",
				}
				fakePluginActor.GetSpaceByNameAndOrganizationReturns(expectedSpace, v7action.Warnings{}, nil)
			})

			It("populates the plugin Space object with the current space settings in config", func() {
				var space plugin_models.Space
				err = client.Call("CliRpcCmd.GetCurrentSpace", "", &space)

				Expect(err).ToNot(HaveOccurred())

				result := plugin_models.DetailedApplicationSummary{}
				err := client.Call("CliRpcCmd.GetApp", "some-app", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePluginActor.GetSpaceByNameAndOrganizationCallCount()).To(Equal(1))
				spaceName, orgGUID := fakePluginActor.GetSpaceByNameAndOrganizationArgsForCall(0)
				Expect(spaceName).To(Equal("the-charlatans"))
				Expect(orgGUID).To(Equal("family"))

				Expect(space.Name).To(Equal("the-charlatans"))
				Expect(space.GUID).To(Equal("united-travel-service"))
			})

			Context("when retrieving the current space fails", func() {
				BeforeEach(func() {
					fakePluginActor.GetSpaceByNameAndOrganizationReturns(v7action.Space{}, nil, errors.New("some-error"))
				})
				It("returns an error", func() {
					result := plugin_models.DetailedApplicationSummary{}
					err := client.Call("CliRpcCmd.GetCurrentSpace", "", &result)
					Expect(err).To(MatchError("some-error"))
				})
			})
		})

		Describe("AccessToken", func() {

			BeforeEach(func() {
				fakePluginActor.RefreshAccessTokenReturns("token example", nil)
			})
			It("retrieves the access token", func() {

				result := ""
				err := client.Call("CliRpcCmd.AccessToken", "", &result)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakePluginActor.RefreshAccessTokenCallCount()).To(Equal(1))
				Expect(result).To(Equal("token example"))

			})

		})
	})

	Describe(".CallCoreCommand", func() {

		Describe("CLI Config object methods", func() {
			var (
				fakeConfig *commandfakes.FakeConfig
			)

			BeforeEach(func() {
				fakeConfig = new(commandfakes.FakeConfig)
			})

			AfterEach(func() {
				//give time for server to stop
				time.Sleep(50 * time.Millisecond)
			})

			Context(".ApiEndpoint", func() {
				BeforeEach(func() {
					rpcService, err = cmdRunner.NewRpcService(nil, nil, nil, rpc.DefaultServer, fakeConfig, nil)
					err := rpcService.Start()
					Expect(err).ToNot(HaveOccurred())

					pingCli(rpcService.Port())
					fakeConfig.TargetReturns("www.fake-domain.com")
				})
				AfterEach(func() {
					rpcService.Stop()

					//give time for server to stop
					time.Sleep(50 * time.Millisecond)
				})

				It("returns the ApiEndpoint() setting in config", func() {
					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var result string
					err = client.Call("CliRpcCmd.ApiEndpoint", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal("www.fake-domain.com"))
					Expect(fakeConfig.TargetCallCount()).To(Equal(1))
				})
			})
		})
	})

})

func pingCli(port string) {
	var connErr error
	var conn net.Conn
	for i := 0; i < 5; i++ {
		conn, connErr = net.Dial("tcp", "127.0.0.1:"+port)
		if connErr != nil {
			time.Sleep(200 * time.Millisecond)
		} else {
			conn.Close()
			break
		}
	}
	Expect(connErr).ToNot(HaveOccurred())
}
