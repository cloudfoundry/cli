// +build !windows,!386

// skipping 386 because lager uses UInt64 in Session()
// skipping windows because Unix/Linux only syscall in test.
// should refactor out the conflicts so we could test this package in multi platforms.

package sshCmd_test

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
	"time"

	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/ssh"
	"code.cloudfoundry.org/cli/cf/ssh/options"
	"code.cloudfoundry.org/cli/cf/ssh/sshfakes"
	"code.cloudfoundry.org/cli/cf/ssh/terminal"
	"code.cloudfoundry.org/cli/cf/ssh/terminal/terminalfakes"
	"code.cloudfoundry.org/diego-ssh/server"
	fake_server "code.cloudfoundry.org/diego-ssh/server/fakes"
	"code.cloudfoundry.org/diego-ssh/test_helpers"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_io"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_net"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_ssh"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/docker/docker/pkg/term"
	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSH", func() {
	var (
		fakeTerminalHelper  *terminalfakes.FakeTerminalHelper
		fakeListenerFactory *sshfakes.FakeListenerFactory

		fakeConnection    *fake_ssh.FakeConn
		fakeSecureClient  *sshfakes.FakeSecureClient
		fakeSecureDialer  *sshfakes.FakeSecureDialer
		fakeSecureSession *sshfakes.FakeSecureSession

		terminalHelper    terminal.TerminalHelper
		keepAliveDuration time.Duration
		secureShell       sshCmd.SecureShell

		stdinPipe *fake_io.FakeWriteCloser

		currentApp             models.Application
		sshEndpointFingerprint string
		sshEndpoint            string
		token                  string
	)

	BeforeEach(func() {
		fakeTerminalHelper = new(terminalfakes.FakeTerminalHelper)
		terminalHelper = terminal.DefaultHelper()

		fakeListenerFactory = new(sshfakes.FakeListenerFactory)
		fakeListenerFactory.ListenStub = net.Listen

		keepAliveDuration = 30 * time.Second

		currentApp = models.Application{}
		sshEndpoint = ""
		sshEndpointFingerprint = ""
		token = ""

		fakeConnection = new(fake_ssh.FakeConn)
		fakeSecureClient = new(sshfakes.FakeSecureClient)
		fakeSecureDialer = new(sshfakes.FakeSecureDialer)
		fakeSecureSession = new(sshfakes.FakeSecureSession)

		fakeSecureDialer.DialReturns(fakeSecureClient, nil)
		fakeSecureClient.NewSessionReturns(fakeSecureSession, nil)
		fakeSecureClient.ConnReturns(fakeConnection)

		stdinPipe = &fake_io.FakeWriteCloser{}
		stdinPipe.WriteStub = func(p []byte) (int, error) {
			return len(p), nil
		}

		stdoutPipe := &fake_io.FakeReader{}
		stdoutPipe.ReadStub = func(p []byte) (int, error) {
			return 0, io.EOF
		}

		stderrPipe := &fake_io.FakeReader{}
		stderrPipe.ReadStub = func(p []byte) (int, error) {
			return 0, io.EOF
		}

		fakeSecureSession.StdinPipeReturns(stdinPipe, nil)
		fakeSecureSession.StdoutPipeReturns(stdoutPipe, nil)
		fakeSecureSession.StderrPipeReturns(stderrPipe, nil)
	})

	JustBeforeEach(func() {
		secureShell = sshCmd.NewSecureShell(
			fakeSecureDialer,
			terminalHelper,
			fakeListenerFactory,
			keepAliveDuration,
			currentApp,
			sshEndpointFingerprint,
			sshEndpoint,
			token,
		)
	})

	Describe("Validation", func() {
		var connectErr error
		var opts *options.SSHOptions

		BeforeEach(func() {
			opts = &options.SSHOptions{
				AppName: "app-1",
			}
		})

		JustBeforeEach(func() {
			connectErr = secureShell.Connect(opts)
		})

		Context("when the app model and endpoint info are successfully acquired", func() {
			BeforeEach(func() {
				token = ""
				currentApp.State = "STARTED"
				currentApp.Diego = true
			})

			Context("when the app is not in the 'STARTED' state", func() {
				BeforeEach(func() {
					currentApp.State = "STOPPED"
					currentApp.Diego = true
				})

				It("returns an error", func() {
					Expect(connectErr).To(MatchError(MatchRegexp("Application.*not in the STARTED state")))
				})
			})

			Context("when the app is not a Diego app", func() {
				BeforeEach(func() {
					currentApp.State = "STARTED"
					currentApp.Diego = false
				})

				It("returns an error", func() {
					Expect(connectErr).To(MatchError(MatchRegexp("Application.*not running on Diego")))
				})
			})

			Context("when dialing fails", func() {
				var dialError = errors.New("woops")

				BeforeEach(func() {
					fakeSecureDialer.DialReturns(nil, dialError)
				})

				It("returns the dial error", func() {
					Expect(connectErr).To(Equal(dialError))
					Expect(fakeSecureDialer.DialCallCount()).To(Equal(1))
				})
			})
		})
	})

	Describe("InteractiveSession", func() {
		var opts *options.SSHOptions
		var sessionError error
		var interactiveSessionInvoker func(secureShell sshCmd.SecureShell)

		BeforeEach(func() {
			sshEndpoint = "ssh.example.com:22"

			opts = &options.SSHOptions{
				AppName: "app-name",
				Index:   2,
			}

			currentApp.State = "STARTED"
			currentApp.Diego = true
			currentApp.GUID = "app-guid"
			token = "bearer token"

			interactiveSessionInvoker = func(secureShell sshCmd.SecureShell) {
				sessionError = secureShell.InteractiveSession()
			}
		})

		JustBeforeEach(func() {
			connectErr := secureShell.Connect(opts)
			Expect(connectErr).NotTo(HaveOccurred())
			interactiveSessionInvoker(secureShell)
		})

		It("dials the correct endpoint as the correct user", func() {
			Expect(fakeSecureDialer.DialCallCount()).To(Equal(1))

			network, address, config := fakeSecureDialer.DialArgsForCall(0)
			Expect(network).To(Equal("tcp"))
			Expect(address).To(Equal("ssh.example.com:22"))
			Expect(config.Auth).NotTo(BeEmpty())
			Expect(config.User).To(Equal("cf:app-guid/2"))
			Expect(config.HostKeyCallback).NotTo(BeNil())
		})

		Context("when host key validation is enabled", func() {
			var callback func(hostname string, remote net.Addr, key ssh.PublicKey) error
			var addr net.Addr

			JustBeforeEach(func() {
				Expect(fakeSecureDialer.DialCallCount()).To(Equal(1))
				_, _, config := fakeSecureDialer.DialArgsForCall(0)
				callback = config.HostKeyCallback

				listener, err := net.Listen("tcp", "localhost:0")
				Expect(err).NotTo(HaveOccurred())

				addr = listener.Addr()
				listener.Close()
			})

			Context("when the SHA1 fingerprint does not match", func() {
				BeforeEach(func() {
					sshEndpointFingerprint = "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00"
				})

				It("returns an error'", func() {
					err := callback("", addr, TestHostKey.PublicKey())
					Expect(err).To(MatchError(MatchRegexp("Host key verification failed\\.")))
					Expect(err).To(MatchError(MatchRegexp("The fingerprint of the received key was \".*\"")))
				})
			})

			Context("when the MD5 fingerprint does not match", func() {
				BeforeEach(func() {
					sshEndpointFingerprint = "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00"
				})

				It("returns an error'", func() {
					err := callback("", addr, TestHostKey.PublicKey())
					Expect(err).To(MatchError(MatchRegexp("Host key verification failed\\.")))
					Expect(err).To(MatchError(MatchRegexp("The fingerprint of the received key was \".*\"")))
				})
			})

			Context("when no fingerprint is present in endpoint info", func() {
				BeforeEach(func() {
					sshEndpointFingerprint = ""
					sshEndpoint = ""
				})

				It("returns an error'", func() {
					err := callback("", addr, TestHostKey.PublicKey())
					Expect(err).To(MatchError(MatchRegexp("Unable to verify identity of host\\.")))
					Expect(err).To(MatchError(MatchRegexp("The fingerprint of the received key was \".*\"")))
				})
			})

			Context("when the fingerprint length doesn't make sense", func() {
				BeforeEach(func() {
					sshEndpointFingerprint = "garbage"
				})

				It("returns an error", func() {
					err := callback("", addr, TestHostKey.PublicKey())
					Eventually(err).Should(MatchError(MatchRegexp("Unsupported host key fingerprint format")))
				})
			})
		})

		Context("when the skip host validation flag is set", func() {
			BeforeEach(func() {
				opts.SkipHostValidation = true
			})

			It("removes the HostKeyCallback from the client config", func() {
				Expect(fakeSecureDialer.DialCallCount()).To(Equal(1))

				_, _, config := fakeSecureDialer.DialArgsForCall(0)
				Expect(config.HostKeyCallback).To(BeNil())
			})
		})

		Context("when dialing is successful", func() {
			BeforeEach(func() {
				fakeTerminalHelper.StdStreamsStub = terminalHelper.StdStreams
				terminalHelper = fakeTerminalHelper
			})

			It("creates a new secure shell session", func() {
				Expect(fakeSecureClient.NewSessionCallCount()).To(Equal(1))
			})

			It("closes the session", func() {
				Expect(fakeSecureSession.CloseCallCount()).To(Equal(1))
			})

			It("allocates standard streams", func() {
				Expect(fakeTerminalHelper.StdStreamsCallCount()).To(Equal(1))
			})

			It("gets a stdin pipe for the session", func() {
				Expect(fakeSecureSession.StdinPipeCallCount()).To(Equal(1))
			})

			Context("when getting the stdin pipe fails", func() {
				BeforeEach(func() {
					fakeSecureSession.StdinPipeReturns(nil, errors.New("woops"))
				})

				It("returns the error", func() {
					Expect(sessionError).Should(MatchError("woops"))
				})
			})

			It("gets a stdout pipe for the session", func() {
				Expect(fakeSecureSession.StdoutPipeCallCount()).To(Equal(1))
			})

			Context("when getting the stdout pipe fails", func() {
				BeforeEach(func() {
					fakeSecureSession.StdoutPipeReturns(nil, errors.New("woops"))
				})

				It("returns the error", func() {
					Expect(sessionError).Should(MatchError("woops"))
				})
			})

			It("gets a stderr pipe for the session", func() {
				Expect(fakeSecureSession.StderrPipeCallCount()).To(Equal(1))
			})

			Context("when getting the stderr pipe fails", func() {
				BeforeEach(func() {
					fakeSecureSession.StderrPipeReturns(nil, errors.New("woops"))
				})

				It("returns the error", func() {
					Expect(sessionError).Should(MatchError("woops"))
				})
			})
		})

		Context("when stdin is a terminal", func() {
			var master, slave *os.File

			BeforeEach(func() {
				_, stdout, stderr := terminalHelper.StdStreams()

				var err error
				master, slave, err = pty.Open()
				Expect(err).NotTo(HaveOccurred())

				fakeTerminalHelper.IsTerminalStub = terminalHelper.IsTerminal
				fakeTerminalHelper.GetFdInfoStub = terminalHelper.GetFdInfo
				fakeTerminalHelper.GetWinsizeStub = terminalHelper.GetWinsize
				fakeTerminalHelper.StdStreamsReturns(slave, stdout, stderr)
				terminalHelper = fakeTerminalHelper
			})

			AfterEach(func() {
				master.Close()
				// slave.Close() // race
			})

			Context("when a command is not specified", func() {
				var terminalType string

				BeforeEach(func() {
					terminalType = os.Getenv("TERM")
					os.Setenv("TERM", "test-terminal-type")

					winsize := &term.Winsize{Width: 1024, Height: 256}
					fakeTerminalHelper.GetWinsizeReturns(winsize, nil)

					fakeSecureSession.ShellStub = func() error {
						Expect(fakeTerminalHelper.SetRawTerminalCallCount()).To(Equal(1))
						Expect(fakeTerminalHelper.RestoreTerminalCallCount()).To(Equal(0))
						return nil
					}
				})

				AfterEach(func() {
					os.Setenv("TERM", terminalType)
				})

				It("requests a pty with the correct terminal type, window size, and modes", func() {
					Expect(fakeSecureSession.RequestPtyCallCount()).To(Equal(1))
					Expect(fakeTerminalHelper.GetWinsizeCallCount()).To(Equal(1))

					termType, height, width, modes := fakeSecureSession.RequestPtyArgsForCall(0)
					Expect(termType).To(Equal("test-terminal-type"))
					Expect(height).To(Equal(256))
					Expect(width).To(Equal(1024))

					expectedModes := ssh.TerminalModes{
						ssh.ECHO:          1,
						ssh.TTY_OP_ISPEED: 115200,
						ssh.TTY_OP_OSPEED: 115200,
					}
					Expect(modes).To(Equal(expectedModes))
				})

				Context("when the TERM environment variable is not set", func() {
					BeforeEach(func() {
						os.Unsetenv("TERM")
					})

					It("requests a pty with the default terminal type", func() {
						Expect(fakeSecureSession.RequestPtyCallCount()).To(Equal(1))

						termType, _, _, _ := fakeSecureSession.RequestPtyArgsForCall(0)
						Expect(termType).To(Equal("xterm"))
					})
				})

				It("puts the terminal into raw mode and restores it after running the shell", func() {
					Expect(fakeSecureSession.ShellCallCount()).To(Equal(1))
					Expect(fakeTerminalHelper.SetRawTerminalCallCount()).To(Equal(1))
					Expect(fakeTerminalHelper.RestoreTerminalCallCount()).To(Equal(1))
				})

				Context("when the pty allocation fails", func() {
					var ptyError error

					BeforeEach(func() {
						ptyError = errors.New("pty allocation error")
						fakeSecureSession.RequestPtyReturns(ptyError)
					})

					It("returns the error", func() {
						Expect(sessionError).To(Equal(ptyError))
					})
				})

				Context("when placing the terminal into raw mode fails", func() {
					BeforeEach(func() {
						fakeTerminalHelper.SetRawTerminalReturns(nil, errors.New("woops"))
					})

					It("keeps calm and carries on", func() {
						Expect(fakeSecureSession.ShellCallCount()).To(Equal(1))
					})

					It("does not not restore the terminal", func() {
						Expect(fakeSecureSession.ShellCallCount()).To(Equal(1))
						Expect(fakeTerminalHelper.SetRawTerminalCallCount()).To(Equal(1))
						Expect(fakeTerminalHelper.RestoreTerminalCallCount()).To(Equal(0))
					})
				})
			})

			Context("when a command is specified", func() {
				BeforeEach(func() {
					opts.Command = []string{"echo", "-n", "hello"}
				})

				Context("when a terminal is requested", func() {
					BeforeEach(func() {
						opts.TerminalRequest = options.RequestTTYYes
					})

					It("requests a pty", func() {
						Expect(fakeSecureSession.RequestPtyCallCount()).To(Equal(1))
					})
				})

				Context("when a terminal is not explicitly requested", func() {
					It("does not request a pty", func() {
						Expect(fakeSecureSession.RequestPtyCallCount()).To(Equal(0))
					})
				})
			})
		})

		Context("when stdin is not a terminal", func() {
			BeforeEach(func() {
				_, stdout, stderr := terminalHelper.StdStreams()

				stdin := &fake_io.FakeReadCloser{}
				stdin.ReadStub = func(p []byte) (int, error) {
					return 0, io.EOF
				}

				fakeTerminalHelper.IsTerminalStub = terminalHelper.IsTerminal
				fakeTerminalHelper.GetFdInfoStub = terminalHelper.GetFdInfo
				fakeTerminalHelper.GetWinsizeStub = terminalHelper.GetWinsize
				fakeTerminalHelper.StdStreamsReturns(stdin, stdout, stderr)
				terminalHelper = fakeTerminalHelper
			})

			Context("when a terminal is not requested", func() {
				It("does not request a pty", func() {
					Expect(fakeSecureSession.RequestPtyCallCount()).To(Equal(0))
				})
			})

			Context("when a terminal is requested", func() {
				BeforeEach(func() {
					opts.TerminalRequest = options.RequestTTYYes
				})

				It("does not request a pty", func() {
					Expect(fakeSecureSession.RequestPtyCallCount()).To(Equal(0))
				})
			})
		})

		Context("when a terminal is forced", func() {
			BeforeEach(func() {
				opts.TerminalRequest = options.RequestTTYForce
			})

			It("requests a pty", func() {
				Expect(fakeSecureSession.RequestPtyCallCount()).To(Equal(1))
			})
		})

		Context("when a terminal is disabled", func() {
			BeforeEach(func() {
				opts.TerminalRequest = options.RequestTTYNo
			})

			It("does not request a pty", func() {
				Expect(fakeSecureSession.RequestPtyCallCount()).To(Equal(0))
			})
		})

		Context("when a command is not specified", func() {
			It("requests an interactive shell", func() {
				Expect(fakeSecureSession.ShellCallCount()).To(Equal(1))
			})

			Context("when the shell request returns an error", func() {
				BeforeEach(func() {
					fakeSecureSession.ShellReturns(errors.New("oh bother"))
				})

				It("returns the error", func() {
					Expect(sessionError).To(MatchError("oh bother"))
				})
			})
		})

		Context("when a command is specifed", func() {
			BeforeEach(func() {
				opts.Command = []string{"echo", "-n", "hello"}
			})

			It("starts the command", func() {
				Expect(fakeSecureSession.StartCallCount()).To(Equal(1))
				Expect(fakeSecureSession.StartArgsForCall(0)).To(Equal("echo -n hello"))
			})

			Context("when the command fails to start", func() {
				BeforeEach(func() {
					fakeSecureSession.StartReturns(errors.New("oh well"))
				})

				It("returns the error", func() {
					Expect(sessionError).To(MatchError("oh well"))
				})
			})
		})

		Context("when the shell or command has started", func() {
			var (
				stdin                  *fake_io.FakeReadCloser
				stdout, stderr         *fake_io.FakeWriter
				stdinPipe              *fake_io.FakeWriteCloser
				stdoutPipe, stderrPipe *fake_io.FakeReader
			)

			BeforeEach(func() {
				stdin = &fake_io.FakeReadCloser{}
				stdin.ReadStub = func(p []byte) (int, error) {
					p[0] = 0
					return 1, io.EOF
				}
				stdinPipe = &fake_io.FakeWriteCloser{}
				stdinPipe.WriteStub = func(p []byte) (int, error) {
					defer GinkgoRecover()
					Expect(p[0]).To(Equal(byte(0)))
					return 1, nil
				}

				stdoutPipe = &fake_io.FakeReader{}
				stdoutPipe.ReadStub = func(p []byte) (int, error) {
					p[0] = 1
					return 1, io.EOF
				}
				stdout = &fake_io.FakeWriter{}
				stdout.WriteStub = func(p []byte) (int, error) {
					defer GinkgoRecover()
					Expect(p[0]).To(Equal(byte(1)))
					return 1, nil
				}

				stderrPipe = &fake_io.FakeReader{}
				stderrPipe.ReadStub = func(p []byte) (int, error) {
					p[0] = 2
					return 1, io.EOF
				}
				stderr = &fake_io.FakeWriter{}
				stderr.WriteStub = func(p []byte) (int, error) {
					defer GinkgoRecover()
					Expect(p[0]).To(Equal(byte(2)))
					return 1, nil
				}

				fakeTerminalHelper.StdStreamsReturns(stdin, stdout, stderr)
				terminalHelper = fakeTerminalHelper

				fakeSecureSession.StdinPipeReturns(stdinPipe, nil)
				fakeSecureSession.StdoutPipeReturns(stdoutPipe, nil)
				fakeSecureSession.StderrPipeReturns(stderrPipe, nil)

				fakeSecureSession.WaitReturns(errors.New("error result"))
			})

			It("copies data from the stdin stream to the session stdin pipe", func() {
				Eventually(stdin.ReadCallCount).Should(Equal(1))
				Eventually(stdinPipe.WriteCallCount).Should(Equal(1))
			})

			It("copies data from the session stdout pipe to the stdout stream", func() {
				Eventually(stdoutPipe.ReadCallCount).Should(Equal(1))
				Eventually(stdout.WriteCallCount).Should(Equal(1))
			})

			It("copies data from the session stderr pipe to the stderr stream", func() {
				Eventually(stderrPipe.ReadCallCount).Should(Equal(1))
				Eventually(stderr.WriteCallCount).Should(Equal(1))
			})

			It("waits for the session to end", func() {
				Expect(fakeSecureSession.WaitCallCount()).To(Equal(1))
			})

			It("returns the result from wait", func() {
				Expect(sessionError).To(MatchError("error result"))
			})

			Context("when the session terminates before stream copies complete", func() {
				var sessionErrorCh chan error

				BeforeEach(func() {
					sessionErrorCh = make(chan error, 1)

					interactiveSessionInvoker = func(secureShell sshCmd.SecureShell) {
						go func() { sessionErrorCh <- secureShell.InteractiveSession() }()
					}

					stdoutPipe.ReadStub = func(p []byte) (int, error) {
						defer GinkgoRecover()
						Eventually(fakeSecureSession.WaitCallCount).Should(Equal(1))
						Consistently(sessionErrorCh).ShouldNot(Receive())

						p[0] = 1
						return 1, io.EOF
					}

					stderrPipe.ReadStub = func(p []byte) (int, error) {
						defer GinkgoRecover()
						Eventually(fakeSecureSession.WaitCallCount).Should(Equal(1))
						Consistently(sessionErrorCh).ShouldNot(Receive())

						p[0] = 2
						return 1, io.EOF
					}
				})

				It("waits for the copies to complete", func() {
					Eventually(sessionErrorCh).Should(Receive())
					Expect(stdoutPipe.ReadCallCount()).To(Equal(1))
					Expect(stderrPipe.ReadCallCount()).To(Equal(1))
				})
			})

			Context("when stdin is closed", func() {
				BeforeEach(func() {
					stdin.ReadStub = func(p []byte) (int, error) {
						defer GinkgoRecover()
						Consistently(stdinPipe.CloseCallCount).Should(Equal(0))
						p[0] = 0
						return 1, io.EOF
					}
				})

				It("closes the stdinPipe", func() {
					Eventually(stdinPipe.CloseCallCount).Should(Equal(1))
				})
			})
		})

		Context("when stdout is a terminal and a window size change occurs", func() {
			var master, slave *os.File

			BeforeEach(func() {
				stdin, _, stderr := terminalHelper.StdStreams()

				var err error
				master, slave, err = pty.Open()
				Expect(err).NotTo(HaveOccurred())

				fakeTerminalHelper.IsTerminalStub = terminalHelper.IsTerminal
				fakeTerminalHelper.GetFdInfoStub = terminalHelper.GetFdInfo
				fakeTerminalHelper.GetWinsizeStub = terminalHelper.GetWinsize
				fakeTerminalHelper.StdStreamsReturns(stdin, slave, stderr)
				terminalHelper = fakeTerminalHelper

				winsize := &term.Winsize{Height: 100, Width: 100}
				err = term.SetWinsize(slave.Fd(), winsize)
				Expect(err).NotTo(HaveOccurred())

				fakeSecureSession.WaitStub = func() error {
					fakeSecureSession.SendRequestCallCount()
					Expect(fakeSecureSession.SendRequestCallCount()).To(Equal(0))

					// No dimension change
					for i := 0; i < 3; i++ {
						winsize := &term.Winsize{Height: 100, Width: 100}
						err = term.SetWinsize(slave.Fd(), winsize)
						Expect(err).NotTo(HaveOccurred())
					}

					winsize := &term.Winsize{Height: 100, Width: 200}
					err = term.SetWinsize(slave.Fd(), winsize)
					Expect(err).NotTo(HaveOccurred())

					err = syscall.Kill(syscall.Getpid(), syscall.SIGWINCH)
					Expect(err).NotTo(HaveOccurred())

					Eventually(fakeSecureSession.SendRequestCallCount).Should(Equal(1))
					return nil
				}
			})

			AfterEach(func() {
				master.Close()
				slave.Close()
			})

			It("sends window change events when the window dimensions change", func() {
				Expect(fakeSecureSession.SendRequestCallCount()).To(Equal(1))

				requestType, wantReply, message := fakeSecureSession.SendRequestArgsForCall(0)
				Expect(requestType).To(Equal("window-change"))
				Expect(wantReply).To(BeFalse())

				type resizeMessage struct {
					Width       uint32
					Height      uint32
					PixelWidth  uint32
					PixelHeight uint32
				}
				var resizeMsg resizeMessage

				err := ssh.Unmarshal(message, &resizeMsg)
				Expect(err).NotTo(HaveOccurred())

				Expect(resizeMsg).To(Equal(resizeMessage{Height: 100, Width: 200}))
			})
		})

		Describe("keep alive messages", func() {
			var times []time.Time
			var timesCh chan []time.Time
			var done chan struct{}

			BeforeEach(func() {
				keepAliveDuration = 100 * time.Millisecond

				times = []time.Time{}
				timesCh = make(chan []time.Time, 1)
				done = make(chan struct{}, 1)

				fakeConnection.SendRequestStub = func(reqName string, wantReply bool, message []byte) (bool, []byte, error) {
					Expect(reqName).To(Equal("keepalive@cloudfoundry.org"))
					Expect(wantReply).To(BeTrue())
					Expect(message).To(BeNil())

					times = append(times, time.Now())
					if len(times) == 3 {
						timesCh <- times
						close(done)
					}
					return true, nil, nil
				}

				fakeSecureSession.WaitStub = func() error {
					Eventually(done).Should(BeClosed())
					return nil
				}
			})

			It("sends keep alive messages at the expected interval", func() {
				times := <-timesCh
				Expect(times[2]).To(BeTemporally("~", times[0].Add(200*time.Millisecond), 100*time.Millisecond))
			})
		})
	})

	Describe("LocalPortForward", func() {
		var (
			opts              *options.SSHOptions
			localForwardError error

			echoAddress  string
			echoListener *fake_net.FakeListener
			echoHandler  *fake_server.FakeConnectionHandler
			echoServer   *server.Server

			localAddress string

			realLocalListener net.Listener
			fakeLocalListener *fake_net.FakeListener
		)

		BeforeEach(func() {
			logger := lagertest.NewTestLogger("test")

			var err error
			realLocalListener, err = net.Listen("tcp", "127.0.0.1:0")
			Expect(err).NotTo(HaveOccurred())

			localAddress = realLocalListener.Addr().String()
			fakeListenerFactory.ListenReturns(realLocalListener, nil)

			echoHandler = &fake_server.FakeConnectionHandler{}
			echoHandler.HandleConnectionStub = func(conn net.Conn) {
				io.Copy(conn, conn)
				conn.Close()
			}

			realListener, err := net.Listen("tcp", "127.0.0.1:0")
			Expect(err).NotTo(HaveOccurred())
			echoAddress = realListener.Addr().String()

			echoListener = &fake_net.FakeListener{}
			echoListener.AcceptStub = realListener.Accept
			echoListener.CloseStub = realListener.Close
			echoListener.AddrStub = realListener.Addr

			fakeLocalListener = &fake_net.FakeListener{}
			fakeLocalListener.AcceptReturns(nil, errors.New("Not Accepting Connections"))

			echoServer = server.NewServer(logger.Session("echo"), "", echoHandler)
			echoServer.SetListener(echoListener)
			go echoServer.Serve()

			opts = &options.SSHOptions{
				AppName: "app-1",
				ForwardSpecs: []options.ForwardSpec{{
					ListenAddress:  localAddress,
					ConnectAddress: echoAddress,
				}},
			}

			currentApp.State = "STARTED"
			currentApp.Diego = true

			sshEndpointFingerprint = ""
			sshEndpoint = ""

			token = ""

			fakeSecureClient.DialStub = net.Dial
		})

		JustBeforeEach(func() {
			connectErr := secureShell.Connect(opts)
			Expect(connectErr).NotTo(HaveOccurred())

			localForwardError = secureShell.LocalPortForward()
		})

		AfterEach(func() {
			err := secureShell.Close()
			Expect(err).NotTo(HaveOccurred())
			echoServer.Shutdown()

			realLocalListener.Close()
		})

		validateConnectivity := func(addr string) {
			conn, err := net.Dial("tcp", addr)
			Expect(err).NotTo(HaveOccurred())

			msg := fmt.Sprintf("Hello from %s\n", addr)
			n, err := conn.Write([]byte(msg))
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(len(msg)))

			response := make([]byte, len(msg))
			n, err = conn.Read(response)
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(len(msg)))

			err = conn.Close()
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(Equal([]byte(msg)))
		}

		It("dials the connect address when a local connection is made", func() {
			Expect(localForwardError).NotTo(HaveOccurred())

			conn, err := net.Dial("tcp", localAddress)
			Expect(err).NotTo(HaveOccurred())

			Eventually(echoListener.AcceptCallCount).Should(BeNumerically(">=", 1))
			Eventually(fakeSecureClient.DialCallCount).Should(Equal(1))

			network, addr := fakeSecureClient.DialArgsForCall(0)
			Expect(network).To(Equal("tcp"))
			Expect(addr).To(Equal(echoAddress))

			Expect(conn.Close()).NotTo(HaveOccurred())
		})

		It("copies data between the local and remote connections", func() {
			validateConnectivity(localAddress)
		})

		Context("when a local connection is already open", func() {
			var (
				conn net.Conn
				err  error
			)

			JustBeforeEach(func() {
				conn, err = net.Dial("tcp", localAddress)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				err = conn.Close()
				Expect(err).NotTo(HaveOccurred())
			})

			It("allows for new incoming connections as well", func() {
				validateConnectivity(localAddress)
			})
		})

		Context("when there are multiple port forward specs", func() {
			var realLocalListener2 net.Listener
			var localAddress2 string

			BeforeEach(func() {
				var err error
				realLocalListener2, err = net.Listen("tcp", "127.0.0.1:0")
				Expect(err).NotTo(HaveOccurred())

				localAddress2 = realLocalListener2.Addr().String()

				fakeListenerFactory.ListenStub = func(network, addr string) (net.Listener, error) {
					if addr == localAddress {
						return realLocalListener, nil
					}

					if addr == localAddress2 {
						return realLocalListener2, nil
					}

					return nil, errors.New("unexpected address")
				}

				opts = &options.SSHOptions{
					AppName: "app-1",
					ForwardSpecs: []options.ForwardSpec{{
						ListenAddress:  localAddress,
						ConnectAddress: echoAddress,
					}, {
						ListenAddress:  localAddress2,
						ConnectAddress: echoAddress,
					}},
				}
			})

			AfterEach(func() {
				realLocalListener2.Close()
			})

			It("listens to all the things", func() {
				Eventually(fakeListenerFactory.ListenCallCount).Should(Equal(2))

				network, addr := fakeListenerFactory.ListenArgsForCall(0)
				Expect(network).To(Equal("tcp"))
				Expect(addr).To(Equal(localAddress))

				network, addr = fakeListenerFactory.ListenArgsForCall(1)
				Expect(network).To(Equal("tcp"))
				Expect(addr).To(Equal(localAddress2))
			})

			It("forwards to the correct target", func() {
				validateConnectivity(localAddress)
				validateConnectivity(localAddress2)
			})

			Context("when the secure client is closed", func() {
				BeforeEach(func() {
					fakeListenerFactory.ListenReturns(fakeLocalListener, nil)
					fakeLocalListener.AcceptReturns(nil, errors.New("not accepting connections"))
				})

				It("closes the listeners ", func() {
					Eventually(fakeListenerFactory.ListenCallCount).Should(Equal(2))
					Eventually(fakeLocalListener.AcceptCallCount).Should(Equal(2))

					originalCloseCount := fakeLocalListener.CloseCallCount()
					err := secureShell.Close()
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeLocalListener.CloseCallCount()).Should(Equal(originalCloseCount + 2))
				})
			})
		})

		Context("when listen fails", func() {
			BeforeEach(func() {
				fakeListenerFactory.ListenReturns(nil, errors.New("failure is an option"))
			})

			It("returns the error", func() {
				Expect(localForwardError).To(MatchError("failure is an option"))
			})
		})

		Context("when the client it closed", func() {
			BeforeEach(func() {
				fakeListenerFactory.ListenReturns(fakeLocalListener, nil)
				fakeLocalListener.AcceptReturns(nil, errors.New("not accepting and connections"))
			})

			It("closes the listener when the client is closed", func() {
				Eventually(fakeListenerFactory.ListenCallCount).Should(Equal(1))
				Eventually(fakeLocalListener.AcceptCallCount).Should(Equal(1))

				originalCloseCount := fakeLocalListener.CloseCallCount()
				err := secureShell.Close()
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeLocalListener.CloseCallCount()).Should(Equal(originalCloseCount + 1))
			})
		})

		Context("when accept fails", func() {
			var fakeConn *fake_net.FakeConn
			BeforeEach(func() {
				fakeConn = &fake_net.FakeConn{}
				fakeConn.ReadReturns(0, io.EOF)

				fakeListenerFactory.ListenReturns(fakeLocalListener, nil)
			})

			Context("with a permanent error", func() {
				BeforeEach(func() {
					fakeLocalListener.AcceptReturns(nil, errors.New("boom"))
				})

				It("stops trying to accept connections", func() {
					Eventually(fakeLocalListener.AcceptCallCount).Should(Equal(1))
					Consistently(fakeLocalListener.AcceptCallCount).Should(Equal(1))
					Expect(fakeLocalListener.CloseCallCount()).To(Equal(1))
				})
			})

			Context("with a temporary error", func() {
				var timeCh chan time.Time

				BeforeEach(func() {
					timeCh = make(chan time.Time, 3)

					fakeLocalListener.AcceptStub = func() (net.Conn, error) {
						timeCh := timeCh
						if fakeLocalListener.AcceptCallCount() > 3 {
							close(timeCh)
							return nil, test_helpers.NewTestNetError(false, false)
						} else {
							timeCh <- time.Now()
							return nil, test_helpers.NewTestNetError(false, true)
						}
					}
				})

				It("retries connecting after a short delay", func() {
					Eventually(fakeLocalListener.AcceptCallCount).Should(Equal(3))
					Expect(timeCh).To(HaveLen(3))

					times := make([]time.Time, 0)
					for t := range timeCh {
						times = append(times, t)
					}

					Expect(times[1]).To(BeTemporally("~", times[0].Add(115*time.Millisecond), 30*time.Millisecond))
					Expect(times[2]).To(BeTemporally("~", times[1].Add(115*time.Millisecond), 30*time.Millisecond))
				})
			})
		})

		Context("when dialing the connect address fails", func() {
			var fakeTarget *fake_net.FakeConn

			BeforeEach(func() {
				fakeTarget = &fake_net.FakeConn{}
				fakeSecureClient.DialReturns(fakeTarget, errors.New("boom"))
			})

			It("does not call close on the target connection", func() {
				Consistently(fakeTarget.CloseCallCount).Should(Equal(0))
			})
		})
	})

	Describe("Wait", func() {
		var opts *options.SSHOptions
		var waitErr error

		BeforeEach(func() {
			opts = &options.SSHOptions{
				AppName: "app-1",
			}

			currentApp.State = "STARTED"
			currentApp.Diego = true

			sshEndpointFingerprint = ""
			sshEndpoint = ""

			token = ""
		})

		JustBeforeEach(func() {
			connectErr := secureShell.Connect(opts)
			Expect(connectErr).NotTo(HaveOccurred())

			waitErr = secureShell.Wait()
		})

		It("calls wait on the secureClient", func() {
			Expect(waitErr).NotTo(HaveOccurred())
			Expect(fakeSecureClient.WaitCallCount()).To(Equal(1))
		})

		Describe("keep alive messages", func() {
			var times []time.Time
			var timesCh chan []time.Time
			var done chan struct{}

			BeforeEach(func() {
				keepAliveDuration = 100 * time.Millisecond

				times = []time.Time{}
				timesCh = make(chan []time.Time, 1)
				done = make(chan struct{}, 1)

				fakeConnection.SendRequestStub = func(reqName string, wantReply bool, message []byte) (bool, []byte, error) {
					Expect(reqName).To(Equal("keepalive@cloudfoundry.org"))
					Expect(wantReply).To(BeTrue())
					Expect(message).To(BeNil())

					times = append(times, time.Now())
					if len(times) == 3 {
						timesCh <- times
						close(done)
					}
					return true, nil, nil
				}

				fakeSecureClient.WaitStub = func() error {
					Eventually(done).Should(BeClosed())
					return nil
				}
			})

			It("sends keep alive messages at the expected interval", func() {
				Expect(waitErr).NotTo(HaveOccurred())
				times := <-timesCh
				Expect(times[2]).To(BeTemporally("~", times[0].Add(200*time.Millisecond), 100*time.Millisecond))
			})
		})
	})

	Describe("Close", func() {
		var opts *options.SSHOptions

		BeforeEach(func() {
			opts = &options.SSHOptions{
				AppName: "app-1",
			}

			currentApp.State = "STARTED"
			currentApp.Diego = true

			sshEndpointFingerprint = ""
			sshEndpoint = ""

			token = ""
		})

		JustBeforeEach(func() {
			connectErr := secureShell.Connect(opts)
			Expect(connectErr).NotTo(HaveOccurred())
		})

		It("calls close on the secureClient", func() {
			err := secureShell.Close()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeSecureClient.CloseCallCount()).To(Equal(1))
		})
	})
})
