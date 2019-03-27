package clissh

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"code.cloudfoundry.org/cli/cf/ssh/sigwinch"
	"code.cloudfoundry.org/cli/util/clissh/ssherror"
	"github.com/moby/moby/pkg/term"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const (
	md5FingerprintLength          = 47 // inclusive of space between bytes
	hexSha1FingerprintLength      = 59 // inclusive of space between bytes
	base64Sha256FingerprintLength = 43

	DefaultKeepAliveInterval = 30 * time.Second
)

type LocalPortForward struct {
	LocalAddress  string
	RemoteAddress string
}

type SecureShell struct {
	secureDialer    SecureDialer
	secureClient    SecureClient
	terminalHelper  TerminalHelper
	listenerFactory ListenerFactory

	localListeners    []net.Listener
	keepAliveInterval time.Duration
}

func NewDefaultSecureShell() *SecureShell {
	defaultSecureDialer := DefaultSecureDialer()
	defaultTerminalHelper := DefaultTerminalHelper()
	defaultListenerFactory := DefaultListenerFactory()
	return &SecureShell{
		secureDialer:      defaultSecureDialer,
		terminalHelper:    defaultTerminalHelper,
		listenerFactory:   defaultListenerFactory,
		keepAliveInterval: DefaultKeepAliveInterval,
		localListeners:    []net.Listener{},
	}
}

func NewSecureShell(
	secureDialer SecureDialer,
	terminalHelper TerminalHelper,
	listenerFactory ListenerFactory,
	keepAliveInterval time.Duration,
) *SecureShell {
	return &SecureShell{
		secureDialer:      secureDialer,
		terminalHelper:    terminalHelper,
		listenerFactory:   listenerFactory,
		keepAliveInterval: keepAliveInterval,
		localListeners:    []net.Listener{},
	}
}

func (c *SecureShell) Close() error {
	for _, listener := range c.localListeners {
		listener.Close()
	}
	return c.secureClient.Close()
}

func (c *SecureShell) Connect(username string, passcode string, appSSHEndpoint string, appSSHHostKeyFingerprint string, skipHostValidation bool) error {
	hostKeyCallbackFunction := fingerprintCallback(skipHostValidation, appSSHHostKeyFingerprint)

	clientConfig := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(passcode)},
		HostKeyCallback: hostKeyCallbackFunction,
	}

	secureClient, err := c.secureDialer.Dial("tcp", appSSHEndpoint, clientConfig)
	if err != nil {
		if strings.Contains(err.Error(), "ssh: unable to authenticate") {
			return ssherror.UnableToAuthenticateError{Err: err}
		}
		return err
	}

	c.secureClient = secureClient
	return nil
}

func (c *SecureShell) InteractiveSession(commands []string, terminalRequest TTYRequest) error {
	session, err := c.secureClient.NewSession()
	if err != nil {
		return fmt.Errorf("SSH session allocation failed: %s", err.Error())
	}
	defer session.Close()

	stdin, stdout, stderr := c.terminalHelper.StdStreams()

	inPipe, err := session.StdinPipe()
	if err != nil {
		return err
	}

	outPipe, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	errPipe, err := session.StderrPipe()
	if err != nil {
		return err
	}

	stdinFd, stdinIsTerminal := c.terminalHelper.GetFdInfo(stdin)
	stdoutFd, stdoutIsTerminal := c.terminalHelper.GetFdInfo(stdout)

	if c.shouldAllocateTerminal(commands, terminalRequest, stdinIsTerminal) {
		modes := ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 115200,
			ssh.TTY_OP_OSPEED: 115200,
		}

		width, height := c.getWindowDimensions(stdoutFd)

		err = session.RequestPty(c.terminalType(), height, width, modes)
		if err != nil {
			return err
		}

		var state *term.State
		state, err = c.terminalHelper.SetRawTerminal(stdinFd)
		if err == nil {
			defer func() {
				err := c.terminalHelper.RestoreTerminal(stdinFd, state)
				log.Errorln("restore terminal", err)
			}()
		}
	}

	if len(commands) > 0 {
		cmd := strings.Join(commands, " ")
		err = session.Start(cmd)
		if err != nil {
			return err
		}
	} else {
		err = session.Shell()
		if err != nil {
			return err
		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go copyAndClose(nil, inPipe, stdin)
	go copyAndDone(wg, stdout, outPipe)
	go copyAndDone(wg, stderr, errPipe)

	if stdoutIsTerminal {
		resized := make(chan os.Signal, 16)

		if runtime.GOOS == "windows" {
			ticker := time.NewTicker(250 * time.Millisecond)
			defer ticker.Stop()

			go func() {
				for range ticker.C {
					resized <- syscall.Signal(-1)
				}
				close(resized)
			}()
		} else {
			signal.Notify(resized, sigwinch.SIGWINCH())
			defer func() { signal.Stop(resized); close(resized) }()
		}

		go c.resize(resized, session, stdoutFd)
	}

	keepaliveStopCh := make(chan struct{})
	defer close(keepaliveStopCh)

	go keepalive(c.secureClient.Conn(), time.NewTicker(c.keepAliveInterval), keepaliveStopCh)

	result := session.Wait()
	wg.Wait()
	return result
}

func (c *SecureShell) LocalPortForward(localPortForwardSpecs []LocalPortForward) error {
	for _, spec := range localPortForwardSpecs {
		listener, err := c.listenerFactory.Listen("tcp", spec.LocalAddress)
		if err != nil {
			return err
		}
		c.localListeners = append(c.localListeners, listener)

		go c.localForwardAcceptLoop(listener, spec.RemoteAddress)
	}

	return nil
}

func (c *SecureShell) Wait() error {
	keepaliveStopCh := make(chan struct{})
	defer close(keepaliveStopCh)

	go keepalive(c.secureClient.Conn(), time.NewTicker(c.keepAliveInterval), keepaliveStopCh)

	return c.secureClient.Wait()
}

func (c *SecureShell) getWindowDimensions(terminalFd uintptr) (width int, height int) {
	winSize, err := c.terminalHelper.GetWinsize(terminalFd)
	if err != nil {
		winSize = &term.Winsize{
			Width:  80,
			Height: 43,
		}
	}

	return int(winSize.Width), int(winSize.Height)
}

func (c *SecureShell) handleForwardConnection(conn net.Conn, targetAddr string) {
	defer conn.Close()

	target, err := c.secureClient.Dial("tcp", targetAddr)
	if err != nil {
		fmt.Printf("connect to %s failed: %s\n", targetAddr, err.Error())
		return
	}
	defer target.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go copyAndClose(wg, conn, target)
	go copyAndClose(wg, target, conn)
	wg.Wait()
}

func (c *SecureShell) localForwardAcceptLoop(listener net.Listener, addr string) {
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return
		}

		go c.handleForwardConnection(conn, addr)
	}
}

func (c *SecureShell) resize(resized <-chan os.Signal, session SecureSession, terminalFd uintptr) {
	type resizeMessage struct {
		Width       uint32
		Height      uint32
		PixelWidth  uint32
		PixelHeight uint32
	}

	var previousWidth, previousHeight int

	for range resized {
		width, height := c.getWindowDimensions(terminalFd)

		if width == previousWidth && height == previousHeight {
			continue
		}

		message := resizeMessage{
			Width:  uint32(width),
			Height: uint32(height),
		}

		_, err := session.SendRequest("window-change", false, ssh.Marshal(message))
		if err != nil {
			log.Errorln("window-change:", err)
		}

		previousWidth = width
		previousHeight = height
	}
}

func (c *SecureShell) shouldAllocateTerminal(commands []string, terminalRequest TTYRequest, stdinIsTerminal bool) bool {
	switch terminalRequest {
	case RequestTTYForce:
		return true
	case RequestTTYNo:
		return false
	case RequestTTYYes:
		return stdinIsTerminal
	case RequestTTYAuto:
		return len(commands) == 0 && stdinIsTerminal
	default:
		return false
	}
}

func (c *SecureShell) terminalType() string {
	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm"
	}
	return term
}

func base64Sha256Fingerprint(key ssh.PublicKey) string {
	sum := sha256.Sum256(key.Marshal())
	return base64.RawStdEncoding.EncodeToString(sum[:])
}

func copyAndClose(wg *sync.WaitGroup, dest io.WriteCloser, src io.Reader) {
	_, err := io.Copy(dest, src)
	if err != nil {
		log.Errorln("copy and close:", err)
	}
	_ = dest.Close()
	if wg != nil {
		wg.Done()
	}
}

func copyAndDone(wg *sync.WaitGroup, dest io.Writer, src io.Reader) {
	_, err := io.Copy(dest, src)
	if err != nil {
		log.Errorln("copy and done:", err)
	}
	wg.Done()
}

func fingerprintCallback(skipHostValidation bool, expectedFingerprint string) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if skipHostValidation {
			return nil
		}

		var fingerprint string

		switch len(expectedFingerprint) {
		case base64Sha256FingerprintLength:
			fingerprint = base64Sha256Fingerprint(key)
		case hexSha1FingerprintLength:
			fingerprint = hexSha1Fingerprint(key)
		case md5FingerprintLength:
			fingerprint = md5Fingerprint(key)
		case 0:
			fingerprint = md5Fingerprint(key)
			return fmt.Errorf("Unable to verify identity of host.\n\nThe fingerprint of the received key was %q.", fingerprint)
		default:
			return errors.New("Unsupported host key fingerprint format")
		}

		if fingerprint != expectedFingerprint {
			return fmt.Errorf("Host key verification failed.\n\nThe fingerprint of the received key was %q.", fingerprint)
		}
		return nil
	}
}

func hexSha1Fingerprint(key ssh.PublicKey) string {
	sum := sha1.Sum(key.Marshal())
	return strings.Replace(fmt.Sprintf("% x", sum), " ", ":", -1)
}

func keepalive(conn ssh.Conn, ticker *time.Ticker, stopCh chan struct{}) {
	for {
		select {
		case <-ticker.C:
			_, _, err := conn.SendRequest("keepalive@cloudfoundry.org", true, nil)
			if err != nil {
				log.Errorln("err sending keep alive:", err)
			}
		case <-stopCh:
			ticker.Stop()
			return
		}
	}
}

func md5Fingerprint(key ssh.PublicKey) string {
	sum := md5.Sum(key.Marshal())
	return strings.Replace(fmt.Sprintf("% x", sum), " ", ":", -1)
}
