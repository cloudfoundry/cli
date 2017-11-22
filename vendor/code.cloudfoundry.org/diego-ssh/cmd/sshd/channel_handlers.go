// +build !windows

package main

import (
	"net"
	"time"

	"code.cloudfoundry.org/diego-ssh/handlers"
)

func newChannelHandlers() map[string]handlers.NewChannelHandler {
	runner := handlers.NewCommandRunner()
	shellLocator := handlers.NewShellLocator()
	dialer := &net.Dialer{}

	return map[string]handlers.NewChannelHandler{
		"session":      handlers.NewSessionChannelHandler(runner, shellLocator, getDaemonEnvironment(), 15*time.Second),
		"direct-tcpip": handlers.NewDirectTcpipChannelHandler(dialer),
	}
}
