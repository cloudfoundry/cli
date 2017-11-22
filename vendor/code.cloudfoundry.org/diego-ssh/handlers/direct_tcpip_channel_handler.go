package handlers

import (
	"fmt"
	"net"
	"sync"

	"code.cloudfoundry.org/diego-ssh/helpers"
	"code.cloudfoundry.org/lager"
	"golang.org/x/crypto/ssh"
)

//go:generate counterfeiter -o fakes/fake_dialer.go . Dialer
type Dialer interface {
	Dial(net, addr string) (net.Conn, error)
}

type DirectTcpipChannelHandler struct {
	dialer Dialer
}

func NewDirectTcpipChannelHandler(dialer Dialer) *DirectTcpipChannelHandler {
	return &DirectTcpipChannelHandler{
		dialer: dialer,
	}
}

func (handler *DirectTcpipChannelHandler) HandleNewChannel(logger lager.Logger, newChannel ssh.NewChannel) {
	logger = logger.Session("directtcip-handle-new-channel")
	logger.Debug("starting")
	defer logger.Debug("complete")

	// RFC 4254 Section 7.1
	type channelOpenDirectTcpipMsg struct {
		TargetAddr string
		TargetPort uint32
		OriginAddr string
		OriginPort uint32
	}
	var directTcpipMessage channelOpenDirectTcpipMsg

	err := ssh.Unmarshal(newChannel.ExtraData(), &directTcpipMessage)
	if err != nil {
		logger.Error("failed-unmarshalling-ssh-message", err)
		newChannel.Reject(ssh.ConnectionFailed, "Failed to parse open channel message")
		return
	}

	destination := fmt.Sprintf("%s:%d", directTcpipMessage.TargetAddr, directTcpipMessage.TargetPort)
	conn, err := handler.dialer.Dial("tcp", destination)
	if err != nil {
		logger.Error("failed-connecting-to-target", err)
		newChannel.Reject(ssh.ConnectionFailed, err.Error())
		return
	}

	channel, requests, err := newChannel.Accept()
	go ssh.DiscardRequests(requests)

	wg := &sync.WaitGroup{}

	wg.Add(2)

	defer func() {
		conn.Close()
		channel.Close()
	}()

	go helpers.CopyAndClose(logger.Session("to-target"), wg, conn, channel,
		func() {
			conn.(*net.TCPConn).CloseWrite()
		},
	)
	go helpers.CopyAndClose(logger.Session("to-channel"), wg, channel, conn,
		func() {
			channel.CloseWrite()
		},
	)

	wg.Wait()
}
