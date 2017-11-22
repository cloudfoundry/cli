// +build !windows

package main

import (
	"code.cloudfoundry.org/diego-ssh/server"
	"code.cloudfoundry.org/lager"
)

func createServer(
	logger lager.Logger,
	address string,
	sshDaemon server.ConnectionHandler,
) (*server.Server, error) {
	return server.NewServer(logger, address, sshDaemon), nil
}
