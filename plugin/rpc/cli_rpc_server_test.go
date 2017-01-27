package rpc_test

import (
	"errors"
	"net"
	"net/rpc"
	"os"
	"time"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/authentication/authenticationfakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/plugin/models"
	. "code.cloudfoundry.org/cli/plugin/rpc"
	cmdRunner "code.cloudfoundry.org/cli/plugin/rpc"
	. "code.cloudfoundry.org/cli/plugin/rpc/fakecommand"
	"code.cloudfoundry.org/cli/plugin/rpc/rpcfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
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

			Expect(err).ToNot(HaveOccurred())
			Expect(runner.CommandCallCount()).To(Equal(1))
			arg1, _, pluginApiCall := runner.CommandArgsForCall(0)
			Expect(arg1[0]).To(Equal("app"))
			Expect(arg1[1]).To(Equal("fake-app"))
			Expect(pluginApiCall).To(BeTrue())
		})

		It("calls GetOrg() with 'my-org' as argument", func() {
			result := plugin_models.GetOrg_Model{}
			err = client.Call("CliRpcCmd.GetOrg", "my-org", &result)

			Expect(err).ToNot(HaveOccurred())
			Expect(runner.CommandCallCount()).To(Equal(1))
			arg1, _, pluginApiCall := runner.CommandArgsForCall(0)
			Expect(arg1[0]).To(Equal("org"))
			Expect(arg1[1]).To(Equal("my-org"))
			Expect(pluginApiCall).To(BeTrue())
		})

		It("calls GetSpace() with 'my-space' as argument", func() {
			result := plugin_models.GetSpace_Model{}
			err = client.Call("CliRpcCmd.GetSpace", "my-space", &result)

			Expect(err).ToNot(HaveOccurred())
			Expect(runner.CommandCallCount()).To(Equal(1))
			arg1, _, pluginApiCall := runner.CommandArgsForCall(0)
			Expect(arg1[0]).To(Equal("space"))
			Expect(arg1[1]).To(Equal("my-space"))
			Expect(pluginApiCall).To(BeTrue())
		})

		It("calls GetApps() ", func() {
			result := []plugin_models.GetAppsModel{}
			err = client.Call("CliRpcCmd.GetApps", "", &result)

			Expect(err).ToNot(HaveOccurred())
			Expect(runner.CommandCallCount()).To(Equal(1))
			arg1, _, pluginApiCall := runner.CommandArgsForCall(0)
			Expect(arg1[0]).To(Equal("apps"))
			Expect(pluginApiCall).To(BeTrue())
		})

		It("calls GetOrgs() ", func() {
			result := []plugin_models.GetOrgs_Model{}
			err = client.Call("CliRpcCmd.GetOrgs", "", &result)

			Expect(err).ToNot(HaveOccurred())
			Expect(runner.CommandCallCount()).To(Equal(1))
			arg1, _, pluginApiCall := runner.CommandArgsForCall(0)
			Expect(arg1[0]).To(Equal("orgs"))
			Expect(pluginApiCall).To(BeTrue())
		})

		It("calls GetServices() ", func() {
			result := []plugin_models.GetServices_Model{}
			err = client.Call("CliRpcCmd.GetServices", "", &result)

			Expect(err).ToNot(HaveOccurred())
			Expect(runner.CommandCallCount()).To(Equal(1))
			arg1, _, pluginApiCall := runner.CommandArgsForCall(0)
			Expect(arg1[0]).To(Equal("services"))
			Expect(pluginApiCall).To(BeTrue())
		})

		It("calls GetSpaces() ", func() {
			result := []plugin_models.GetSpaces_Model{}
			err = client.Call("CliRpcCmd.GetSpaces", "", &result)

			Expect(err).ToNot(HaveOccurred())
			Expect(runner.CommandCallCount()).To(Equal(1))
			arg1, _, pluginApiCall := runner.CommandArgsForCall(0)
			Expect(arg1[0]).To(Equal("spaces"))
			Expect(pluginApiCall).To(BeTrue())
		})

		It("calls GetOrgUsers() ", func() {
			result := []plugin_models.GetOrgUsers_Model{}
			args := []string{"orgName1", "-a"}
			err = client.Call("CliRpcCmd.GetOrgUsers", args, &result)

			Expect(err).ToNot(HaveOccurred())
			Expect(runner.CommandCallCount()).To(Equal(1))
			arg1, _, pluginApiCall := runner.CommandArgsForCall(0)
			Expect(arg1[0]).To(Equal("org-users"))
			Expect(pluginApiCall).To(BeTrue())
		})

		It("calls GetSpaceUsers() ", func() {
			result := []plugin_models.GetSpaceUsers_Model{}
			args := []string{"orgName1", "spaceName1"}
			err = client.Call("CliRpcCmd.GetSpaceUsers", args, &result)

			Expect(err).ToNot(HaveOccurred())
			Expect(runner.CommandCallCount()).To(Equal(1))
			arg1, _, pluginApiCall := runner.CommandArgsForCall(0)
			Expect(arg1[0]).To(Equal("space-users"))
			Expect(pluginApiCall).To(BeTrue())
		})

		It("calls GetService() with 'serviceInstance' as argument", func() {
			result := plugin_models.GetService_Model{}
			err = client.Call("CliRpcCmd.GetService", "fake-service-instance", &result)

			Expect(err).ToNot(HaveOccurred())
			Expect(runner.CommandCallCount()).To(Equal(1))
			arg1, _, pluginApiCall := runner.CommandArgsForCall(0)
			Expect(arg1[0]).To(Equal("service"))
			Expect(arg1[1]).To(Equal("fake-service-instance"))
			Expect(pluginApiCall).To(BeTrue())
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

		Describe("CLI Config object methods", func() {
			var (
				config coreconfig.Repository
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
						GUID: "test-guid",
						Name: "test-org",
						QuotaDefinition: models.QuotaFields{
							GUID:                    "guid123",
							Name:                    "quota123",
							MemoryLimit:             128,
							InstanceMemoryLimit:     16,
							RoutesLimit:             5,
							ServicesLimit:           6,
							NonBasicServicesAllowed: true,
						},
					})

					rpcService, err = NewRpcService(nil, nil, config, api.RepositoryLocator{}, nil, nil, nil, rpc.DefaultServer)
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
				})
			})

			Context(".GetCurrentSpace", func() {
				BeforeEach(func() {
					config.SetSpaceFields(models.SpaceFields{
						GUID: "space-guid",
						Name: "space-name",
					})

					rpcService, err = NewRpcService(nil, nil, config, api.RepositoryLocator{}, nil, nil, nil, rpc.DefaultServer)
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
					rpcService, err = NewRpcService(nil, nil, config, api.RepositoryLocator{}, nil, nil, nil, rpc.DefaultServer)
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
					rpcService, err = NewRpcService(nil, nil, config, api.RepositoryLocator{}, nil, nil, nil, rpc.DefaultServer)
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
					rpcService, err = NewRpcService(nil, nil, config, api.RepositoryLocator{}, nil, nil, nil, rpc.DefaultServer)
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
					rpcService, err = NewRpcService(nil, nil, config, api.RepositoryLocator{}, nil, nil, nil, rpc.DefaultServer)
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
					rpcService, err = NewRpcService(nil, nil, config, api.RepositoryLocator{}, nil, nil, nil, rpc.DefaultServer)
					err := rpcService.Start()
					Expect(err).ToNot(HaveOccurred())

					pingCli(rpcService.Port())
				})

				It("returns the LoggregatorEndpoint() and DopplerEndpoint() setting in config", func() {
					config.SetDopplerEndpoint("doppler-endpoint-sample")

					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var result string
					err = client.Call("CliRpcCmd.LoggregatorEndpoint", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(""))

					err = client.Call("CliRpcCmd.DopplerEndpoint", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal("doppler-endpoint-sample"))
				})
			})

			Context(".ApiEndpoint, .ApiVersion and .HasAPIEndpoint", func() {
				BeforeEach(func() {
					rpcService, err = NewRpcService(nil, nil, config, api.RepositoryLocator{}, nil, nil, nil, rpc.DefaultServer)
					err := rpcService.Start()
					Expect(err).ToNot(HaveOccurred())

					pingCli(rpcService.Port())
				})

				It("returns the ApiEndpoint(), ApiVersion() and HasAPIEndpoint() setting in config", func() {
					config.SetAPIVersion("v1.1.1")
					config.SetAPIEndpoint("www.fake-domain.com")

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
				var authRepo *authenticationfakes.FakeRepository

				BeforeEach(func() {
					authRepo = new(authenticationfakes.FakeRepository)
					locator := api.RepositoryLocator{}
					locator = locator.SetAuthenticationRepository(authRepo)

					rpcService, err = NewRpcService(nil, nil, config, locator, nil, nil, nil, rpc.DefaultServer)
					err := rpcService.Start()
					Expect(err).ToNot(HaveOccurred())

					pingCli(rpcService.Port())
				})

				It("refreshes the token", func() {
					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var result string
					err = client.Call("CliRpcCmd.AccessToken", "", &result)
					Expect(err).ToNot(HaveOccurred())

					Expect(authRepo.RefreshAuthTokenCallCount()).To(Equal(1))
				})

				It("returns the access token", func() {
					authRepo.RefreshAuthTokenReturns("fake-access-token", nil)

					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var result string
					err = client.Call("CliRpcCmd.AccessToken", "", &result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal("fake-access-token"))
				})

				It("returns the error from refreshing the access token", func() {
					authRepo.RefreshAuthTokenReturns("", errors.New("refresh error"))

					client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
					Expect(err).ToNot(HaveOccurred())

					var result string
					err = client.Call("CliRpcCmd.AccessToken", "", &result)
					Expect(err.Error()).To(Equal("refresh error"))
				})
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
