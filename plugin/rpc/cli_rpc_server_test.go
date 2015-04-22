package rpc_test

import (
	"net"
	"net/rpc"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal/fakes"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/cli/plugin/models"
	. "github.com/cloudfoundry/cli/plugin/rpc"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
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
			rpcService, err = NewRpcService(nil, nil, nil, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an err of another Rpc process is already registered", func() {
			_, err := NewRpcService(nil, nil, nil, nil)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe(".Stop", func() {
		BeforeEach(func() {
			rpcService, err = NewRpcService(nil, nil, nil, nil)
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
			rpcService, err = NewRpcService(nil, nil, nil, nil)
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
			rpcService, err = NewRpcService(nil, nil, nil, nil)
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
				outputCapture := &fakes.FakeOutputCapture{}
				outputCapture.GetOutputAndResetReturns([]string{"hi from command"})
				rpcService, err = NewRpcService(nil, outputCapture, nil, nil)
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

	Describe("disabling terminal output", func() {
		var terminalOutputSwitch *fakes.FakeTerminalOutputSwitch

		BeforeEach(func() {
			terminalOutputSwitch = &fakes.FakeTerminalOutputSwitch{}
			rpcService, err = NewRpcService(nil, nil, terminalOutputSwitch, nil)
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

				outputCapture := &fakes.FakeOutputCapture{}

				rpcService, err = NewRpcService(app, outputCapture, nil, nil)
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

		Describe("CLI Config object methods", func() {
			var (
				config core_config.Repository
			)

			BeforeEach(func() {
				config = testconfig.NewRepositoryWithDefaults()
			})

			AfterEach(func() {
				rpcService.Stop()

				//give time for server to stop
				time.Sleep(50 * time.Millisecond)
			})

			Context(".GetCurrentOrg", func() {
				BeforeEach(func() {
					config.SetOrganizationFields(models.OrganizationFields{
						Guid: "test-guid",
						Name: "test-org",
						QuotaDefinition: models.QuotaFields{
							Guid:                    "guid123",
							Name:                    "quota123",
							MemoryLimit:             128,
							InstanceMemoryLimit:     16,
							RoutesLimit:             5,
							ServicesLimit:           6,
							NonBasicServicesAllowed: true,
						},
					})

					rpcService, err = NewRpcService(nil, nil, nil, config)
					err := rpcService.Start()
					Expect(err).ToNot(HaveOccurred())

					pingCli(rpcService.Port())
				})

				It("populates the plugin Organization object with the current org settings in config", func() {
					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var org plugin_models.Organization
					err = client.Call("CliRpcCmd.GetCurrentOrg", "", &org)

					Expect(err).ToNot(HaveOccurred())
					Expect(org.Name).To(Equal("test-org"))
					Expect(org.Guid).To(Equal("test-guid"))
					Expect(org.QuotaDefinition.Guid).To(Equal("guid123"))
					Expect(org.QuotaDefinition.Name).To(Equal("quota123"))
					Expect(org.QuotaDefinition.MemoryLimit).To(Equal(int64(128)))
					Expect(org.QuotaDefinition.InstanceMemoryLimit).To(Equal(int64(16)))
					Expect(org.QuotaDefinition.RoutesLimit).To(Equal(5))
					Expect(org.QuotaDefinition.ServicesLimit).To(Equal(6))
					Expect(org.QuotaDefinition.NonBasicServicesAllowed).To(BeTrue())
				})
			})

			Context(".GetCurrentSpace", func() {
				BeforeEach(func() {
					config.SetSpaceFields(models.SpaceFields{
						Guid: "space-guid",
						Name: "space-name",
					})

					rpcService, err = NewRpcService(nil, nil, nil, config)
					err := rpcService.Start()
					Expect(err).ToNot(HaveOccurred())

					pingCli(rpcService.Port())
				})

				It("populates the plugin Space object with the current space settings in config", func() {
					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var space plugin_models.Space
					err = client.Call("CliRpcCmd.GetCurrentSpace", "", &space)

					Expect(err).ToNot(HaveOccurred())
					Expect(space.Name).To(Equal("space-name"))
					Expect(space.Guid).To(Equal("space-guid"))
				})
			})

			Context(".Username, .UserGuid, .UserEmail", func() {
				BeforeEach(func() {
					rpcService, err = NewRpcService(nil, nil, nil, config)
					err := rpcService.Start()
					Expect(err).ToNot(HaveOccurred())

					pingCli(rpcService.Port())
				})

				It("returns username, user guid and user email", func() {
					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var result string
					err = client.Call("CliRpcCmd.Username", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal("my-user"))

					err = client.Call("CliRpcCmd.UserGuid", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal("my-user-guid"))

					err = client.Call("CliRpcCmd.UserEmail", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal("my-user-email"))
				})
			})

			Context(".IsSSLDisabled", func() {
				BeforeEach(func() {
					rpcService, err = NewRpcService(nil, nil, nil, config)
					err := rpcService.Start()
					Expect(err).ToNot(HaveOccurred())

					pingCli(rpcService.Port())
				})

				It("returns the IsSSLDisabled setting in config", func() {
					config.SetSSLDisabled(true)
					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var result bool
					err = client.Call("CliRpcCmd.IsSSLDisabled", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(BeTrue())
				})
			})

			Context(".IsLoggedIn", func() {
				BeforeEach(func() {
					rpcService, err = NewRpcService(nil, nil, nil, config)
					err := rpcService.Start()
					Expect(err).ToNot(HaveOccurred())

					pingCli(rpcService.Port())
				})

				It("returns the IsLoggedIn setting in config", func() {
					config.SetAccessToken("Logged-In-Token")
					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var result bool
					err = client.Call("CliRpcCmd.IsLoggedIn", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(BeTrue())
				})
			})

			Context(".HasOrganization and .HasSpace ", func() {
				BeforeEach(func() {
					rpcService, err = NewRpcService(nil, nil, nil, config)
					err := rpcService.Start()
					Expect(err).ToNot(HaveOccurred())

					pingCli(rpcService.Port())
				})

				It("returns the HasOrganization() and HasSpace() setting in config", func() {
					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var result bool
					err = client.Call("CliRpcCmd.HasOrganization", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(BeTrue())

					err = client.Call("CliRpcCmd.HasSpace", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(BeTrue())
				})
			})

			Context(".LoggregatorEndpoint and .DopplerEndpoint ", func() {
				BeforeEach(func() {
					rpcService, err = NewRpcService(nil, nil, nil, config)
					err := rpcService.Start()
					Expect(err).ToNot(HaveOccurred())

					pingCli(rpcService.Port())
				})

				It("returns the LoggregatorEndpoint() and DopplerEndpoint() setting in config", func() {
					config.SetLoggregatorEndpoint("loggregator-endpoint-sample")
					config.SetDopplerEndpoint("doppler-endpoint-sample")

					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var result string
					err = client.Call("CliRpcCmd.LoggregatorEndpoint", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal("loggregator-endpoint-sample"))

					err = client.Call("CliRpcCmd.DopplerEndpoint", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal("doppler-endpoint-sample"))
				})
			})

			Context(".ApiEndpoint, .ApiVersion and .HasAPIEndpoint", func() {
				BeforeEach(func() {
					rpcService, err = NewRpcService(nil, nil, nil, config)
					err := rpcService.Start()
					Expect(err).ToNot(HaveOccurred())

					pingCli(rpcService.Port())
				})

				It("returns the ApiEndpoint(), ApiVersion() and HasAPIEndpoint() setting in config", func() {
					config.SetApiVersion("v1.1.1")
					config.SetApiEndpoint("www.fake-domain.com")

					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var result string
					err = client.Call("CliRpcCmd.ApiEndpoint", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal("www.fake-domain.com"))

					err = client.Call("CliRpcCmd.ApiVersion", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal("v1.1.1"))

					var exists bool
					err = client.Call("CliRpcCmd.HasAPIEndpoint", "", &exists)
					Expect(err).ToNot(HaveOccurred())
					Expect(exists).To(BeTrue())

				})
			})

			Context(".AccessToken", func() {
				BeforeEach(func() {
					rpcService, err = NewRpcService(nil, nil, nil, config)
					err := rpcService.Start()
					Expect(err).ToNot(HaveOccurred())

					pingCli(rpcService.Port())
				})

				It("returns the LoggregatorEndpoint() and DopplerEndpoint() setting in config", func() {
					config.SetAccessToken("fake-access-token")

					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var result string
					err = client.Call("CliRpcCmd.AccessToken", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal("fake-access-token"))
				})
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
				outputCapture := &fakes.FakeOutputCapture{}
				rpcService, err = NewRpcService(app, outputCapture, nil, nil)
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
