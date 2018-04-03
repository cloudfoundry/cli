package registry

import (
	"fmt"
	"net"
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type ServerManager interface {
	Start(string, string, string, int) (Server, error)
}

type serverManager struct {
	logger boshlog.Logger
	logTag string
}

func NewServerManager(logger boshlog.Logger) ServerManager {
	return &serverManager{
		logger: logger,
		logTag: "registryServer",
	}
}

// Create starts a new server on a goroutine and returns it
// The returned error is only for starting. Error while running is logged.
func (s *serverManager) Start(username string, password string, host string, port int) (Server, error) {
	startedCh := make(chan error)
	server := &server{
		logger: s.logger,
		logTag: "registryServer",
	}
	go func() {
		err := server.start(username, password, host, port, startedCh)
		if err != nil {
			s.logger.Debug(s.logTag, "Registry error occurred: %s", err.Error())
		}
	}()

	// block until started
	err := <-startedCh
	if err != nil {
		if stopErr := server.Stop(); stopErr != nil {
			s.logger.Warn(s.logTag, "Failed to stop server: %s", stopErr.Error())
		}
	}
	return server, err
}

type Server interface {
	Stop() error
}

type server struct {
	listener net.Listener
	logger   boshlog.Logger
	logTag   string
}

func NewServer(logger boshlog.Logger) Server {
	return &server{
		logger: logger,
		logTag: "registryServer",
	}
}

func (s *server) start(username string, password string, host string, port int, readyErrCh chan error) error {
	s.logger.Debug(s.logTag, "Starting registry server at %s:%d", host, port)
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		readyErrCh <- bosherr.WrapError(err, "Starting registry listener")
		return nil
	}

	readyErrCh <- nil

	httpServer := http.Server{}
	mux := http.NewServeMux()
	httpServer.Handler = mux

	registry := NewRegistry()
	instanceHandler := newInstanceHandler(username, password, registry, s.logger)
	mux.HandleFunc("/instances/", instanceHandler.HandleFunc)

	return httpServer.Serve(s.listener)
}

func (s *server) Stop() error {
	if s.listener == nil {
		return bosherr.Error("Stopping not-started registry server")
	}

	s.logger.Debug(s.logTag, "Stopping registry server")
	err := s.listener.Close()
	if err != nil {
		return bosherr.WrapError(err, "Stopping registry server")
	}

	return nil
}
