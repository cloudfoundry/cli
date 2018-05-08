package rpcserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"

	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/plugin/models"
)

//go:generate counterfeiter . Handlers

type Handlers interface {
	IsMinCliVersion(args string, retVal *bool) error
	SetPluginMetadata(pluginMetadata plugin.PluginMetadata, retVal *bool) error
	DisableTerminalOutput(disable bool, retVal *bool) error
	CallCoreCommand(args []string, retVal *bool) error
	GetOutputAndReset(args bool, retVal *[]string) error
	GetCurrentOrg(args string, retVal *plugin_models.Organization) error
	GetCurrentSpace(args string, retVal *plugin_models.Space) error
	Username(args string, retVal *string) error
	UserGuid(args string, retVal *string) error
	UserEmail(args string, retVal *string) error
	IsLoggedIn(args string, retVal *bool) error
	IsSSLDisabled(args string, retVal *bool) error
	HasOrganization(args string, retVal *bool) error
	HasSpace(args string, retVal *bool) error
	ApiEndpoint(args string, retVal *string) error
	HasAPIEndpoint(args string, retVal *bool) error
	ApiVersion(args string, retVal *string) error
	LoggregatorEndpoint(args string, retVal *string) error
	DopplerEndpoint(args string, retVal *string) error
	AccessToken(args string, retVal *string) error
	GetApp(appName string, retVal *plugin_models.GetAppModel) error
	GetApps(args string, retVal *[]plugin_models.GetAppsModel) error
	GetOrgs(args string, retVal *[]plugin_models.GetOrgs_Model) error
	GetSpaces(args string, retVal *[]plugin_models.GetSpaces_Model) error
	GetServices(args string, retVal *[]plugin_models.GetServices_Model) error
	GetOrgUsers(args []string, retVal *[]plugin_models.GetOrgUsers_Model) error
	GetSpaceUsers(args []string, retVal *[]plugin_models.GetSpaceUsers_Model) error
	GetOrg(orgName string, retVal *plugin_models.GetOrg_Model) error
	GetSpace(spaceName string, retVal *plugin_models.GetSpace_Model) error
	GetService(serviceInstance string, retVal *plugin_models.GetService_Model) error
}

type TestServer struct {
	listener net.Listener
	Handlers Handlers
	stopCh   chan struct{}
	server   *rpc.Server
}

func NewTestRPCServer(handlers Handlers) (*TestServer, error) {
	ts := &TestServer{
		Handlers: handlers,
	}

	// discard the warning about non-rpc method in counterfeiter fakes module
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stdout)

	server := rpc.NewServer()
	err := server.RegisterName("CliRpcCmd", ts.Handlers)
	if err != nil {
		return nil, err
	}

	ts.server = server
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
				go ts.server.ServeConn(conn)
			}
		}
	}()

	return nil
}
