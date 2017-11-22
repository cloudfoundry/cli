package daemon

import (
	"net"

	"code.cloudfoundry.org/diego-ssh/handlers"
	"code.cloudfoundry.org/lager"
	"golang.org/x/crypto/ssh"
)

type Daemon struct {
	logger                lager.Logger
	serverConfig          *ssh.ServerConfig
	globalRequestHandlers map[string]handlers.GlobalRequestHandler
	newChannelHandlers    map[string]handlers.NewChannelHandler
}

func New(
	logger lager.Logger,
	serverConfig *ssh.ServerConfig,
	globalRequestHandlers map[string]handlers.GlobalRequestHandler,
	newChannelHandlers map[string]handlers.NewChannelHandler,
) *Daemon {
	return &Daemon{
		logger:                logger,
		serverConfig:          serverConfig,
		globalRequestHandlers: globalRequestHandlers,
		newChannelHandlers:    newChannelHandlers,
	}
}

func (d *Daemon) HandleConnection(netConn net.Conn) {
	logger := d.logger.Session("handle-connection")

	logger.Info("started")
	defer logger.Info("completed")
	defer netConn.Close()

	serverConn, serverChannels, serverRequests, err := ssh.NewServerConn(netConn, d.serverConfig)
	if err != nil {
		logger.Error("handshake-failed", err)
		return
	}

	go d.handleGlobalRequests(logger, serverRequests)
	go d.handleNewChannels(logger, serverChannels)

	serverConn.Wait()
}

func (d *Daemon) handleGlobalRequests(logger lager.Logger, requests <-chan *ssh.Request) {
	logger = logger.Session("handle-global-requests")
	logger.Info("starting")
	defer logger.Info("finished")

	for req := range requests {
		logger.Debug("request", lager.Data{
			"request-type": req.Type,
			"want-reply":   req.WantReply,
		})

		handler, ok := d.globalRequestHandlers[req.Type]
		if ok {
			handler.HandleRequest(logger, req)
			continue
		}

		if req.WantReply {
			req.Reply(false, nil)
		}
	}
}

func (d *Daemon) handleNewChannels(logger lager.Logger, newChannelRequests <-chan ssh.NewChannel) {
	logger = logger.Session("handle-new-channels")
	logger.Info("starting")
	defer logger.Info("finished")

	for newChannel := range newChannelRequests {
		logger.Info("new-channel", lager.Data{
			"channelType": newChannel.ChannelType(),
			"extraData":   newChannel.ExtraData(),
		})

		if handler, ok := d.newChannelHandlers[newChannel.ChannelType()]; ok {
			go handler.HandleNewChannel(logger, newChannel)
			continue
		}

		newChannel.Reject(ssh.UnknownChannelType, newChannel.ChannelType())
	}
}
