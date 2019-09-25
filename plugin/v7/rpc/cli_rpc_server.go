// +build V7

package rpc

import (
	"os"
	"strings"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin/v7"
	"code.cloudfoundry.org/cli/plugin/v7/models"
	"code.cloudfoundry.org/cli/version"
	"github.com/blang/semver"

	"fmt"
	"net"
	"net/rpc"
	"strconv"

	"bytes"
	"io"

	"sync"

	"code.cloudfoundry.org/cli/cf/trace"
)

var dialTimeout = os.Getenv("CF_DIAL_TIMEOUT")

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

//go:generate counterfeiter . OutputCapture

type OutputCapture interface {
	SetOutputBucket(io.Writer)
}

func NewRpcService(
	outputCapture OutputCapture,
	terminalOutputSwitch TerminalOutputSwitch,
	cliConfig coreconfig.Repository,
	repoLocator api.RepositoryLocator,
	newCmdRunner CommandRunner,
	logger trace.Printer,
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
			cliConfig:            cliConfig,
			repoLocator:          repoLocator,
			newCmdRunner:         newCmdRunner,
			logger:               logger,
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

