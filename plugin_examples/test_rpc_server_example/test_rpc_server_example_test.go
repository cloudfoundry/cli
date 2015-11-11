package main_test

import (
	"encoding/json"
	"errors"
	"os/exec"

	. "github.com/cloudfoundry/cli/plugin_examples/test_rpc_server_example"

	"github.com/cloudfoundry/cli/testhelpers/rpc_server"
	fake_rpc_handlers "github.com/cloudfoundry/cli/testhelpers/rpc_server/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

const validPluginPath = "./test_rpc_server_example.exe"

var _ = Describe("App-Lister", func() {

	var (
		rpcHandlers *fake_rpc_handlers.FakeHandlers
		ts          *test_rpc_server.TestServer
		err         error
	)

	BeforeEach(func() {
		rpcHandlers = &fake_rpc_handlers.FakeHandlers{}
		ts, err = test_rpc_server.NewTestRpcServer(rpcHandlers)
		Ω(err).NotTo(HaveOccurred())

		err = ts.Start()
		Ω(err).NotTo(HaveOccurred())

		//set rpc.CallCoreCommand to a successful call
		//rpc.CallCoreCommand is used in both cliConnection.CliCommand() and
		//cliConnection.CliWithoutTerminalOutput()
		rpcHandlers.CallCoreCommandStub = func(_ []string, retVal *bool) error {
			*retVal = true
			return nil
		}

		//set rpc.GetOutputAndReset to return empty string; this is used by CliCommand()/CliWithoutTerminalOutput()
		rpcHandlers.GetOutputAndResetStub = func(_ bool, retVal *[]string) error {
			*retVal = []string{"{}"}
			return nil
		}
	})

	AfterEach(func() {
		ts.Stop()
	})

	Describe("list-apps", func() {
		Context("Option flags", func() {
			It("accept --started or --stopped as valid optional flag", func() {
				args := []string{ts.Port(), "list-apps", "--started"}
				session, err := gexec.Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
				session.Wait()
				Ω(err).NotTo(HaveOccurred())

				args = []string{ts.Port(), "list-apps", "--stopped"}
				session, err = gexec.Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
				session.Wait()
				Ω(err).NotTo(HaveOccurred())
			})

			It("raises error when invalid flag is provided", func() {
				args := []string{ts.Port(), "list-apps", "--invalid_flag"}
				session, err := gexec.Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
				session.Wait()
				Ω(err).NotTo(HaveOccurred())
				Ω(session).To(gbytes.Say("FAILED"))
				Ω(session).To(gbytes.Say("invalid_flag"))
			})
		})

		Context("Running the command", func() {
			Context("Curling v2/apps endpoint", func() {
				It("shows the endpoint it is curling", func() {
					rpcHandlers.ApiEndpointStub = func(_ string, retVal *string) error {
						*retVal = "api.example.com"
						return nil
					}

					args := []string{ts.Port(), "list-apps"}
					session, err := gexec.Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
					session.Wait()
					Ω(err).NotTo(HaveOccurred())
					Ω(session).To(gbytes.Say("api.example.com/v2/apps"))
				})

				It("raises an error when ApiEndpoint() returns an error", func() {
					rpcHandlers.ApiEndpointStub = func(_ string, retVal *string) error {
						*retVal = ""
						return errors.New("Bad bad error")
					}

					args := []string{ts.Port(), "list-apps"}
					session, err := gexec.Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
					session.Wait()
					Ω(err).NotTo(HaveOccurred())
					Ω(session).To(gbytes.Say("FAILED"))
					Ω(session).To(gbytes.Say("Bad bad error"))
					Ω(session.ExitCode()).To(Equal(1))
				})

				Context("when getting a list of apps", func() {
					Context("without option flag", func() {
						It("lists all apps", func() {
							rpcHandlers.GetOutputAndResetStub = func(_ bool, retVal *[]string) error {
								*retVal = []string{marshal(sampleApps())}
								return nil
							}

							args := []string{ts.Port(), "list-apps"}
							session, err := gexec.Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
							session.Wait()
							Ω(err).NotTo(HaveOccurred())
							Ω(session).To(gbytes.Say("app1"))
							Ω(session).To(gbytes.Say("app2"))
							Ω(session).To(gbytes.Say("app3"))
						})
					})

					Context("with --started", func() {
						It("lists only started apps", func() {
							rpcHandlers.GetOutputAndResetStub = func(_ bool, retVal *[]string) error {
								*retVal = []string{marshal(sampleApps())}
								return nil
							}

							args := []string{ts.Port(), "list-apps", "--started"}
							session, err := gexec.Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
							session.Wait()
							Ω(err).NotTo(HaveOccurred())
							Ω(session).To(gbytes.Say("app1"))
							Ω(session).To(gbytes.Say("app2"))
							Ω(session).ToNot(gbytes.Say("app3"))
						})
					})

					Context("with --stopped", func() {
						It("lists only stopped apps", func() {
							rpcHandlers.GetOutputAndResetStub = func(_ bool, retVal *[]string) error {
								*retVal = []string{marshal(sampleApps())}
								return nil
							}

							args := []string{ts.Port(), "list-apps", "--stopped"}
							session, err := gexec.Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
							session.Wait()
							Ω(err).NotTo(HaveOccurred())
							Ω(session).ToNot(gbytes.Say("app1"))
							Ω(session).ToNot(gbytes.Say("app2"))
							Ω(session).To(gbytes.Say("app3"))
						})
					})

					Context("when CliCommandWithoutTerminalOutput() returns an error", func() {
						It("notifies the user about the error", func() {
							rpcHandlers.CallCoreCommandStub = func(_ []string, retVal *bool) error {
								return errors.New("something went wrong")
							}

							args := []string{ts.Port(), "list-apps", "--stopped"}
							session, err := gexec.Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
							session.Wait()
							Ω(err).NotTo(HaveOccurred())
							Ω(session).To(gbytes.Say("FAILED"))
							Ω(session).To(gbytes.Say("something went wrong"))
						})
					})

					Context("when 'next url' is present in the JSON response", func() {
						BeforeEach(func() {
							count := 0
							rpcHandlers.GetOutputAndResetStub = func(_ bool, retVal *[]string) error {
								apps := sampleApps()
								if count == 0 {
									apps.NextUrl = "v2/apps?page=2"
									*retVal = []string{marshal(apps)}
									count++
								} else {
									apps.Resources = append(apps.Resources, AppModel{Entity: EntityModel{Name: "app4", State: "STARTED"}})
									*retVal = []string{marshal(apps)}
								}
								return nil
							}
						})

						It("follows and curl the next url", func() {
							args := []string{ts.Port(), "list-apps"}
							session, err := gexec.Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
							session.Wait()
							Ω(err).NotTo(HaveOccurred())
							Ω(rpcHandlers.CallCoreCommandCallCount()).To(Equal(2))

							params, _ := rpcHandlers.CallCoreCommandArgsForCall(0)
							Ω(params[1]).To(Equal("v2/apps"))

							params, _ = rpcHandlers.CallCoreCommandArgsForCall(1)
							Ω(params[1]).To(Equal("v2/apps?page=2"))
						})

						It("traverses through all pages and list all the apps", func() {
							args := []string{ts.Port(), "list-apps"}
							session, err := gexec.Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
							session.Wait()
							Ω(err).NotTo(HaveOccurred())
							Ω(session).To(gbytes.Say("app1"))
							Ω(session).To(gbytes.Say("app2"))
							Ω(session).To(gbytes.Say("app3"))
							Ω(session).To(gbytes.Say("app4"))
						})
					})
				})
			})
		})
	})
})

func sampleApps() AppsModel {
	allApps := AppsModel{
		Resources: []AppModel{
			AppModel{
				EntityModel{Name: "app1", State: "STARTED"},
			},
			AppModel{
				EntityModel{Name: "app2", State: "STARTED"},
			},
			AppModel{
				EntityModel{Name: "app3", State: "STOPPED"},
			},
		},
	}

	return allApps
}

func marshal(apps AppsModel) string {
	b, err := json.Marshal(apps)
	Ω(err).ToNot(HaveOccurred())

	return string(b)
}
