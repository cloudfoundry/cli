package application_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	cmdFakes "github.com/cloudfoundry/cli/cf/commands/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testssh "github.com/cloudfoundry/cli/cf/ssh/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSH command", func() {
	var (
		ui *testterm.FakeUI

		sshCodeGetter         *cmdFakes.FakeSSHCodeGetter
		originalSSHCodeGetter command_registry.Command

		requirementsFactory *testreq.FakeReqFactory
		configRepo          core_config.Repository
		deps                command_registry.Dependency
		ccGateway           net.Gateway

		fakeSecureShell *testssh.FakeSecureShell
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		deps.Gateways = make(map[string]net.Gateway)

		//save original command and restore later
		originalSSHCodeGetter = command_registry.Commands.FindCommand("ssh-code")

		sshCodeGetter = &cmdFakes.FakeSSHCodeGetter{}

		//setup fakes to correctly interact with command_registry
		sshCodeGetter.SetDependencyStub = func(_ command_registry.Dependency, _ bool) command_registry.Command {
			return sshCodeGetter
		}
		sshCodeGetter.MetaDataReturns(command_registry.CommandMetadata{Name: "ssh-code"})
	})

	AfterEach(func() {
		//restore original command
		command_registry.Register(originalSSHCodeGetter)
	})

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo

		//inject fake 'sshCodeGetter' into registry
		command_registry.Register(sshCodeGetter)

		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("ssh").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("ssh", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("Requirements", func() {
		It("fails with usage when not provided exactly one arg", func() {
			requirementsFactory.LoginSuccess = true

			runCommand()
			Ω(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
		})

		It("fails requirements when not logged in", func() {
			Ω(runCommand("my-app")).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = false
			Ω(runCommand("my-app")).To(BeFalse())
		})

		It("fails if a application is not found", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
			requirementsFactory.ApplicationFails = true

			Ω(runCommand("my-app")).To(BeFalse())
		})

		Describe("Flag options", func() {
			var args []string

			BeforeEach(func() {
				requirementsFactory.LoginSuccess = true
				requirementsFactory.TargetedSpaceSuccess = true
			})

			Context("when an -i flag is provided", func() {
				BeforeEach(func() {
					args = append(args, "app-name")
				})

				Context("with a negative integer argument", func() {
					BeforeEach(func() {
						args = append(args, "-i", "-3")
					})

					It("returns an error", func() {
						Ω(runCommand(args...)).To(BeFalse())
						Ω(ui.Outputs).To(ContainSubstrings(
							[]string{"Incorrect Usage", "cannot be negative"},
						))
					})
				})
			})
		})

		Describe("SSHOptions", func() {
			Context("when an error is returned during initialization", func() {
				It("shows error and prints command usage", func() {
					Ω(runCommand("app_name", "-L", "[9999:localhost...")).To(BeFalse())
					Ω(ui.Outputs).To(ContainSubstrings(
						[]string{"Incorrect Usage"},
						[]string{"USAGE:"},
					))
				})
			})
		})

	})

	Describe("ssh", func() {
		var (
			currentApp models.Application
		)

		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
			currentApp = models.Application{}
			currentApp.Name = "my-app"
			currentApp.State = "started"
			currentApp.Guid = "my-app-guid"
			currentApp.EnableSsh = true
			currentApp.Diego = true

			requirementsFactory.Application = currentApp
		})

		Describe("Error getting required info to run ssh", func() {
			var (
				testServer *httptest.Server
				handler    *testnet.TestHandler
			)

			AfterEach(func() {
				testServer.Close()
			})

			Context("error when getting SSH info from /v2/info", func() {
				BeforeEach(func() {
					getRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
						Method: "GET",
						Path:   "/v2/info",
						Response: testnet.TestResponse{
							Status: http.StatusNotFound,
							Body:   `{}`,
						},
					})

					testServer, handler = testnet.NewServer([]testnet.TestRequest{getRequest})
					configRepo.SetApiEndpoint(testServer.URL)
					ccGateway = net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{})
					deps.Gateways["cloud-controller"] = ccGateway
				})

				It("notifies users", func() {
					runCommand("my-app")

					Expect(handler).To(HaveAllRequestsCalled())
					Ω(ui.Outputs).To(ContainSubstrings(
						[]string{"Error getting SSH info", "404"},
					))
				})
			})

			Context("error when getting oauth token", func() {
				BeforeEach(func() {
					sshCodeGetter.GetReturns("", errors.New("auth api error"))

					getRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
						Method: "GET",
						Path:   "/v2/info",
						Response: testnet.TestResponse{
							Status: http.StatusOK,
							Body:   `{}`,
						},
					})

					testServer, handler = testnet.NewServer([]testnet.TestRequest{getRequest})
					configRepo.SetApiEndpoint(testServer.URL)
					ccGateway = net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{})
					deps.Gateways["cloud-controller"] = ccGateway
				})

				It("notifies users", func() {
					runCommand("my-app")

					Expect(handler).To(HaveAllRequestsCalled())
					Ω(ui.Outputs).To(ContainSubstrings(
						[]string{"Error getting one time auth code", "auth api error"},
					))
				})
			})
		})

		Describe("Connecting to ssh server", func() {
			var testServer *httptest.Server

			AfterEach(func() {
				testServer.Close()
			})

			BeforeEach(func() {
				fakeSecureShell = &testssh.FakeSecureShell{}

				deps.WilecardDependency = fakeSecureShell

				getRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/info",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body:   getInfoResponseBody,
					},
				})

				testServer, _ = testnet.NewServer([]testnet.TestRequest{getRequest})
				configRepo.SetApiEndpoint(testServer.URL)
				ccGateway = net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{})
				deps.Gateways["cloud-controller"] = ccGateway
			})

			Context("Error when connecting", func() {
				It("notifies users", func() {
					fakeSecureShell.ConnectReturns(errors.New("dial errorrr"))

					runCommand("my-app")

					Ω(ui.Outputs).To(ContainSubstrings(
						[]string{"Error opening SSH connection", "dial error"},
					))
				})
			})

			Context("Error port forwarding when -L is provided", func() {
				It("notifies users", func() {
					fakeSecureShell.LocalPortForwardReturns(errors.New("listen error"))

					runCommand("my-app", "-L", "8000:localhost:8000")

					Ω(ui.Outputs).To(ContainSubstrings(
						[]string{"Error forwarding port", "listen error"},
					))
				})
			})

			Context("when -N is provided", func() {
				It("calls secureShell.Wait()", func() {
					fakeSecureShell.ConnectReturns(nil)
					fakeSecureShell.LocalPortForwardReturns(nil)

					runCommand("my-app", "-N")

					Ω(fakeSecureShell.WaitCallCount()).To(Equal(1))
				})
			})

			Context("when -N is provided", func() {
				It("calls secureShell.InteractiveSession()", func() {
					fakeSecureShell.ConnectReturns(nil)
					fakeSecureShell.LocalPortForwardReturns(nil)

					runCommand("my-app", "-k")

					Ω(fakeSecureShell.InteractiveSessionCallCount()).To(Equal(1))
				})
			})

			Context("when Wait() or InteractiveSession() returns error", func() {

				It("notifities users", func() {
					fakeSecureShell.ConnectReturns(nil)
					fakeSecureShell.LocalPortForwardReturns(nil)

					fakeSecureShell.InteractiveSessionReturns(errors.New("ssh exit error"))
					runCommand("my-app", "-k")

					Ω(ui.Outputs).To(ContainSubstrings(
						[]string{"ssh exit error"},
					))
				})
			})
		})
	})
})

const getInfoResponseBody string = `
{
   "name": "vcap",
   "build": "2222",
   "support": "http://support.cloudfoundry.com",
   "version": 2,
   "description": "Cloud Foundry sponsored by ABC",
   "authorization_endpoint": "https://login.run.abc.com",
   "token_endpoint": "https://uaa.run.abc.com",
   "min_cli_version": null,
   "min_recommended_cli_version": null,
   "api_version": "2.35.0",
   "app_ssh_endpoint": "ssh.run.pivotal.io:2222",
   "app_ssh_host_key_fingerprint": "11:11:11:11:11:11:11:11:11:11:11:11:11:11:11:11",
   "logging_endpoint": "wss://loggregator.run.abc.com:443",
   "doppler_logging_endpoint": "wss://doppler.run.abc.com:443",
   "user": "6e477566-ac8d-4653-98c6-d319595ec7b0"
}`
