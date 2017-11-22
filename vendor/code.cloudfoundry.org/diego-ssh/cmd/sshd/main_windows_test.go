// +build windows

package main_test

import (
	"fmt"
	"os"
	"time"

	"code.cloudfoundry.org/diego-ssh/cmd/sshd/testrunner"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
	"golang.org/x/crypto/ssh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func startSshd(address string) ifrit.Process {
	args := testrunner.Args{
		Address:       address,
		HostKey:       string(privateKeyPem),
		AuthorizedKey: string(publicAuthorizedKey),

		AllowUnauthenticatedClients: true,
		InheritDaemonEnv:            false,
	}

	runner := testrunner.New(sshdPath, args)
	runner.Command.Env = append(
		os.Environ(),
		fmt.Sprintf(`CF_INSTANCE_PORTS=[{"external":%d,"internal":%d}]`, sshdPort, 2222),
	)
	process := ifrit.Invoke(runner)
	return process
}

var _ = Describe("SSH daemon", func() {
	It("maps the internal port to the external port", func() {
		process := startSshd("127.0.0.1:2222")
		defer ginkgomon.Kill(process, 3*time.Second)

		clientConfig := &ssh.ClientConfig{}
		Expect(process).NotTo(BeNil())

		client, err := ssh.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", sshdPort), clientConfig)
		Expect(err).NotTo(HaveOccurred())
		client.Close()
	})

	Describe("SSH features", func() {
		var (
			process      ifrit.Process
			address      string
			clientConfig *ssh.ClientConfig
			client       *ssh.Client
		)

		BeforeEach(func() {
			address = fmt.Sprintf("127.0.0.1:%d", sshdPort)
			process = startSshd(address)
			clientConfig = &ssh.ClientConfig{}
			Expect(process).NotTo(BeNil())

			var dialErr error
			client, dialErr = ssh.Dial("tcp", address, clientConfig)
			Expect(dialErr).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			ginkgomon.Kill(process, 3*time.Second)
			client.Close()
		})

		Context("when a client requests the execution of a command", func() {
			It("runs the command", func() {
				_, err := client.NewSession()
				Expect(err).To(MatchError(ContainSubstring("not supported")))
			})
		})

		Context("when a client requests a local port forward", func() {
			var server *ghttp.Server
			BeforeEach(func() {
				server = ghttp.NewServer()
			})

			It("forwards the local port to the target from the server side", func() {
				_, err := client.Dial("tcp", server.Addr())
				Expect(err).To(MatchError(ContainSubstring("unknown channel type")))
			})

			It("server should not receive any connections", func() {
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})
		})
	})
})
