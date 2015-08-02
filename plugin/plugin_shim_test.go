package plugin_test

import (
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/cli/testhelpers/rpc_server"
	fake_rpc_handlers "github.com/cloudfoundry/cli/testhelpers/rpc_server/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Command", func() {
	var (
		validPluginPath = filepath.Join("..", "fixtures", "plugins", "test_1.exe")
	)

	Describe(".Start", func() {
		It("Exits with status 1 if it cannot ping the host port passed as an argument", func() {
			args := []string{"0", "0"}
			session, err := Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 2).Should(Exit(1))
		})

		Context("Executing plugins with '.Start()'", func() {
			var (
				rpcHandlers *fake_rpc_handlers.FakeHandlers
				ts          *test_rpc_server.TestServer
				err         error
			)

			BeforeEach(func() {
				rpcHandlers = &fake_rpc_handlers.FakeHandlers{}
				ts, err = test_rpc_server.NewTestRpcServer(rpcHandlers)
				Expect(err).NotTo(HaveOccurred())

				err = ts.Start()
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				ts.Stop()
			})

			Context("checking MinCliVersion", func() {
				It("it calls rpc cmd 'IsMinCliVersion' if plugin metadata 'MinCliVersion' is set", func() {
					args := []string{ts.Port(), "0"}
					session, err := Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())

					session.Wait()

					Expect(rpcHandlers.IsMinCliVersionCallCount()).To(Equal(1))
				})

				It("notifies the user 'min cli version is not met'", func() {
					rpcHandlers.IsMinCliVersionStub = func(_ string, result *bool) error {
						*result = false
						return nil
					}

					args := []string{ts.Port(), "0"}
					session, err := Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())

					session.Wait()

					Expect(session).To(gbytes.Say("Minimum CLI version 5.0.0 is required to run this plugin command"))
				})
			})
		})
	})

	Describe("MinCliVersionStr", func() {
		It("returns a string representation of VersionType{}", func() {
			version := plugin.VersionType{
				Major: 1,
				Minor: 2,
				Build: 3,
			}

			str := plugin.MinCliVersionStr(version)
			Expect(str).To(Equal("1.2.3"))
		})

		It("returns a empty string if no field in VersionType is set", func() {
			version := plugin.VersionType{}

			str := plugin.MinCliVersionStr(version)
			Expect(str).To(Equal(""))
		})

		It("uses '0' as return value for field that is not set", func() {
			version := plugin.VersionType{
				Build: 5,
			}

			str := plugin.MinCliVersionStr(version)
			Expect(str).To(Equal("0.0.5"))
		})

	})
})
