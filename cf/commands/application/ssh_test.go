package application_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/commandsfakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"code.cloudfoundry.org/cli/cf/ssh/sshfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSH command", func() {
	var (
		ui *testterm.FakeUI

		sshCodeGetter         *commandsfakes.FakeSSHCodeGetter
		originalSSHCodeGetter commandregistry.Command

		requirementsFactory *requirementsfakes.FakeFactory
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
		ccGateway           net.Gateway

		fakeSecureShell *sshfakes.FakeSecureShell
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		deps.Gateways = make(map[string]net.Gateway)

		//save original command and restore later
		originalSSHCodeGetter = commandregistry.Commands.FindCommand("ssh-code")

		sshCodeGetter = new(commandsfakes.FakeSSHCodeGetter)

		//setup fakes to correctly interact with commandregistry
		sshCodeGetter.SetDependencyStub = func(_ commandregistry.Dependency, _ bool) commandregistry.Command {
			return sshCodeGetter
		}
		sshCodeGetter.MetaDataReturns(commandregistry.CommandMetadata{Name: "ssh-code"})
	})

	AfterEach(func() {
		//restore original command
		commandregistry.Register(originalSSHCodeGetter)
	})

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo

		//inject fake 'sshCodeGetter' into registry
		commandregistry.Register(sshCodeGetter)

		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("ssh").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("ssh", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("Requirements", func() {
		It("fails with usage when not provided exactly one arg", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})

			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))

		})

		It("fails requirements when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("my-app")).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "not targeting space"})
			Expect(runCommand("my-app")).To(BeFalse())
		})

		It("fails if a application is not found", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
			applicationReq := new(requirementsfakes.FakeApplicationRequirement)
			applicationReq.ExecuteReturns(errors.New("no app"))
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)

			Expect(runCommand("my-app")).To(BeFalse())
		})

		Describe("Flag options", func() {
			var args []string

			BeforeEach(func() {
				requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
				requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
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
						Expect(runCommand(args...)).To(BeFalse())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"Incorrect Usage", "cannot be negative"},
						))

					})
				})
			})
		})

		Describe("SSHOptions", func() {
			Context("when an error is returned during initialization", func() {
				It("shows error and prints command usage", func() {
					Expect(runCommand("app_name", "-L", "[9999:localhost...")).To(BeFalse())
					Expect(ui.Outputs()).To(ContainSubstrings(
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
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
			currentApp = models.Application{}
			currentApp.Name = "my-app"
			currentApp.State = "started"
			currentApp.GUID = "my-app-guid"
			currentApp.EnableSSH = true
			currentApp.Diego = true

			applicationReq := new(requirementsfakes.FakeApplicationRequirement)
			applicationReq.GetApplicationReturns(currentApp)
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)
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
					getRequest := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
						Method: "GET",
						Path:   "/v2/info",
						Response: testnet.TestResponse{
							Status: http.StatusNotFound,
							Body:   `{}`,
						},
					})

					testServer, handler = testnet.NewServer([]testnet.TestRequest{getRequest})
					configRepo.SetAPIEndpoint(testServer.URL)
					ccGateway = net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{}, new(tracefakes.FakePrinter), "")
					deps.Gateways["cloud-controller"] = ccGateway
				})

				It("notifies users", func() {
					runCommand("my-app")

					Expect(handler).To(HaveAllRequestsCalled())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Error getting SSH info", "404"},
					))

				})
			})

			Context("error when getting oauth token", func() {
				BeforeEach(func() {
					sshCodeGetter.GetReturns("", errors.New("auth api error"))

					getRequest := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
						Method: "GET",
						Path:   "/v2/info",
						Response: testnet.TestResponse{
							Status: http.StatusOK,
							Body:   `{}`,
						},
					})

					testServer, handler = testnet.NewServer([]testnet.TestRequest{getRequest})
					configRepo.SetAPIEndpoint(testServer.URL)
					ccGateway = net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{}, new(tracefakes.FakePrinter), "")
					deps.Gateways["cloud-controller"] = ccGateway
				})

				It("notifies users", func() {
					runCommand("my-app")

					Expect(handler).To(HaveAllRequestsCalled())
					Expect(ui.Outputs()).To(ContainSubstrings(
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
				fakeSecureShell = new(sshfakes.FakeSecureShell)

				deps.WildcardDependency = fakeSecureShell

				getRequest := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/info",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body:   getInfoResponseBody,
					},
				})

				testServer, _ = testnet.NewServer([]testnet.TestRequest{getRequest})
				configRepo.SetAPIEndpoint(testServer.URL)
				ccGateway = net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{}, new(tracefakes.FakePrinter), "")
				deps.Gateways["cloud-controller"] = ccGateway
			})

			Context("Error when connecting", func() {
				It("notifies users", func() {
					fakeSecureShell.ConnectReturns(errors.New("dial errorrr"))

					runCommand("my-app")

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Error opening SSH connection", "dial error"},
					))

				})
			})

			Context("Error port forwarding when -L is provided", func() {
				It("notifies users", func() {
					fakeSecureShell.LocalPortForwardReturns(errors.New("listen error"))

					runCommand("my-app", "-L", "8000:localhost:8000")

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Error forwarding port", "listen error"},
					))

				})
			})

			Context("when -N is provided", func() {
				It("calls secureShell.Wait()", func() {
					fakeSecureShell.ConnectReturns(nil)
					fakeSecureShell.LocalPortForwardReturns(nil)

					runCommand("my-app", "-N")

					Expect(fakeSecureShell.WaitCallCount()).To(Equal(1))
				})
			})

			Context("when -N is provided", func() {
				It("calls secureShell.InteractiveSession()", func() {
					fakeSecureShell.ConnectReturns(nil)
					fakeSecureShell.LocalPortForwardReturns(nil)

					runCommand("my-app", "-k")

					Expect(fakeSecureShell.InteractiveSessionCallCount()).To(Equal(1))
				})
			})

			Context("when Wait() or InteractiveSession() returns error", func() {

				It("notifities users", func() {
					fakeSecureShell.ConnectReturns(nil)
					fakeSecureShell.LocalPortForwardReturns(nil)

					fakeSecureShell.InteractiveSessionReturns(errors.New("ssh exit error"))
					runCommand("my-app", "-k")

					Expect(ui.Outputs()).To(ContainSubstrings(
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
