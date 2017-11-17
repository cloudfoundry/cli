// +build !windows

package ntlm

import (
	"errors"
	"net"
)

func ProxySetup(conn net.Conn, targetAddr string) error {
	return errors.New("NTLM currently only supported on Windows")
}
