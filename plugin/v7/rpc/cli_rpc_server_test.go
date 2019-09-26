// +build V7

package rpc_test

import (
	"net"
	"net/rpc"
	"os"
	"time"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/terminal"
	plugin "code.cloudfoundry.org/cli/plugin/v7"
	plugin_models "code.cloudfoundry.org/cli/plugin/v7/models"
	. "code.cloudfoundry.org/cli/plugin/v7/rpc"
	cmdRunner "code.cloudfoundry.org/cli/plugin/v7/rpc"
	. "code.cloudfoundry.org/cli/plugin/v7/rpc/fakecommand"
	"code.cloudfoundry.org/cli/plugin/v7/rpc/rpcfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {

	_ = FakeCommand1{} //make sure fake_command is imported and self-registered with init()

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
			rpcService, err = NewRpcService(nil, nil, nil, api.RepositoryLocator{}, nil, nil, nil, rpc.DefaultServer)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an err of another Rpc process is already registered", func() {
			_, err := NewRpcService(nil, nil, nil, api.RepositoryLocator{}, nil, nil, nil, rpc.DefaultServer)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe(".Stop", func() {
		BeforeEach(func() {
			rpcService, err = NewRpcService(nil, nil, nil, api.RepositoryLocator{}, nil, nil, nil, rpc.DefaultServer)
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
			rpcService, err = NewRpcService(nil, nil, nil, api.RepositoryLocator{}, nil, nil, nil, rpc.DefaultServer)
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

	// Describe(".IsMinCliVersion()", func() {
	// 	BeforeEach(func() {
	// 		rpcService, err = NewRpcService(nil, nil, nil, api.RepositoryLocator{}, nil, nil, nil, rpc.DefaultServer)
	// 		Expect(err).ToNot(HaveOccurred())

	// 		err := rpcService.Start()
	// 		Expect(err).ToNot(HaveOccurred())

	// 		pingCli(rpcService.Port())

	// 		client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
	// 		Expect(err).ToNot(HaveOccurred())
	// 	})

	// 	AfterEach(func() {
	// 		rpcService.Stop()

	// 		//give time for server to stop
	// 		time.Sleep(50 * time.Millisecond)
	// 	})

	// 	It("returns true if cli version is greater than the required version", func() {
	// 		version.BinaryVersion = "1.2.3+abc123"

	// 		var result bool
	// 		err = client.Call("CliRpcCmd.IsMinCliVersion", "1.2.2", &result)
	// 		Expect(err).ToNot(HaveOccurred())

	// 		Expect(result).To(BeTrue())
	// 	})

	// 	It("returns true if cli version is equal to the required version", func() {
	// 		version.BinaryVersion = "1.2.3+abc123"

	// 		var result bool
	// 		err = client.Call("CliRpcCmd.IsMinCliVersion", "1.2.3", &result)
	// 		Expect(err).ToNot(HaveOccurred())

	// 		Expect(result).To(BeTrue())
	// 	})

	// 	It("returns false if cli version is less than the required version", func() {
	// 		version.BinaryVersion = "1.2.3+abc123"

	// 		var result bool
	// 		err = client.Call("CliRpcCmd.IsMinCliVersion", "1.2.4", &result)
	// 		Expect(err).ToNot(HaveOccurred())

	// 		Expect(result).To(BeFalse())
	// 	})

	// 	It("returns true if cli version is 'BUILT_FROM_SOURCE'", func() {
	// 		version.BinaryVersion = "BUILT_FROM_SOURCE"

	// 		var result bool
	// 		err = client.Call("CliRpcCmd.IsMinCliVersion", "12.0.6", &result)
	// 		Expect(err).ToNot(HaveOccurred())

	// 		Expect(result).To(BeTrue())
	// 	})
	// })

	Describe(".SetPluginMetadata", func() {
		var (
			metadata *plugin.PluginMetadata
		)

		BeforeEach(func() {
			rpcService, err = NewRpcService(nil, nil, nil, api.RepositoryLocator{}, nil, nil, nil, rpc.DefaultServer)
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

	Describe(".GetOutputAndReset", func() {
		Context("success", func() {
			BeforeEach(func() {
				outputCapture := terminal.NewTeePrinter(os.Stdout)
				rpcService, err = NewRpcService(outputCapture, nil, nil, api.RepositoryLocator{}, cmdRunner.NewCommandRunner(), nil, nil, rpc.DefaultServer)
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

			It("should return the logs from the output capture", func() {
				client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
				Expect(err).ToNot(HaveOccurred())

				success := false

				oldStd := os.Stdout
				os.Stdout = nil
				client.Call("CliRpcCmd.CallCoreCommand", []string{"fake-command"}, &success)
				Expect(success).To(BeTrue())
				os.Stdout = oldStd

				var output []string
				client.Call("CliRpcCmd.GetOutputAndReset", false, &output)

				Expect(output).To(Equal([]string{"Requirement executed", "Command Executed"}))
			})
		})
	})

	Describe("disabling terminal output", func() {
		var terminalOutputSwitch *rpcfakes.FakeTerminalOutputSwitch

		BeforeEach(func() {
			terminalOutputSwitch = new(rpcfakes.FakeTerminalOutputSwitch)
			rpcService, err = NewRpcService(nil, terminalOutputSwitch, nil, api.RepositoryLocator{}, nil, nil, nil, rpc.DefaultServer)
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

		var runner *rpcfakes.FakeCommandRunner

		BeforeEach(func() {
			outputCapture := terminal.NewTeePrinter(os.Stdout)
			terminalOutputSwitch := terminal.NewTeePrinter(os.Stdout)

			runner = new(rpcfakes.FakeCommandRunner)
			rpcService, err = NewRpcService(outputCapture, terminalOutputSwitch, nil, api.RepositoryLocator{}, runner, nil, nil, rpc.DefaultServer)
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

		It("calls GetApp() with 'app' as argument", func() {
			result := plugin_models.GetAppModel{}
			err = client.Call("CliRpcCmd.GetApp", "fake-app", &result)

			Expect(err).To(HaveOccurred())
		})

	})

	Describe(".CallCoreCommand", func() {
		var runner *rpcfakes.FakeCommandRunner

		Context("success", func() {
			BeforeEach(func() {

				outputCapture := terminal.NewTeePrinter(os.Stdout)
				runner = new(rpcfakes.FakeCommandRunner)

				rpcService, err = NewRpcService(outputCapture, nil, nil, api.RepositoryLocator{}, runner, nil, nil, rpc.DefaultServer)
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

			It("is able to call a command", func() {
				client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
				Expect(err).ToNot(HaveOccurred())

				var success bool
				err = client.Call("CliRpcCmd.CallCoreCommand", []string{"fake-command3"}, &success)

				Expect(err).ToNot(HaveOccurred())
				Expect(runner.CommandCallCount()).To(Equal(1))

				_, _, pluginApiCall := runner.CommandArgsForCall(0)
				Expect(pluginApiCall).To(BeFalse())
			})
		})

		Context("fail", func() {
			BeforeEach(func() {
				outputCapture := terminal.NewTeePrinter(os.Stdout)
				rpcService, err = NewRpcService(outputCapture, nil, nil, api.RepositoryLocator{}, cmdRunner.NewCommandRunner(), nil, nil, rpc.DefaultServer)
				Expect(err).ToNot(HaveOccurred())

				err := rpcService.Start()
				Expect(err).ToNot(HaveOccurred())

				pingCli(rpcService.Port())
			})

			It("returns false in success if the command cannot be found", func() {
				client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
				Expect(err).ToNot(HaveOccurred())

				var success bool
				err = client.Call("CliRpcCmd.CallCoreCommand", []string{"not_a_cmd"}, &success)
				Expect(success).To(BeFalse())
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error if a command cannot parse provided flags", func() {
				client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
				Expect(err).ToNot(HaveOccurred())

				var success bool
				err = client.Call("CliRpcCmd.CallCoreCommand", []string{"fake-command", "-invalid_flag"}, &success)

				Expect(err).To(HaveOccurred())
				Expect(success).To(BeFalse())
			})

			It("recovers from a panic from any core command", func() {
				client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
				Expect(err).ToNot(HaveOccurred())

				var success bool
				err = client.Call("CliRpcCmd.CallCoreCommand", []string{"fake-command3"}, &success)

				Expect(success).To(BeFalse())
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
