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
	. "code.cloudfoundry.org/cli/plugin/v7/rpc"
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
		rpcService *CliRpcService
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
			rpcService, err = NewRpcService(nil, nil, nil, rpc.DefaultServer, nil, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an err of another Rpc process is already registered", func() {
			_, err := NewRpcService(nil, nil, nil, rpc.DefaultServer, nil, nil)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe(".Stop", func() {
		BeforeEach(func() {
			rpcService, err = NewRpcService(nil, nil, nil, rpc.DefaultServer, nil, nil)
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
			rpcService, err = NewRpcService(nil, nil, nil, rpc.DefaultServer, nil, nil)
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
			rpcService, err = NewRpcService(nil, nil, nil, rpc.DefaultServer, nil, nil)
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
			rpcService, err = NewRpcService(nil, terminalOutputSwitch, nil, rpc.DefaultServer, nil, nil)
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
			summary         v7action.DetailedApplicationSummary
		)

		BeforeEach(func() {
			fakePluginActor = new(rpcfakes.FakePluginActor)
			fakeConfig = new(commandfakes.FakeConfig)
			outputCapture := terminal.NewTeePrinter(os.Stdout)
			terminalOutputSwitch := terminal.NewTeePrinter(os.Stdout)

			rpcService, err = NewRpcService(outputCapture, terminalOutputSwitch, nil, rpc.DefaultServer, fakeConfig, fakePluginActor)
			Expect(err).ToNot(HaveOccurred())

			err := rpcService.Start()
			Expect(err).ToNot(HaveOccurred())

			pingCli(rpcService.Port())

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

			client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			rpcService.Stop()

			//give time for server to stop
			time.Sleep(50 * time.Millisecond)
		})

		FDescribe("GetApp", func() {

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
			PContext("when there are warnings returned", func() {

			})
		})
		FDescribe("AccessToken", func() {

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
