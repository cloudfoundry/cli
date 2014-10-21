package plugin_test

import (
	"net"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Command", func() {
	var (
		err             error
		validPluginPath = filepath.Join("..", "fixtures", "plugins", "test_1.exe")
	)

	obtainPort := func() string {
		//assign 0 to port to get a random open port
		l, _ := net.Listen("tcp", "127.0.0.1:0")
		port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
		l.Close()
		return port
	}

	Describe(".Start", func() {
		It("prints a warning if a plugin does not implement the rpc interface", func() {
			//This would seem like a valid test, but the plugin itself will not compile
		})

		It("Exits with status 1 if it cannot ping the host port passed as an argument", func() {
			args := []string{"0", "0"}
			session, err := Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 2).Should(Exit(1))
		})

		Describe("able to ping cli rpc server", func() {
			var (
				args       []string
				listener   net.Listener
				pluginPort string
				cliPort    string
			)

			BeforeEach(func() {
				pluginPort = obtainPort()
				cliPort = obtainPort()

				args = append(args, pluginPort)
				args = append(args, cliPort)

				listener, err = net.Listen("tcp", ":"+cliPort)
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				err = listener.Close()
				Expect(err).ToNot(HaveOccurred())
			})

			It("sets up a listener", func() {
				_, err := Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				var connErr error
				var conn net.Conn
				for i := 0; i < 5; i++ {
					conn, connErr = net.Dial("tcp", "127.0.0.1:"+pluginPort)
					if connErr != nil {
						time.Sleep(200 * time.Millisecond)
					} else {
						conn.Close()
						break
					}
				}
				Expect(connErr).ToNot(HaveOccurred())
			})

			Context("when called to install by `cf install-plugin`", func() {
				It("exits 1 when we cannot dial the cli rpc server", func() {
					args = append(args, "SendMetadata")

					session, err := Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())

					Eventually(session, 2).Should(Exit(1))
				})
			})
		})
	})
})
