package rpc

import (
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/codegangsta/cli"

	"fmt"
	"net"
	"net/rpc"
	"strconv"
)

type CliRpcService struct {
	listener net.Listener
	stopCh   chan struct{}
	Pinged   bool
	RpcCmd   *CliRpcCmd
}

type CliRpcCmd struct {
	ReturnData        interface{}
	coreCommandRunner *cli.App
	outputCapture     terminal.OutputCapture
}

func NewRpcService(commandRunner *cli.App, outputCapture terminal.OutputCapture) (*CliRpcService, error) {
	rpcService := &CliRpcService{
		RpcCmd: &CliRpcCmd{
			ReturnData:        new(interface{}),
			coreCommandRunner: commandRunner,
			outputCapture:     outputCapture,
		},
	}

	err := rpc.Register(rpcService.RpcCmd)
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
				go rpc.ServeConn(conn)
			}
		}
	}()

	return nil
}

func (cmd *CliRpcCmd) SetPluginMetadata(pluginMetadata plugin.PluginMetadata, retVal *bool) error {
	cmd.ReturnData = interface{}(pluginMetadata)
	*retVal = true
	return nil
}

func (cmd *CliRpcCmd) CallCoreCommand(args []string, retVal *bool) error {
	defer func() {
		recover()
	}()

	err := cmd.coreCommandRunner.Run(append([]string{"cf"}, args...))

	if err != nil {
		*retVal = false
		return err
	}

	*retVal = true
	return nil
}

func (cmd *CliRpcCmd) GetOutputAndReset(args bool, retVal *[]string) error {
	*retVal = cmd.outputCapture.GetOutputAndReset()
	return nil
}
