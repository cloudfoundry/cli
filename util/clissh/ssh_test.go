// +build !windows,!386

// skipping 386 because lager uses UInt64 in Session()
// skipping windows because Unix/Linux only syscall in test.
// should refactor out the conflicts so we could test this package in multi platforms.

package clissh_test

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"syscall"
	"time"

	. "code.cloudfoundry.org/cli/util/clissh"
	"code.cloudfoundry.org/cli/util/clissh/clisshfakes"
	"code.cloudfoundry.org/cli/util/clissh/ssherror"
	"code.cloudfoundry.org/diego-ssh/server"
	fake_server "code.cloudfoundry.org/diego-ssh/server/fakes"
	"code.cloudfoundry.org/diego-ssh/test_helpers"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_io"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_net"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_ssh"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/kr/pty"
	"github.com/moby/moby/pkg/term"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
)

func BlockAcceptOnClose(fake *fake_net.FakeListener) {
	waitUntilClosed := make(chan bool)
	fake.AcceptStub = func() (net.Conn, error) {
		<-waitUntilClosed
		return nil, errors.New("banana")
	}

	var once sync.Once
	fake.CloseStub = func() error {
		once.Do(func() {
			close(waitUntilClosed)
		})
		return nil
	}
}

var _ = Describe("CLI SSH", func() {
	var (
		fakeSecureDialer    *clisshfakes.FakeSecureDialer
		fakeSecureClient    *clisshfakes.FakeSecureClient
		fakeTerminalHelper  *clisshfakes.FakeTerminalHelper
		fakeListenerFactory *clisshfakes.FakeListenerFactory
		fakeSecureSession   *clisshfakes.FakeSecureSession

		fakeConnection *fake_ssh.FakeConn
		stdinPipe      *fake_io.FakeWriteCloser
		stdoutPipe     *fake_io.FakeReader
		stderrPipe     *fake_io.FakeReader
		secureShell    *SecureShell

		username               string
		passcode               string
		sshEndpoint            string
		sshEndpointFingerprint string
		skipHostValidation     bool
		commands               []string
		terminalRequest        TTYRequest
		keepAliveDuration      time.Duration
	)

	BeforeEach(func() {
		fakeSecureDialer = new(clisshfakes.FakeSecureDialer)
		fakeSecureClient = new(clisshfakes.FakeSecureClient)
		fakeTerminalHelper = new(clisshfakes.FakeTerminalHelper)
		fakeListenerFactory = new(clisshfakes.FakeListenerFactory)
		fakeSecureSession = new(clisshfakes.FakeSecureSession)

		fakeConnection = new(fake_ssh.FakeConn)
		stdinPipe = new(fake_io.FakeWriteCloser)
		stdoutPipe = new(fake_io.FakeReader)
		stderrPipe = new(fake_io.FakeReader)

		fakeListenerFactory.ListenStub = net.Listen
		fakeSecureClient.NewSessionReturns(fakeSecureSession, nil)
		fakeSecureClient.ConnReturns(fakeConnection)
		fakeSecureDialer.DialReturns(fakeSecureClient, nil)

		stdinPipe.WriteStub = func(p []byte) (int, error) {
			return len(p), nil
		}
		fakeSecureSession.StdinPipeReturns(stdinPipe, nil)

		stdoutPipe.ReadStub = func(p []byte) (int, error) {
			return 0, io.EOF
		}
		fakeSecureSession.StdoutPipeReturns(stdoutPipe, nil)

		stderrPipe.ReadStub = func(p []byte) (int, error) {
			return 0, io.EOF
		}
		fakeSecureSession.StderrPipeReturns(stderrPipe, nil)

		username = "some-user"
		passcode = "some-passcode"
		sshEndpoint = "some-endpoint"
		sshEndpointFingerprint = "some-fingerprint"
		skipHostValidation = false
		commands = []string{}
		terminalRequest = RequestTTYAuto
		keepAliveDuration = DefaultKeepAliveInterval
	})

	JustBeforeEach(func() {
		secureShell = NewSecureShell(
			fakeSecureDialer,
			fakeTerminalHelper,
			fakeListenerFactory,
			keepAliveDuration,
		)
	})

	Describe("Connect", func() {
		var connectErr error

		JustBeforeEach(func() {
			connectErr = secureShell.Connect(username, passcode, sshEndpoint, sshEndpointFingerprint, skipHostValidation)
		})

		When("dialing succeeds", func() {
			It("creates the ssh client", func() {
				Expect(connectErr).ToNot(HaveOccurred())

				Expect(fakeSecureDialer.DialCallCount()).To(Equal(1))
				protocolArg, sshEndpointArg, sshConfigArg := fakeSecureDialer.DialArgsForCall(0)
				Expect(protocolArg).To(Equal("tcp"))
				Expect(sshEndpointArg).To(Equal(sshEndpoint))
				Expect(sshConfigArg.User).To(Equal(username))
				Expect(sshConfigArg.Auth).To(HaveLen(1))
				Expect(sshConfigArg.HostKeyCallback).ToNot(BeNil())
			})
		})

		When("dialing fails", func() {
			var dialError error

			When("the error is a generic Dial error", func() {
				BeforeEach(func() {
					dialError = errors.New("woops")
					fakeSecureDialer.DialReturns(nil, dialError)
				})

				It("returns the dial error", func() {
					Expect(connectErr).To(Equal(dialError))
					Expect(fakeSecureDialer.DialCallCount()).To(Equal(1))
				})
			})

			When("the dialing error is a golang 'unable to authenticate' error", func() {
				BeforeEach(func() {
					dialError = fmt.Errorf("ssh: unable to authenticate, attempted methods %v, no supported methods remain", []string{"none", "password"})
					fakeSecureDialer.DialReturns(nil, dialError)
				})

				It("returns an UnableToAuthenticateError", func() {
					Expect(connectErr).To(MatchError(ssherror.UnableToAuthenticateError{Err: dialError}))
					Expect(fakeSecureDialer.DialCallCount()).To(Equal(1))
				})
			})
		})
	})

	Describe("InteractiveSession", func() {
		var (
			stdin          *fake_io.FakeReadCloser
			stdout, stderr *fake_io.FakeWriter

			sessionErr                error
			interactiveSessionInvoker func(secureShell *SecureShell)
		)

		BeforeEach(func() {
			stdin = new(fake_io.FakeReadCloser)
			stdout = new(fake_io.FakeWriter)
			stderr = new(fake_io.FakeWriter)

			fakeTerminalHelper.StdStreamsReturns(stdin, stdout, stderr)
			interactiveSessionInvoker = func(secureShell *SecureShell) {
				sessionErr = secureShell.InteractiveSession(commands, terminalRequest)
			}
		})

		JustBeforeEach(func() {
			connectErr := secureShell.Connect(username, passcode, sshEndpoint, sshEndpointFingerprint, skipHostValidation)
			Expect(connectErr).NotTo(HaveOccurred())
			interactiveSessionInvoker(secureShell)
		})

		When("host key validation is enabled", func() {
			var (
				callback func(hostname string, remote net.Addr, key ssh.PublicKey) error
				addr     net.Addr
			)

			BeforeEach(func() {
				skipHostValidation = false
			})

			JustBeforeEach(func() {
				Expect(fakeSecureDialer.DialCallCount()).To(Equal(1))
				_, _, config := fakeSecureDialer.DialArgsForCall(0)
				callback = config.HostKeyCallback

				listener, err := net.Listen("tcp", "localhost:0")
				Expect(err).NotTo(HaveOccurred())

				addr = listener.Addr()
				listener.Close()
			})

			When("the md5 fingerprint matches", func() {
				BeforeEach(func() {
					sshEndpointFingerprint = "41:ce:56:e6:9c:42:a9:c6:9e:68:ac:e3:4d:f6:38:79"
				})

				It("does not return an error", func() {
					Expect(callback("", addr, TestHostKey.PublicKey())).ToNot(HaveOccurred())
				})
			})

			When("the hex sha1 fingerprint matches", func() {
				BeforeEach(func() {
					sshEndpointFingerprint = "a8:e2:67:cb:ea:2a:6e:23:a1:72:ce:8f:07:92:15:ee:1f:82:f8:ca"
				})

				It("does not return an error", func() {
					Expect(callback("", addr, TestHostKey.PublicKey())).ToNot(HaveOccurred())
				})
			})

			When("the base64 sha256 fingerprint matches", func() {
				BeforeEach(func() {
					sshEndpointFingerprint = "sp/jrLuj66r+yrLDUKZdJU5tdzt4mq/UaSiNBjpgr+8"
				})

				It("does not return an error", func() {
					Expect(callback("", addr, TestHostKey.PublicKey())).ToNot(HaveOccurred())
				})
			})

			When("the base64 SHA256 fingerprint does not match", func() {
				BeforeEach(func() {
					sshEndpointFingerprint = "0000000000000000000000000000000000000000000"
				})

				It("returns an error'", func() {
					err := callback("", addr, TestHostKey.PublicKey())
					Expect(err).To(MatchError(MatchRegexp(`Host key verification failed\.`)))
					Expect(err).To(MatchError(MatchRegexp("The fingerprint of the received key was \".*\"")))
				})
			})

			When("the hex SHA1 fingerprint does not match", func() {
				BeforeEach(func() {
					sshEndpointFingerprint = "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00"
				})

				It("returns an error'", func() {
					err := callback("", addr, TestHostKey.PublicKey())
					Expect(err).To(MatchError(MatchRegexp(`Host key verification failed\.`)))
					Expect(err).To(MatchError(MatchRegexp("The fingerprint of the received key was \".*\"")))
				})
			})

			When("the MD5 fingerprint does not match", func() {
				BeforeEach(func() {
					sshEndpointFingerprint = "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00"
				})

				It("returns an error'", func() {
					err := callback("", addr, TestHostKey.PublicKey())
					Expect(err).To(MatchError(MatchRegexp(`Host key verification failed\.`)))
					Expect(err).To(MatchError(MatchRegexp("The fingerprint of the received key was \".*\"")))
				})
			})

			When("no fingerprint is present in endpoint info", func() {
				BeforeEach(func() {
					sshEndpointFingerprint = ""
					sshEndpoint = ""
				})

				It("returns an error'", func() {
					err := callback("", addr, TestHostKey.PublicKey())
					Expect(err).To(MatchError(MatchRegexp(`Unable to verify identity of host\.`)))
					Expect(err).To(MatchError(MatchRegexp("The fingerprint of the received key was \".*\"")))
				})
			})

			When("the fingerprint length doesn't make sense", func() {
				BeforeEach(func() {
					sshEndpointFingerprint = "garbage"
				})

				It("returns an error", func() {
					err := callback("", addr, TestHostKey.PublicKey())
					Eventually(err).Should(MatchError(MatchRegexp("Unsupported host key fingerprint format")))
				})
			})
		})

		When("the skip host validation flag is set", func() {
			BeforeEach(func() {
				skipHostValidation = true
			})

			It("the HostKeyCallback on the Config to always return nil", func() {
				Expect(fakeSecureDialer.DialCallCount()).To(Equal(1))

				_, _, config := fakeSecureDialer.DialArgsForCall(0)
				Expect(config.HostKeyCallback("some-hostname", nil, nil)).To(BeNil())
			})
		})

		// TODO: see if it's possible to test the piping between the ss client input and outputs and the UI object we pass in
		When("dialing is successful", func() {
			It("creates a new secure shell session", func() {
				Expect(fakeSecureClient.NewSessionCallCount()).To(Equal(1))
			})

			It("closes the session", func() {
				Expect(fakeSecureSession.CloseCallCount()).To(Equal(1))
			})

			It("gets a stdin pipe for the session", func() {
				Expect(fakeSecureSession.StdinPipeCallCount()).To(Equal(1))
			})

			When("getting the stdin pipe fails", func() {
				BeforeEach(func() {
					fakeSecureSession.StdinPipeReturns(nil, errors.New("woops"))
				})

				It("returns the error", func() {
					Expect(sessionErr).Should(MatchError("woops"))
				})
			})

			It("gets a stdout pipe for the session", func() {
				Expect(fakeSecureSession.StdoutPipeCallCount()).To(Equal(1))
			})

			When("getting the stdout pipe fails", func() {
				BeforeEach(func() {
					fakeSecureSession.StdoutPipeReturns(nil, errors.New("woops"))
				})

				It("returns the error", func() {
					Expect(sessionErr).Should(MatchError("woops"))
				})
			})

			It("gets a stderr pipe for the session", func() {
				Expect(fakeSecureSession.StderrPipeCallCount()).To(Equal(1))
			})

			When("getting the stderr pipe fails", func() {
				BeforeEach(func() {
					fakeSecureSession.StderrPipeReturns(nil, errors.New("woops"))
				})

				It("returns the error", func() {
					Expect(sessionErr).Should(MatchError("woops"))
				})
			})
		})

		When("stdin is a terminal", func() {
			var master *os.File

			BeforeEach(func() {
				var err error
				master, _, err = pty.Open()
				Expect(err).NotTo(HaveOccurred())

				terminalRequest = RequestTTYForce

				terminalHelper := DefaultTerminalHelper()
				fakeTerminalHelper.GetFdInfoStub = terminalHelper.GetFdInfo
				fakeTerminalHelper.GetWinsizeStub = terminalHelper.GetWinsize
			})

			AfterEach(func() {
				master.Close()
			})

			When("a command is not specified", func() {
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

				When("the TERM environment variable is not set", func() {
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

				When("the pty allocation fails", func() {
					var ptyError error

					BeforeEach(func() {
						ptyError = errors.New("pty allocation error")
						fakeSecureSession.RequestPtyReturns(ptyError)
					})

					It("returns the error", func() {
						Expect(sessionErr).To(Equal(ptyError))
					})
				})

				When("placing the terminal into raw mode fails", func() {
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

			When("a command is specified", func() {
				BeforeEach(func() {
					commands = []string{"echo", "-n", "hello"}
				})

				When("a terminal is requested", func() {
					BeforeEach(func() {
						terminalRequest = RequestTTYYes
						fakeTerminalHelper.GetFdInfoReturns(0, true)
					})

					It("requests a pty", func() {
						Expect(fakeSecureSession.RequestPtyCallCount()).To(Equal(1))
					})
				})

				When("a terminal is not explicitly requested", func() {
					BeforeEach(func() {
						terminalRequest = RequestTTYAuto
					})

					It("does not request a pty", func() {
						Expect(fakeSecureSession.RequestPtyCallCount()).To(Equal(0))
					})
				})
			})
		})

		When("stdin is not a terminal", func() {
			BeforeEach(func() {
				stdin.ReadStub = func(p []byte) (int, error) {
					return 0, io.EOF
				}

				terminalHelper := DefaultTerminalHelper()
				fakeTerminalHelper.GetFdInfoStub = terminalHelper.GetFdInfo
				fakeTerminalHelper.GetWinsizeStub = terminalHelper.GetWinsize
			})

			When("a terminal is not requested", func() {
				It("does not request a pty", func() {
					Expect(fakeSecureSession.RequestPtyCallCount()).To(Equal(0))
				})
			})

			When("a terminal is requested", func() {
				BeforeEach(func() {
					terminalRequest = RequestTTYYes
				})

				It("does not request a pty", func() {
					Expect(fakeSecureSession.RequestPtyCallCount()).To(Equal(0))
				})
			})
		})

		PWhen("a terminal is forced", func() {
			BeforeEach(func() {
				terminalRequest = RequestTTYForce
			})

			It("requests a pty", func() {
				Expect(fakeSecureSession.RequestPtyCallCount()).To(Equal(1))
			})
		})

		When("a terminal is disabled", func() {
			BeforeEach(func() {
				terminalRequest = RequestTTYNo
			})

			It("does not request a pty", func() {
				Expect(fakeSecureSession.RequestPtyCallCount()).To(Equal(0))
			})
		})

		When("a command is not specified", func() {
			It("requests an interactive shell", func() {
				Expect(fakeSecureSession.ShellCallCount()).To(Equal(1))
			})

			When("the shell request returns an error", func() {
				BeforeEach(func() {
					fakeSecureSession.ShellReturns(errors.New("oh bother"))
				})

				It("returns the error", func() {
					Expect(sessionErr).To(MatchError("oh bother"))
				})
			})
		})

		When("a command is specifed", func() {
			BeforeEach(func() {
				commands = []string{"echo", "-n", "hello"}
			})

			It("starts the command", func() {
				Expect(fakeSecureSession.StartCallCount()).To(Equal(1))
				Expect(fakeSecureSession.StartArgsForCall(0)).To(Equal("echo -n hello"))
			})

			When("the command fails to start", func() {
				BeforeEach(func() {
					fakeSecureSession.StartReturns(errors.New("oh well"))
				})

				It("returns the error", func() {
					Expect(sessionErr).To(MatchError("oh well"))
				})
			})
		})

		When("the shell or command has started", func() {
			BeforeEach(func() {
				stdin.ReadStub = func(p []byte) (int, error) {
					p[0] = 0
					return 1, io.EOF
				}
				stdinPipe.WriteStub = func(p []byte) (int, error) {
					defer GinkgoRecover()
					Expect(p[0]).To(Equal(byte(0)))
					return 1, nil
				}

				stdoutPipe.ReadStub = func(p []byte) (int, error) {
					p[0] = 1
					return 1, io.EOF
				}
				stdout.WriteStub = func(p []byte) (int, error) {
					defer GinkgoRecover()
					Expect(p[0]).To(Equal(byte(1)))
					return 1, nil
				}

				stderrPipe.ReadStub = func(p []byte) (int, error) {
					p[0] = 2
					return 1, io.EOF
				}
				stderr.WriteStub = func(p []byte) (int, error) {
					defer GinkgoRecover()
					Expect(p[0]).To(Equal(byte(2)))
					return 1, nil
				}

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
				Expect(sessionErr).To(MatchError("error result"))
			})

			When("the session terminates before stream copies complete", func() {
				var sessionErrorCh chan error

				BeforeEach(func() {
					sessionErrorCh = make(chan error, 1)

					interactiveSessionInvoker = func(secureShell *SecureShell) {
						go func() {
							sessionErrorCh <- secureShell.InteractiveSession(commands, terminalRequest)
						}()
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

			When("stdin is closed", func() {
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

		When("stdout is a terminal and a window size change occurs", func() {
			var master, slave *os.File

			BeforeEach(func() {
				var err error
				master, slave, err = pty.Open()
				Expect(err).NotTo(HaveOccurred())

				terminalHelper := DefaultTerminalHelper()
				fakeTerminalHelper.GetFdInfoStub = terminalHelper.GetFdInfo
				fakeTerminalHelper.GetWinsizeStub = terminalHelper.GetWinsize
				fakeTerminalHelper.StdStreamsReturns(stdin, slave, stderr)

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

			PIt("sends keep alive messages at the expected interval", func() {
				times := <-timesCh
				Expect(times[2]).To(BeTemporally("~", times[0].Add(200*time.Millisecond), 160*time.Millisecond))
			})
		})
	})

	Describe("LocalPortForward", func() {
		var (
			forwardErr error

			echoAddress  string
			echoListener *fake_net.FakeListener
			echoHandler  *fake_server.FakeConnectionHandler
			echoServer   *server.Server

			localAddress string

			realLocalListener net.Listener
			fakeLocalListener *fake_net.FakeListener

			forwardSpecs []LocalPortForward
		)

		BeforeEach(func() {
			logger := lagertest.NewTestLogger("test")

			var err error
			realLocalListener, err = net.Listen("tcp", "127.0.0.1:0")
			Expect(err).NotTo(HaveOccurred())

			localAddress = realLocalListener.Addr().String()
			fakeListenerFactory.ListenReturns(realLocalListener, nil)

			echoHandler = new(fake_server.FakeConnectionHandler)
			echoHandler.HandleConnectionStub = func(conn net.Conn) {
				_, err = io.Copy(conn, conn)
				Expect(err).ToNot(HaveOccurred())
				conn.Close()
			}

			realListener, err := net.Listen("tcp", "127.0.0.1:0")
			Expect(err).NotTo(HaveOccurred())
			echoAddress = realListener.Addr().String()

			echoListener = new(fake_net.FakeListener)
			echoListener.AcceptStub = realListener.Accept
			echoListener.CloseStub = realListener.Close
			echoListener.AddrStub = realListener.Addr

			fakeLocalListener = new(fake_net.FakeListener)
			fakeLocalListener.AcceptReturns(nil, errors.New("Not Accepting Connections"))

			echoServer = server.NewServer(logger.Session("echo"), "", echoHandler)
			err = echoServer.SetListener(echoListener)
			Expect(err).NotTo(HaveOccurred())
			go echoServer.Serve()

			forwardSpecs = []LocalPortForward{{
				RemoteAddress: echoAddress,
				LocalAddress:  localAddress,
			}}

			fakeSecureClient.DialStub = net.Dial
		})

		JustBeforeEach(func() {
			connectErr := secureShell.Connect(username, passcode, sshEndpoint, sshEndpointFingerprint, skipHostValidation)
			Expect(connectErr).NotTo(HaveOccurred())

			forwardErr = secureShell.LocalPortForward(forwardSpecs)
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
			Expect(forwardErr).NotTo(HaveOccurred())

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

		When("a local connection is already open", func() {
			var conn net.Conn

			JustBeforeEach(func() {
				var err error
				conn, err = net.Dial("tcp", localAddress)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				err := conn.Close()
				Expect(err).NotTo(HaveOccurred())
			})

			It("allows for new incoming connections as well", func() {
				validateConnectivity(localAddress)
			})
		})

		When("there are multiple port forward specs", func() {
			When("provided a real listener", func() {
				var (
					realLocalListener2 net.Listener
					localAddress2      string
				)

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

					forwardSpecs = []LocalPortForward{
						{
							RemoteAddress: echoAddress,
							LocalAddress:  localAddress,
						},
						{
							RemoteAddress: echoAddress,
							LocalAddress:  localAddress2,
						},
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
			})

			When("the secure client is closed", func() {
				var (
					fakeLocalListener1 *fake_net.FakeListener
					fakeLocalListener2 *fake_net.FakeListener
				)

				BeforeEach(func() {
					forwardSpecs = []LocalPortForward{
						{
							RemoteAddress: echoAddress,
							LocalAddress:  localAddress,
						},
						{
							RemoteAddress: echoAddress,
							LocalAddress:  localAddress,
						},
					}

					fakeLocalListener1 = new(fake_net.FakeListener)
					BlockAcceptOnClose(fakeLocalListener1)
					fakeListenerFactory.ListenReturnsOnCall(0, fakeLocalListener1, nil)

					fakeLocalListener2 = new(fake_net.FakeListener)
					BlockAcceptOnClose(fakeLocalListener2)
					fakeListenerFactory.ListenReturnsOnCall(1, fakeLocalListener2, nil)
				})

				It("closes the listeners", func() {
					Eventually(fakeListenerFactory.ListenCallCount).Should(Equal(2))
					Eventually(fakeLocalListener1.AcceptCallCount).Should(Equal(1))
					Eventually(fakeLocalListener2.AcceptCallCount).Should(Equal(1))

					err := secureShell.Close()
					Expect(err).NotTo(HaveOccurred())
					Eventually(fakeLocalListener1.CloseCallCount).Should(Equal(2))
					Eventually(fakeLocalListener2.CloseCallCount).Should(Equal(2))
				})
			})
		})

		When("listen fails", func() {
			BeforeEach(func() {
				fakeListenerFactory.ListenReturns(nil, errors.New("failure is an option"))
			})

			It("returns the error", func() {
				Expect(forwardErr).To(MatchError("failure is an option"))
			})
		})

		When("the client it closed", func() {
			BeforeEach(func() {
				fakeLocalListener = new(fake_net.FakeListener)
				BlockAcceptOnClose(fakeLocalListener)

				fakeListenerFactory.ListenReturns(fakeLocalListener, nil)
			})

			It("closes the listener when the client is closed", func() {
				Eventually(fakeListenerFactory.ListenCallCount).Should(Equal(1))
				Eventually(fakeLocalListener.AcceptCallCount).Should(Equal(1))

				err := secureShell.Close()
				Expect(err).NotTo(HaveOccurred())
				Eventually(fakeLocalListener.CloseCallCount).Should(Equal(2))
			})
		})

		When("accept fails", func() {
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

				PIt("retries connecting after a short delay", func() {
					Eventually(fakeLocalListener.AcceptCallCount).Should(Equal(3))
					Expect(timeCh).To(HaveLen(3))

					times := make([]time.Time, 0)
					for t := range timeCh {
						times = append(times, t)
					}

					Expect(times[1]).To(BeTemporally("~", times[0].Add(115*time.Millisecond), 80*time.Millisecond))
					Expect(times[2]).To(BeTemporally("~", times[1].Add(115*time.Millisecond), 100*time.Millisecond))
				})
			})
		})

		When("dialing the connect address fails", func() {
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
		var waitErr error

		JustBeforeEach(func() {
			connectErr := secureShell.Connect(username, passcode, sshEndpoint, sshEndpointFingerprint, skipHostValidation)
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

			PIt("sends keep alive messages at the expected interval", func() {
				Expect(waitErr).NotTo(HaveOccurred())
				times := <-timesCh
				Expect(times[2]).To(BeTemporally("~", times[0].Add(200*time.Millisecond), 100*time.Millisecond))
			})
		})
	})

	Describe("Close", func() {
		JustBeforeEach(func() {
			connectErr := secureShell.Connect(username, passcode, sshEndpoint, sshEndpointFingerprint, skipHostValidation)
			Expect(connectErr).NotTo(HaveOccurred())
		})

		It("calls close on the secureClient", func() {
			err := secureShell.Close()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeSecureClient.CloseCallCount()).To(Equal(1))
		})
	})
})
