// +build V7

package rpc

import (
	plugin "code.cloudfoundry.org/cli/plugin/v7"

	"fmt"
	"net"
	"net/rpc"
	"strconv"

	"bytes"
	"io"

	"sync"
)

type CliRpcService struct {
	listener net.Listener
	stopCh   chan struct{}
	Pinged   bool
	RpcCmd   *CliRpcCmd
	Server   *rpc.Server
}

//go:generate counterfeiter . TerminalOutputSwitch

type TerminalOutputSwitch interface {
	DisableTerminalOutput(bool)
}

type OutputCapture interface {
	SetOutputBucket(io.Writer)
}

func NewRpcService(
	outputCapture OutputCapture,
	terminalOutputSwitch TerminalOutputSwitch,
	w io.Writer,
	rpcServer *rpc.Server,
) (*CliRpcService, error) {
	rpcService := &CliRpcService{
		Server: rpcServer,
		RpcCmd: &CliRpcCmd{
			PluginMetadata:       &plugin.PluginMetadata{},
			MetadataMutex:        &sync.RWMutex{},
			outputCapture:        outputCapture,
			terminalOutputSwitch: terminalOutputSwitch,
			outputBucket:         &bytes.Buffer{},
			stdout:               w,
		},
	}

	err := rpcService.Server.Register(rpcService.RpcCmd)
	if err != nil {
		return nil, err
	}

	return rpcService, nil
}

func (cli *CliRpcService) Stop() {
	close(cli.stopCh)
	cli.listener.Close()
}

func (cli *CliRpcService) Port() string {
	return strconv.Itoa(cli.listener.Addr().(*net.TCPAddr).Port)
}

func (cli *CliRpcService) Start() error {
	var err error

	cli.stopCh = make(chan struct{})

	cli.listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := cli.listener.Accept()
			if err != nil {
				select {
				case <-cli.stopCh:
					return
				default:
					fmt.Println(err)
				}
			} else {
				go cli.Server.ServeConn(conn)
			}
		}
	}()

	return nil
}
