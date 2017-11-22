// +build !windows

package main_test

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/diego-ssh/cmd/sshd/testrunner"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
	"golang.org/x/crypto/ssh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("SSH daemon", func() {
	var (
		runner  ifrit.Runner
		process ifrit.Process

		address       string
		hostKey       string
		privateKey    string
		authorizedKey string

		allowedCiphers      string
		allowedMACs         string
		allowedKeyExchanges string

		allowUnauthenticatedClients bool
		inheritDaemonEnv            bool
	)

	BeforeEach(func() {
		hostKey = hostKeyPem
		privateKey = privateKeyPem
		authorizedKey = publicAuthorizedKey

		allowedCiphers = ""
		allowedMACs = ""
		allowedKeyExchanges = ""

		allowUnauthenticatedClients = false
		inheritDaemonEnv = false
		address = fmt.Sprintf("127.0.0.1:%d", sshdPort)
	})

	JustBeforeEach(func() {
		args := testrunner.Args{
			Address:       address,
			HostKey:       string(hostKey),
			AuthorizedKey: string(authorizedKey),

			AllowedCiphers:      string(allowedCiphers),
			AllowedMACs:         string(allowedMACs),
			AllowedKeyExchanges: string(allowedKeyExchanges),

			AllowUnauthenticatedClients: allowUnauthenticatedClients,
			InheritDaemonEnv:            inheritDaemonEnv,
		}

		runner = testrunner.New(sshdPath, args)
		process = ifrit.Invoke(runner)
	})

	AfterEach(func() {
		ginkgomon.Kill(process, 3*time.Second)
	})

	Describe("argument validation", func() {
		Context("when an ill-formed host key is provided", func() {
			BeforeEach(func() {
				hostKey = "host-key"
			})

			It("reports and dies", func() {
				Expect(runner).To(gbytes.Say("failed-to-parse-host-key"))
				Expect(runner).NotTo(gexec.Exit(0))
			})
		})

		Context("when an ill-formed authorized key is provided", func() {
			BeforeEach(func() {
				authorizedKey = "authorized-key"
			})

			It("reports and dies", func() {
				Expect(runner).To(gbytes.Say(`configure-failed.*ssh: no key found`))
				Expect(runner).NotTo(gexec.Exit(0))
			})
		})

		Context("the authorized key is not provided", func() {
			BeforeEach(func() {
				authorizedKey = ""
			})

			Context("and allowUnauthenticatedClients is not true", func() {
				BeforeEach(func() {
					allowUnauthenticatedClients = false
				})

				It("reports and dies", func() {
					Expect(runner).To(gbytes.Say("authorized-key-required"))
					Expect(runner).NotTo(gexec.Exit(0))
				})
			})

			Context("and allowUnauthenticatedClients is true", func() {
				BeforeEach(func() {
					allowUnauthenticatedClients = true
				})

				It("starts normally", func() {
					Expect(process).NotTo(BeNil())
				})
			})
		})
	})

	Describe("env variable validation", func() {
		Context("when an ill-formed host key is provided", func() {
			BeforeEach(func() {
				hostKey = "invalid-host-key"
			})

			It("reports and dies", func() {
				Expect(runner).To(gbytes.Say("failed-to-parse-host-key"))
				Expect(runner).NotTo(gexec.Exit(0))
			})
		})

		Context("when an ill-formed authorized key is provided", func() {
			BeforeEach(func() {
				authorizedKey = "invalid-authorized-key"
			})

			It("reports and dies", func() {
				Expect(runner).To(gbytes.Say(`configure-failed.*ssh: no key found`))
				Expect(runner).NotTo(gexec.Exit(0))
			})
		})

		Context("the authorized key is not provided", func() {
			BeforeEach(func() {
				authorizedKey = ""
			})

			Context("and allowUnauthenticatedClients is not true", func() {
				BeforeEach(func() {
					allowUnauthenticatedClients = false
				})

				It("reports and dies", func() {
					Expect(runner).To(gbytes.Say("authorized-key-required"))
					Expect(runner).NotTo(gexec.Exit(0))
				})
			})

			Context("and allowUnauthenticatedClients is true", func() {
				BeforeEach(func() {
					allowUnauthenticatedClients = true
				})

				It("starts normally", func() {
					Expect(process).NotTo(BeNil())
				})
			})
		})

		Context("when the hostKey is provided as an env variable", func() {
			var (
				client           *ssh.Client
				clientConfig     *ssh.ClientConfig
				handshakeHostKey ssh.PublicKey
			)

			JustBeforeEach(func() {
				Expect(process).NotTo(BeNil())
				client, _ = ssh.Dial("tcp", address, clientConfig)
			})

			AfterEach(func() {
				if client != nil {
					client.Close()
				}
				os.Unsetenv("SSHD_HOSTKEY")
			})

			BeforeEach(func() {
				hostKey = "host-key"
				os.Setenv("SSHD_HOSTKEY", hostKeyPem)
				allowUnauthenticatedClients = true
				clientConfig = &ssh.ClientConfig{
					HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
						handshakeHostKey = key
						return nil
					},
				}
			})

			It("uses the hostKey from the environment", func() {
				sshHostKey, err := ssh.ParsePrivateKey([]byte(hostKeyPem))
				Expect(err).NotTo(HaveOccurred())

				sshPublicHostKey := sshHostKey.PublicKey()
				Expect(sshPublicHostKey.Marshal()).To(Equal(handshakeHostKey.Marshal()))
			})
		})
	})

	Describe("daemon execution", func() {
		var (
			client       *ssh.Client
			dialErr      error
			clientConfig *ssh.ClientConfig
		)

		JustBeforeEach(func() {
			Expect(process).NotTo(BeNil())
			client, dialErr = ssh.Dial("tcp", address, clientConfig)
		})

		AfterEach(func() {
			if client != nil {
				client.Close()
			}
		})

		var ItDoesNotExposeSensitiveInformation = func() {
			It("does not expose the key on the command line", func() {
				pid := runner.(*ginkgomon.Runner).Command.Process.Pid
				command := exec.Command("ps", "-fp", strconv.Itoa(pid))
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				keyRegex := regexp.QuoteMeta(authorizedKey[:len(authorizedKey)-1])
				Expect(session.Out).NotTo(gbytes.Say(keyRegex))
			})
		}

		Context("when a host key is not specified", func() {
			BeforeEach(func() {
				hostKey = ""
				allowUnauthenticatedClients = true
				clientConfig = &ssh.ClientConfig{}
			})

			It("generates one internally", func() {
				Expect(process).NotTo(BeNil())

				Expect(client).NotTo(BeNil())
				Expect(dialErr).NotTo(HaveOccurred())
			})

			ItDoesNotExposeSensitiveInformation()
		})

		Context("when a host key is specified", func() {
			var handshakeHostKey ssh.PublicKey

			BeforeEach(func() {
				allowUnauthenticatedClients = true
				clientConfig = &ssh.ClientConfig{
					HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
						handshakeHostKey = key
						return nil
					},
				}
			})

			It("uses the host key provided on the command line", func() {
				sshHostKey, err := ssh.ParsePrivateKey([]byte(hostKeyPem))
				Expect(err).NotTo(HaveOccurred())

				sshPublicHostKey := sshHostKey.PublicKey()
				Expect(sshPublicHostKey.Marshal()).To(Equal(handshakeHostKey.Marshal()))
			})

			ItDoesNotExposeSensitiveInformation()
		})

		Context("when unauthenticated clients are not allowed", func() {
			BeforeEach(func() {
				clientConfig = &ssh.ClientConfig{}
			})

			It("starts the daemon", func() {
				Expect(process).NotTo(BeNil())
			})

			It("rejects the client handshake", func() {
				Expect(dialErr).To(MatchError(ContainSubstring("ssh: handshake failed")))
			})

			Context("and client has a valid private key", func() {
				BeforeEach(func() {
					key, err := ssh.ParsePrivateKey([]byte(privateKey))
					Expect(err).NotTo(HaveOccurred())

					clientConfig = &ssh.ClientConfig{
						User: os.Getenv("USER"),
						Auth: []ssh.AuthMethod{
							ssh.PublicKeys(key),
						},
					}
				})

				It("can complete a handshake with the daemon", func() {
					Expect(dialErr).NotTo(HaveOccurred())
					Expect(client).NotTo(BeNil())
				})
			})
		})

		Context("when the daemon allows unauthenticated clients", func() {
			BeforeEach(func() {
				allowUnauthenticatedClients = true
				clientConfig = &ssh.ClientConfig{}
			})

			It("starts the daemon", func() {
				Expect(process).NotTo(BeNil())
			})

			It("allows a client without credentials to complete a handshake", func() {
				Expect(dialErr).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})

		})

		Context("when the daemon provides an unsupported cipher algorithm", func() {
			BeforeEach(func() {
				allowedCiphers = "unsupported"
				clientConfig = &ssh.ClientConfig{}
			})

			It("starts the daemon", func() {
				Expect(process).NotTo(BeNil())
			})

			It("rejects the cipher algorithm", func() {
				Expect(dialErr).To(MatchError(ContainSubstring("ssh: no common algorithm for client to server cipher")))
				Expect(client).To(BeNil())
			})
		})

		Context("when the daemon provides a supported cipher algorithm", func() {
			BeforeEach(func() {
				allowUnauthenticatedClients = true
				allowedCiphers = "aes128-ctr,aes256-ctr"
				clientConfig = &ssh.ClientConfig{}
			})

			It("starts the daemon", func() {
				Expect(process).NotTo(BeNil())
			})

			It("allows a client to complete a handshake", func() {
				Expect(dialErr).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})
		})

		Context("when the daemon provides an unsupported MAC algorithm", func() {
			BeforeEach(func() {
				allowedMACs = "unsupported"
				clientConfig = &ssh.ClientConfig{}
			})

			It("starts the daemon", func() {
				Expect(process).NotTo(BeNil())
			})

			It("rejects the MAC algorithm", func() {
				Expect(dialErr).To(MatchError(ContainSubstring("ssh: no common algorithm for client to server MAC")))
				Expect(client).To(BeNil())
			})
		})

		Context("when the daemon provides a supported MAC algorithm", func() {
			BeforeEach(func() {
				allowUnauthenticatedClients = true
				allowedMACs = "hmac-sha2-256,hmac-sha1"
				clientConfig = &ssh.ClientConfig{}
			})

			It("starts the daemon", func() {
				Expect(process).NotTo(BeNil())
			})

			It("allows a client to complete a handshake", func() {
				Expect(dialErr).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})
		})

		Context("when the daemon provides an unsupported key exchange algorithm", func() {
			BeforeEach(func() {
				allowedKeyExchanges = "unsupported"
				clientConfig = &ssh.ClientConfig{}
			})

			It("starts the daemon", func() {
				Expect(process).NotTo(BeNil())
			})

			It("rejects the key exchange algorithm", func() {
				Expect(dialErr).To(MatchError(ContainSubstring("ssh: no common algorithm for key exchange")))
				Expect(client).To(BeNil())
			})
		})

		Context("when the daemon provides a supported key exchange algorithm", func() {
			BeforeEach(func() {
				allowUnauthenticatedClients = true
				allowedKeyExchanges = "curve25519-sha256@libssh.org,ecdh-sha2-nistp384,diffie-hellman-group14-sha1"
				clientConfig = &ssh.ClientConfig{}
			})

			It("starts the daemon", func() {
				Expect(process).NotTo(BeNil())
			})

			It("allows a client to complete a handshake", func() {
				Expect(dialErr).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})
		})
	})

	Describe("SSH features", func() {
		var clientConfig *ssh.ClientConfig
		var client *ssh.Client

		BeforeEach(func() {
			allowUnauthenticatedClients = true
			clientConfig = &ssh.ClientConfig{}
		})

		JustBeforeEach(func() {
			Expect(process).NotTo(BeNil())

			var dialErr error
			client, dialErr = ssh.Dial("tcp", address, clientConfig)
			Expect(dialErr).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			client.Close()
		})

		Context("when a client requests the execution of a command", func() {
			It("runs the command", func() {
				session, err := client.NewSession()
				Expect(err).NotTo(HaveOccurred())

				result, err := session.Output("/bin/echo -n 'Hello there!'")
				Expect(err).NotTo(HaveOccurred())

				Expect(string(result)).To(Equal("Hello there!"))
			})
		})

		Context("when a client requests a shell", func() {
			Context("when inherit daemon env is enabled", func() {
				BeforeEach(func() {
					inheritDaemonEnv = true
					os.Setenv("TEST", "FOO")
					os.Setenv("PATH", os.Getenv("PATH")+":/tmp")
				})

				It("creates a shell environment", func() {
					session, err := client.NewSession()
					Expect(err).NotTo(HaveOccurred())

					stdout := &bytes.Buffer{}

					session.Stdin = strings.NewReader("/bin/echo -n $ENV_VAR")
					session.Stdout = stdout

					session.Setenv("ENV_VAR", "env_var_value")
					err = session.Shell()
					Expect(err).NotTo(HaveOccurred())

					err = session.Wait()
					Expect(err).NotTo(HaveOccurred())

					Expect(stdout.String()).To(ContainSubstring("env_var_value"))
				})

				It("inherits daemon's environment excluding PATH", func() {
					session, err := client.NewSession()
					Expect(err).NotTo(HaveOccurred())

					stdout := &bytes.Buffer{}

					session.Stdin = strings.NewReader("/bin/echo -n $TEST")
					session.Stdout = stdout

					err = session.Shell()
					Expect(err).NotTo(HaveOccurred())

					err = session.Wait()
					Expect(err).NotTo(HaveOccurred())

					Expect(stdout.String()).To(ContainSubstring("FOO"))
				})

				It("does not inherit the daemon's PATH", func() {
					session, err := client.NewSession()
					Expect(err).NotTo(HaveOccurred())

					stdout := &bytes.Buffer{}

					session.Stdin = strings.NewReader("/bin/echo -n $PATH")
					session.Stdout = stdout

					err = session.Shell()
					Expect(err).NotTo(HaveOccurred())

					err = session.Wait()
					Expect(err).NotTo(HaveOccurred())
					Expect(stdout.String()).NotTo(ContainSubstring("/tmp"))
				})
			})

			Context("when inherit daemon env is disabled", func() {
				BeforeEach(func() {
					inheritDaemonEnv = false
					os.Setenv("TEST", "FOO")
				})

				It("creates a shell environment", func() {
					session, err := client.NewSession()
					Expect(err).NotTo(HaveOccurred())

					stdout := &bytes.Buffer{}

					session.Stdin = strings.NewReader("/bin/echo -n $ENV_VAR")
					session.Stdout = stdout

					session.Setenv("ENV_VAR", "env_var_value")
					err = session.Shell()
					Expect(err).NotTo(HaveOccurred())

					err = session.Wait()
					Expect(err).NotTo(HaveOccurred())

					Expect(stdout.String()).To(ContainSubstring("env_var_value"))
				})

				It("does not inherits daemon's environment", func() {
					session, err := client.NewSession()
					Expect(err).NotTo(HaveOccurred())

					stdout := &bytes.Buffer{}

					session.Stdin = strings.NewReader("/bin/echo -n $TEST")
					session.Stdout = stdout

					err = session.Shell()
					Expect(err).NotTo(HaveOccurred())

					err = session.Wait()
					Expect(err).NotTo(HaveOccurred())

					Expect(stdout.String()).NotTo(ContainSubstring("FOO"))
				})
			})
		})

		Context("when a client requests a local port forward", func() {
			var server *ghttp.Server
			BeforeEach(func() {
				server = ghttp.NewServer()
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/"),
						ghttp.RespondWith(http.StatusOK, "hi from jim\n"),
					),
				)
			})

			It("forwards the local port to the target from the server side", func() {
				lconn, err := client.Dial("tcp", server.Addr())
				Expect(err).NotTo(HaveOccurred())

				transport := &http.Transport{
					Dial: func(network, addr string) (net.Conn, error) {
						return lconn, nil
					},
				}
				client := &http.Client{Transport: transport}

				resp, err := client.Get("http://127.0.0.1/")
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				reader := bufio.NewReader(resp.Body)
				line, err := reader.ReadString('\n')
				Expect(err).NotTo(HaveOccurred())
				Expect(line).To(ContainSubstring("hi from jim"))
			})
		})
	})
})
