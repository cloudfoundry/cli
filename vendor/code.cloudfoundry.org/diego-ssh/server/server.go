package server

import (
	"errors"
	"net"
	"os"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter -o fakes/fake_connection_handler.go . ConnectionHandler
type ConnectionHandler interface {
	HandleConnection(net.Conn)
}

type Server struct {
	logger        lager.Logger
	listenAddress string

	connectionHandler ConnectionHandler

	listener net.Listener
	mutex    *sync.Mutex
	stopping bool

	connections          map[net.Conn]struct{}
	connectionsMutex     *sync.Mutex
	connectionsWaitGroup *sync.WaitGroup
}

func NewServer(
	logger lager.Logger,
	listenAddress string,
	connectionHandler ConnectionHandler,
) *Server {
	return &Server{
		logger:               logger,
		listenAddress:        listenAddress,
		connectionHandler:    connectionHandler,
		mutex:                &sync.Mutex{},
		connections:          make(map[net.Conn]struct{}),
		connectionsMutex:     &sync.Mutex{},
		connectionsWaitGroup: &sync.WaitGroup{},
	}
}

func (s *Server) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	listener, err := net.Listen("tcp", s.listenAddress)
	if err != nil {
		return err
	}

	s.SetListener(listener)
	go s.Serve()

	close(ready)

	select {
	case <-signals:
		s.Shutdown()
	}

	return nil
}

func (s *Server) Shutdown() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.stopping {
		s.logger.Info("stopping-server")
		s.stopping = true
		s.listener.Close()

		s.connectionsMutex.Lock()
		for conn, _ := range s.connections {
			conn.Close()
		}
		s.connectionsMutex.Unlock()
	}

	s.connectionsWaitGroup.Wait()
}

func (s *Server) IsStopping() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.stopping
}

func (s *Server) SetListener(listener net.Listener) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.listener != nil {
		err := errors.New("Listener has already been set")
		s.logger.Error("listener-already-set", err)
		return err
	}

	s.listener = listener
	return nil
}

func (s *Server) ListenAddr() (net.Addr, error) {
	if s.listener == nil {
		return nil, errors.New("No listener")
	}

	return s.listener.Addr(), nil
}

func (s *Server) Serve() {
	logger := s.logger.Session("serve")
	defer s.listener.Close()

	for {
		netConn, err := s.listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				logger.Error("accept-temporary-error", netErr)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if s.IsStopping() {
				break
			}

			logger.Error("accept-failed", err)
			return
		}

		s.connectionsMutex.Lock()
		s.connections[netConn] = struct{}{}
		s.connectionsWaitGroup.Add(1)
		s.connectionsMutex.Unlock()

		go func() {
			s.connectionHandler.HandleConnection(netConn)

			s.connectionsMutex.Lock()
			delete(s.connections, netConn)
			s.connectionsWaitGroup.Done()
			s.connectionsMutex.Unlock()
		}()
	}
}
