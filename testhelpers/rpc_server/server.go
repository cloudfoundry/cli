package test_rpc_server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/cli/plugin/models"
)

type Handlers interface {
	IsMinCliVersion(args string, retVal *bool) error
	SetPluginMetadata(pluginMetadata plugin.PluginMetadata, retVal *bool)

	CliCommandWithoutTerminalOutput(args ...string) ([]string, error)
	CliCommand(args ...string) ([]string, error)
	GetCurrentOrg() (plugin_models.Organization, error)
	GetCurrentSpace() (plugin_models.Space, error)
	Username() (string, error)
	UserGuid() (string, error)
	UserEmail() (string, error)
	IsLoggedIn() (bool, error)
	IsSSLDisabled() (bool, error)
	HasOrganization() (bool, error)
	HasSpace() (bool, error)
	ApiEndpoint() (string, error)
	ApiVersion() (string, error)
	HasAPIEndpoint() (bool, error)
	LoggregatorEndpoint() (string, error)
	DopplerEndpoint() (string, error)
	AccessToken() (string, error)
	GetApp(string) (plugin_models.GetAppModel, error)
	GetApps() ([]plugin_models.GetAppsModel, error)
	GetOrgs() ([]plugin_models.GetOrgs_Model, error)
	GetSpaces() ([]plugin_models.GetSpaces_Model, error)
	GetOrgUsers(string, ...string) ([]plugin_models.GetOrgUsers_Model, error)
	GetSpaceUsers(string, string) ([]plugin_models.GetSpaceUsers_Model, error)
	GetServices() ([]plugin_models.GetServices_Model, error)
	GetService(string) (plugin_models.GetService_Model, error)
	GetOrg(string) (plugin_models.GetOrg_Model, error)
	GetSpace(string) (plugin_models.GetSpace_Model, error)
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
