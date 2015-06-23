package test_rpc_server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
)

type Handlers interface {
	IsMinCliVersion(args string, retVal *bool) error
}

type TestServer struct {
	listener net.Listener
	Handlers Handlers
	stopCh   chan struct{}
}

func NewTestRpcServer(handlers Handlers) (*TestServer, error) {
	ts := &TestServer{
		Handlers: handlers,
	}

	// discard the warning about non-rpc method in counterfeiter fakes module
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stdout)

	rpc.DefaultServer = rpc.NewServer()
	err := rpc.RegisterName("CliRpcCmd", ts.Handlers)
	if err != nil {
		return nil, err
	}

	return ts, nil
}

func (ts *TestServer) Stop() {
	close(ts.stopCh)
	ts.listener.Close()
}

func (ts *TestServer) Port() string {
	return strconv.Itoa(ts.listener.Addr().(*net.TCPAddr).Port)
}

func (ts *TestServer) Start() error {
	var err error

	ts.stopCh = make(chan struct{})

	ts.listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := ts.listener.Accept()
			if err != nil {
				select {
				case <-ts.stopCh:
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
