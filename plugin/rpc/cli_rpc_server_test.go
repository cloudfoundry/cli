package rpc_test

import (
	"net"
	"net/rpc"
	"time"

	"github.com/cloudfoundry/cli/plugin"
	. "github.com/cloudfoundry/cli/plugin/rpc"
	io_helpers "github.com/cloudfoundry/cli/testhelpers/io"
	"github.com/codegangsta/cli"
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
			rpcService, err = NewRpcService(nil, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an err of another Rpc process is already registered", func() {
			_, err := NewRpcService(nil, nil)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe(".Stop", func() {
		BeforeEach(func() {
			rpcService, err = NewRpcService(nil, nil)
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
			rpcService, err = NewRpcService(nil, nil)
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
			metadata plugin.PluginMetadata
		)

		BeforeEach(func() {
			rpcService, err = NewRpcService(nil, nil)
			Expect(err).ToNot(HaveOccurred())

			err := rpcService.Start()
			Expect(err).ToNot(HaveOccurred())

			pingCli(rpcService.Port())

			client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
			Expect(err).ToNot(HaveOccurred())

			metadata = plugin.PluginMetadata{
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
			Expect(rpcService.RpcCmd.ReturnData.(plugin.PluginMetadata)).To(Equal(metadata))
		})
	})

	Describe(".GetOutputAndReset", func() {
		Context("success", func() {
			BeforeEach(func() {
				outputCapture := FakeOutputCapture{}
				rpcService, err = NewRpcService(nil, outputCapture)
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
				var output []string
				client.Call("CliRpcCmd.GetOutputAndReset", false, &output)

				Expect(output).To(Equal([]string{"hi from command"}))
			})
		})
	})

	Describe(".CallCoreCommand", func() {
		Context("success", func() {
			BeforeEach(func() {
				app := &cli.App{
					Commands: []cli.Command{
						{
							Name:        "test_cmd",
							Description: "test_cmd description",
							Usage:       "test_cmd usage",
							Action: func(context *cli.Context) {
								return
							},
						},
					},
				}

				outputCapture := FakeOutputCapture{}

				rpcService, err = NewRpcService(app, outputCapture)
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

			It("calls the code gangsta cli App command", func() {
				client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
				Expect(err).ToNot(HaveOccurred())

				var success bool
				err = client.Call("CliRpcCmd.CallCoreCommand", []string{"test_cmd"}, &success)

				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())
			})
		})

		Context("fail", func() {
			BeforeEach(func() {
				app := &cli.App{
					Commands: []cli.Command{
						{
							Name:        "test_cmd",
							Description: "test_cmd description",
							Usage:       "test_cmd usage",
							Action: func(context *cli.Context) {
								panic("ERROR")
							},
						},
					},
				}
				outputCapture := FakeOutputCapture{}
				rpcService, err = NewRpcService(app, outputCapture)
				Expect(err).ToNot(HaveOccurred())

				err := rpcService.Start()
				Expect(err).ToNot(HaveOccurred())

				pingCli(rpcService.Port())
			})

			It("returns false in success if the command cannot be found", func() {
				io_helpers.CaptureOutput(func() {
					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var success bool
					err = client.Call("CliRpcCmd.CallCoreCommand", []string{"not_a_cmd"}, &success)
					Expect(success).To(BeFalse())
					Expect(err).ToNot(HaveOccurred())
				})
			})

			It("returns an error if a command cannot parse provided flags", func() {
				io_helpers.CaptureOutput(func() {
					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var success bool
					err = client.Call("CliRpcCmd.CallCoreCommand", []string{"test_cmd", "-invalid_flag"}, &success)

					Expect(err).To(HaveOccurred())
					Expect(success).To(BeFalse())
				})
			})

			It("recovers from a panic from any core command", func() {
				client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
				Expect(err).ToNot(HaveOccurred())

				var success bool
				err = client.Call("CliRpcCmd.CallCoreCommand", []string{"test_cmd"}, &success)

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

type FakeOutputCapture struct{}

func (o FakeOutputCapture) GetOutputAndReset() []string {
	return []string{"hi from command"}
}
