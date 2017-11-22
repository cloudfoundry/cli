// +build !windows

package handlers

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"syscall"
	"time"

	"code.cloudfoundry.org/diego-ssh/helpers"
	"code.cloudfoundry.org/diego-ssh/scp"
	"code.cloudfoundry.org/diego-ssh/signals"
	"code.cloudfoundry.org/diego-ssh/termcodes"
	"code.cloudfoundry.org/lager"
	"github.com/docker/docker/pkg/term"
	"github.com/kr/pty"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var scpRegex = regexp.MustCompile(`^\s*scp($|\s+)`)

type SessionChannelHandler struct {
	runner       Runner
	shellLocator ShellLocator
	defaultEnv   map[string]string
	keepalive    time.Duration
}

func NewSessionChannelHandler(
	runner Runner,
	shellLocator ShellLocator,
	defaultEnv map[string]string,
	keepalive time.Duration,
) *SessionChannelHandler {
	return &SessionChannelHandler{
		runner:       runner,
		shellLocator: shellLocator,
		defaultEnv:   defaultEnv,
		keepalive:    keepalive,
	}
}

func (handler *SessionChannelHandler) HandleNewChannel(logger lager.Logger, newChannel ssh.NewChannel) {
	channel, requests, err := newChannel.Accept()
	if err != nil {
		logger.Error("handle-new-session-channel-failed", err)
		return
	}

	handler.newSession(logger, channel, handler.keepalive).serviceRequests(requests)
}

type ptyRequestMsg struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	Modelist string
}

type session struct {
	logger            lager.Logger
	complete          bool
	keepaliveDuration time.Duration
	keepaliveStopCh   chan struct{}

	shellPath string
	runner    Runner
	channel   ssh.Channel

	sync.Mutex
	env     map[string]string
	command *exec.Cmd

	wg         sync.WaitGroup
	allocPty   bool
	ptyRequest ptyRequestMsg

	ptyMaster *os.File
}

func (handler *SessionChannelHandler) newSession(logger lager.Logger, channel ssh.Channel, keepalive time.Duration) *session {
	return &session{
		logger:            logger.Session("session-channel"),
		keepaliveDuration: keepalive,
		runner:            handler.runner,
		shellPath:         handler.shellLocator.ShellPath(),
		channel:           channel,
		env:               handler.defaultEnv,
	}
}

func (sess *session) serviceRequests(requests <-chan *ssh.Request) {
	logger := sess.logger
	logger.Info("starting")
	defer logger.Info("finished")

	defer sess.destroy()

	for req := range requests {
		sess.logger.Info("received-request", lager.Data{"type": req.Type})
		switch req.Type {
		case "env":
			sess.handleEnvironmentRequest(req)
		case "signal":
			sess.handleSignalRequest(req)
		case "pty-req":
			sess.handlePtyRequest(req)
		case "window-change":
			sess.handleWindowChangeRequest(req)
		case "exec":
			sess.handleExecRequest(req)
		case "shell":
			sess.handleShellRequest(req)
		case "subsystem":
			sess.handleSubsystemRequest(req)
		default:
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

func (sess *session) handleEnvironmentRequest(request *ssh.Request) {
	logger := sess.logger.Session("handle-environment-request")

	type envMsg struct {
		Name  string
		Value string
	}
	var envMessage envMsg

	err := ssh.Unmarshal(request.Payload, &envMessage)
	if err != nil {
		logger.Error("unmarshal-failed", err)
		request.Reply(false, nil)
		return
	}

	sess.Lock()
	sess.env[envMessage.Name] = envMessage.Value
	sess.Unlock()

	if request.WantReply {
		request.Reply(true, nil)
	}
}

func (sess *session) handleSignalRequest(request *ssh.Request) {
	logger := sess.logger.Session("handle-signal-request")

	type signalMsg struct {
		Signal string
	}
	var signalMessage signalMsg

	err := ssh.Unmarshal(request.Payload, &signalMessage)
	if err != nil {
		logger.Error("unmarshal-failed", err)
		if request.WantReply {
			request.Reply(false, nil)
		}
		return
	}

	sess.Lock()
	defer sess.Unlock()

	cmd := sess.command

	if cmd != nil {
		signal := signals.SyscallSignals[ssh.Signal(signalMessage.Signal)]
		err := sess.runner.Signal(cmd, signal)
		if err != nil {
			logger.Error("process-signal-failed", err)
		}
	}

	if request.WantReply {
		request.Reply(true, nil)
	}
}

func (sess *session) handlePtyRequest(request *ssh.Request) {
	logger := sess.logger.Session("handle-pty-request")

	var ptyRequestMessage ptyRequestMsg

	err := ssh.Unmarshal(request.Payload, &ptyRequestMessage)
	if err != nil {
		logger.Error("unmarshal-failed", err)
		if request.WantReply {
			request.Reply(false, nil)
		}
		return
	}

	sess.Lock()
	defer sess.Unlock()

	sess.allocPty = true
	sess.ptyRequest = ptyRequestMessage
	sess.env["TERM"] = ptyRequestMessage.Term

	if request.WantReply {
		request.Reply(true, nil)
	}
}

func (sess *session) handleWindowChangeRequest(request *ssh.Request) {
	logger := sess.logger.Session("handle-window-change")

	type windowChangeMsg struct {
		Columns  uint32
		Rows     uint32
		WidthPx  uint32
		HeightPx uint32
	}
	var windowChangeMessage windowChangeMsg

	err := ssh.Unmarshal(request.Payload, &windowChangeMessage)
	if err != nil {
		logger.Error("unmarshal-failed", err)
		if request.WantReply {
			request.Reply(false, nil)
		}
		return
	}

	sess.Lock()
	defer sess.Unlock()

	if sess.allocPty {
		sess.ptyRequest.Columns = windowChangeMessage.Columns
		sess.ptyRequest.Rows = windowChangeMessage.Rows
	}

	if sess.ptyMaster != nil {
		err = setWindowSize(logger, sess.ptyMaster, sess.ptyRequest.Columns, sess.ptyRequest.Rows)
		if err != nil {
			logger.Error("failed-to-set-window-size", err)
		}
	}

	if request.WantReply {
		request.Reply(true, nil)
	}
}

func (sess *session) handleExecRequest(request *ssh.Request) {
	logger := sess.logger.Session("handle-exec-request")

	type execMsg struct {
		Command string
	}
	var execMessage execMsg

	err := ssh.Unmarshal(request.Payload, &execMessage)
	if err != nil {
		logger.Error("unmarshal-failed", err)
		if request.WantReply {
			request.Reply(false, nil)
		}
		return
	}

	if scpRegex.MatchString(execMessage.Command) {
		logger.Info("handling-scp-command", lager.Data{"Command": execMessage.Command})
		sess.executeSCP(execMessage.Command, request)
	} else {
		sess.executeShell(request, "-c", execMessage.Command)
	}
}

func (sess *session) handleShellRequest(request *ssh.Request) {
	sess.executeShell(request)
}

func (sess *session) handleSubsystemRequest(request *ssh.Request) {
	logger := sess.logger.Session("handle-subsystem-request")
	logger.Info("starting")
	defer logger.Info("finished")

	type subsysMsg struct {
		Subsystem string
	}
	var subsystemMessage subsysMsg

	err := ssh.Unmarshal(request.Payload, &subsystemMessage)
	if err != nil {
		logger.Error("unmarshal-failed", err)
		if request.WantReply {
			request.Reply(false, nil)
		}
		return
	}

	if subsystemMessage.Subsystem != "sftp" {
		logger.Info("unsupported-subsystem", lager.Data{"subsystem": subsystemMessage.Subsystem})
		if request.WantReply {
			request.Reply(false, nil)
		}
		return
	}

	lagerWriter := helpers.NewLagerWriter(logger.Session("sftp-server"))
	sftpServer, err := sftp.NewServer(sess.channel, sess.channel, sftp.WithDebug(lagerWriter))
	if err != nil {
		logger.Error("sftp-new-server-failed", err)
		if request.WantReply {
			request.Reply(false, nil)
		}
		return
	}

	if request.WantReply {
		request.Reply(true, nil)
	}

	logger.Info("starting-server")
	go func() {
		defer sess.destroy()
		err = sftpServer.Serve()
		if err != nil {
			logger.Error("sftp-serve-error", err)
		}
	}()
}

func (sess *session) executeShell(request *ssh.Request, args ...string) {
	logger := sess.logger.Session("execute-shell")

	sess.Lock()
	cmd, err := sess.createCommand(args...)
	if err != nil {
		sess.Unlock()
		logger.Error("failed-to-create-command", err)
		if request.WantReply {
			request.Reply(false, nil)
		}
		return
	}

	if request.WantReply {
		request.Reply(true, nil)
	}

	if sess.allocPty {
		err = sess.runWithPty(cmd)
	} else {
		err = sess.run(cmd)
	}

	sess.Unlock()

	if err != nil {
		sess.sendExitMessage(err)
		sess.destroy()
		return
	}

	go func() {
		err := sess.wait(cmd)
		sess.sendExitMessage(err)
		sess.destroy()
	}()
}

func (sess *session) createCommand(args ...string) (*exec.Cmd, error) {
	if sess.command != nil {
		return nil, errors.New("command already started")
	}

	cmd := exec.Command(sess.shellPath, args...)
	cmd.Env = sess.environment()
	sess.command = cmd

	return cmd, nil
}

func (sess *session) environment() []string {
	env := []string{}

	env = append(env, "PATH=/bin:/usr/bin")
	env = append(env, "LANG=en_US.UTF8")

	for k, v := range sess.env {
		if k != "HOME" && k != "USER" {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	env = append(env, fmt.Sprintf("HOME=%s", os.Getenv("HOME")))
	env = append(env, fmt.Sprintf("USER=%s", os.Getenv("USER")))

	return env
}

type exitStatusMsg struct {
	Status uint32
}

type exitSignalMsg struct {
	Signal     string
	CoreDumped bool
	Error      string
	Lang       string
}

func (sess *session) sendExitMessage(err error) {
	logger := sess.logger.Session("send-exit-message")
	logger.Info("started")
	defer logger.Info("finished")

	if err != nil {
		logger.Error("building-exit-message-from-error", err)
	}

	if err == nil {
		_, sendErr := sess.channel.SendRequest("exit-status", false, ssh.Marshal(exitStatusMsg{}))
		if sendErr != nil {
			logger.Error("send-exit-status-failed", sendErr)
		}
		return
	}

	exitError, ok := err.(*exec.ExitError)
	if !ok {
		exitMessage := exitStatusMsg{Status: 255}
		_, sendErr := sess.channel.SendRequest("exit-status", false, ssh.Marshal(exitMessage))
		if sendErr != nil {
			logger.Error("send-exit-status-failed", sendErr)
		}
		return
	}

	waitStatus, ok := exitError.Sys().(syscall.WaitStatus)
	if !ok {
		exitMessage := exitStatusMsg{Status: 255}
		_, sendErr := sess.channel.SendRequest("exit-status", false, ssh.Marshal(exitMessage))
		if sendErr != nil {
			logger.Error("send-exit-status-failed", sendErr)
		}
		return
	}

	if waitStatus.Signaled() {
		exitMessage := exitSignalMsg{
			Signal:     string(signals.SSHSignals[waitStatus.Signal()]),
			CoreDumped: waitStatus.CoreDump(),
		}
		_, sendErr := sess.channel.SendRequest("exit-signal", false, ssh.Marshal(exitMessage))
		if sendErr != nil {
			logger.Error("send-exit-status-failed", sendErr)
		}
		return
	}

	exitMessage := exitStatusMsg{Status: uint32(waitStatus.ExitStatus())}
	_, sendErr := sess.channel.SendRequest("exit-status", false, ssh.Marshal(exitMessage))
	if sendErr != nil {
		logger.Error("send-exit-status-failed", sendErr)
	}
}

func setWindowSize(logger lager.Logger, pseudoTty *os.File, columns, rows uint32) error {
	logger.Info("new-size", lager.Data{"columns": columns, "rows": rows})
	return term.SetWinsize(pseudoTty.Fd(), &term.Winsize{
		Width:  uint16(columns),
		Height: uint16(rows),
	})
}

func setTerminalAttributes(logger lager.Logger, pseudoTty *os.File, modelist string) {
	reader := bytes.NewReader([]byte(modelist))

	for {
		var opcode uint8
		var value uint32

		err := binary.Read(reader, binary.BigEndian, &opcode)
		if err != nil {
			logger.Error("failed-to-read-modelist-opcode", err)
			break
		}

		if opcode == 0 || opcode >= 160 {
			break
		}

		err = binary.Read(reader, binary.BigEndian, &value)
		if err != nil {
			logger.Error("failed-to-read-modelist-value", err)
			break
		}

		logger.Info("set-terminal-attribute", lager.Data{
			"opcode": opcode,
			"value":  fmt.Sprintf("%x", value),
		})

		termios, err := termcodes.GetAttr(pseudoTty)
		if err != nil {
			logger.Error("failed-to-get-terminal-attrs", err)
			continue
		}

		err = termcodes.TermAttrSetters[opcode].Set(pseudoTty, termios, value)
		if err != nil {
			logger.Error("failed-to-set-terminal-attrs", err, lager.Data{
				"opcode": opcode,
				"value":  fmt.Sprintf("%x", value),
			})
			continue
		}
	}
}

func (sess *session) run(command *exec.Cmd) error {
	logger := sess.logger.Session("run")

	command.Stdout = sess.channel
	command.Stderr = sess.channel.Stderr()

	stdin, err := command.StdinPipe()
	if err != nil {
		return err
	}

	go helpers.CopyAndClose(logger.Session("to-stdin"), nil, stdin, sess.channel, func() { stdin.Close() })

	return sess.runner.Start(command)
}

func (sess *session) runWithPty(command *exec.Cmd) error {
	logger := sess.logger.Session("run-with-pty")

	ptyMaster, ptySlave, err := pty.Open()
	if err != nil {
		logger.Error("failed-to-open-pty", err)
		return err
	}

	sess.ptyMaster = ptyMaster
	defer ptySlave.Close()

	command.Stdout = ptySlave
	command.Stdin = ptySlave
	command.Stderr = ptySlave

	command.SysProcAttr = &syscall.SysProcAttr{
		Setctty: true,
		Setsid:  true,
	}

	setTerminalAttributes(logger, ptyMaster, sess.ptyRequest.Modelist)
	setWindowSize(logger, ptyMaster, sess.ptyRequest.Columns, sess.ptyRequest.Rows)

	sess.wg.Add(1)
	go helpers.Copy(logger.Session("to-pty"), nil, ptyMaster, sess.channel)
	go func() {
		helpers.Copy(logger.Session("from-pty"), &sess.wg, sess.channel, ptyMaster)
		sess.channel.CloseWrite()
	}()

	err = sess.runner.Start(command)
	if err == nil {
		sess.keepaliveStopCh = make(chan struct{})
		go sess.keepalive(command, sess.keepaliveStopCh)
	}
	return err
}

func (sess *session) keepalive(command *exec.Cmd, stopCh chan struct{}) {
	logger := sess.logger.Session("keepalive")

	ticker := time.NewTicker(sess.keepaliveDuration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_, err := sess.channel.SendRequest("keepalive@cloudfoundry.org", true, nil)
			logger.Info("keepalive", lager.Data{"success": err == nil})

			if err != nil {
				err = sess.runner.Signal(command, syscall.SIGHUP)
				logger.Info("process-signaled", lager.Data{"error": err})
				return
			}
		case <-stopCh:
			return
		}
	}
}

func (sess *session) wait(command *exec.Cmd) error {
	logger := sess.logger.Session("wait")
	logger.Info("started")
	defer logger.Info("done")
	return sess.runner.Wait(command)
}

func (sess *session) destroy() {
	logger := sess.logger.Session("destroy")
	logger.Info("started")
	defer logger.Info("done")

	sess.Lock()
	defer sess.Unlock()

	if sess.complete {
		return
	}

	sess.complete = true
	sess.wg.Wait()

	if sess.channel != nil {
		sess.channel.Close()
	}

	if sess.ptyMaster != nil {
		sess.ptyMaster.Close()
		sess.ptyMaster = nil
	}

	if sess.keepaliveStopCh != nil {
		close(sess.keepaliveStopCh)
	}
}

func (sess *session) executeSCP(command string, request *ssh.Request) {
	logger := sess.logger.Session("execute-scp")

	if request.WantReply {
		request.Reply(true, nil)
	}

	copier, err := scp.NewFromCommand(command, sess.channel, sess.channel, sess.channel.Stderr(), logger)
	if err == nil {
		err = copier.Copy()
	}

	sess.sendSCPExitMessage(err)
	sess.destroy()
}

func (sess *session) sendSCPExitMessage(err error) {
	logger := sess.logger.Session("send-scp-exit-message")
	logger.Info("started")
	defer logger.Info("finished")

	var exitMessage exitStatusMsg
	if err != nil {
		logger.Error("building-scp-exit-message-from-error", err)
		exitMessage = exitStatusMsg{Status: 1}
	}

	_, sendErr := sess.channel.SendRequest("exit-status", false, ssh.Marshal(exitMessage))
	if sendErr != nil {
		logger.Error("send-exit-status-failed", sendErr)
	}
}
