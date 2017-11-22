package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/cfhttp"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/consuladapter"
	"code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/diego-ssh/authenticators"
	"code.cloudfoundry.org/diego-ssh/cmd/ssh-proxy/config"
	"code.cloudfoundry.org/diego-ssh/healthcheck"
	"code.cloudfoundry.org/diego-ssh/proxy"
	"code.cloudfoundry.org/diego-ssh/server"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagerflags"
	"code.cloudfoundry.org/locket"
	"github.com/cloudfoundry/dropsonde"
	"github.com/hashicorp/consul/api"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
	"golang.org/x/crypto/ssh"
)

const (
	dropsondeOrigin = "ssh-proxy"
)

var configPath = flag.String(
	"config",
	"",
	"Path to SSH Proxy config.",
)

func main() {
	debugserver.AddFlags(flag.CommandLine)
	flag.Parse()

	sshProxyConfig, err := config.NewSSHProxyConfig(*configPath)
	if err != nil {
		logger, _ := lagerflags.New("ssh-proxy")
		logger.Fatal("failed-to-parse-config", err)
	}

	logger, reconfigurableSink := lagerflags.NewFromConfig("ssh-proxy", sshProxyConfig.LagerConfig)

	cfhttp.Initialize(time.Duration(sshProxyConfig.CommunicationTimeout))

	initializeDropsonde(logger, sshProxyConfig.DropsondePort)

	proxySSHServerConfig, err := configureProxy(logger, sshProxyConfig)
	if err != nil {
		logger.Error("configure-failed", err)
		os.Exit(1)
	}

	sshProxy := proxy.New(logger, proxySSHServerConfig)
	server := server.NewServer(logger, sshProxyConfig.Address, sshProxy)

	healthCheckHandler := healthcheck.NewHandler(logger)
	httpServer := http_server.New(sshProxyConfig.HealthCheckAddress, healthCheckHandler)

	consulClient, err := consuladapter.NewClientFromUrl(sshProxyConfig.ConsulCluster)
	if err != nil {
		logger.Fatal("new-client-failed", err)
	}

	registrationRunner := initializeRegistrationRunner(logger, consulClient, sshProxyConfig.Address, clock.NewClock())

	members := grouper.Members{
		{"ssh-proxy", server},
		{"registration-runner", registrationRunner},
		{"healthcheck", httpServer},
	}

	if sshProxyConfig.DebugAddress != "" {
		members = append(grouper.Members{{
			"debug-server", debugserver.Runner(sshProxyConfig.DebugAddress, reconfigurableSink),
		}}, members...)
	}

	group := grouper.NewOrdered(os.Interrupt, members)
	monitor := ifrit.Invoke(sigmon.New(group))

	logger.Info("started")

	err = <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}

	logger.Info("exited")
	os.Exit(0)
}

func configureProxy(logger lager.Logger, sshProxyConfig config.SSHProxyConfig) (*ssh.ServerConfig, error) {
	if sshProxyConfig.BBSAddress == "" {
		err := errors.New("bbsAddress is required")
		logger.Fatal("bbs-address-required", err)
	}

	url, err := url.Parse(sshProxyConfig.BBSAddress)
	if err != nil {
		logger.Fatal("failed-to-parse-bbs-address", err)
	}

	bbsClient := initializeBBSClient(
		logger,
		sshProxyConfig.BBSAddress,
		sshProxyConfig.BBSCACert,
		sshProxyConfig.BBSClientCert,
		sshProxyConfig.BBSClientKey,
		sshProxyConfig.BBSClientSessionCacheSize,
		sshProxyConfig.BBSMaxIdleConnsPerHost,
	)
	permissionsBuilder := authenticators.NewPermissionsBuilder(bbsClient)

	authens := []authenticators.PasswordAuthenticator{}

	if sshProxyConfig.EnableDiegoAuth {
		diegoAuthenticator := authenticators.NewDiegoProxyAuthenticator(logger, []byte(sshProxyConfig.DiegoCredentials), permissionsBuilder)
		authens = append(authens, diegoAuthenticator)
	}

	if sshProxyConfig.EnableCFAuth {
		if sshProxyConfig.CCAPIURL == "" {
			return nil, errors.New("ccAPIURL is required for Cloud Foundry authentication")
		}

		_, err = url.Parse(sshProxyConfig.CCAPIURL)
		if err != nil {
			return nil, err
		}

		if sshProxyConfig.UAAPassword == "" {
			return nil, errors.New("UAA password is required for Cloud Foundry authentication")
		}

		if sshProxyConfig.UAAUsername == "" {
			return nil, errors.New("UAA username is required for Cloud Foundry authentication")
		}

		if sshProxyConfig.UAATokenURL == "" {
			return nil, errors.New("uaaTokenURL is required for Cloud Foundry authentication")
		}

		_, err = url.Parse(sshProxyConfig.UAATokenURL)
		if err != nil {
			return nil, err
		}

		client := newHttpClient(sshProxyConfig.SkipCertVerify, time.Duration(sshProxyConfig.CommunicationTimeout))
		cfAuthenticator := authenticators.NewCFAuthenticator(
			logger,
			client,
			sshProxyConfig.CCAPIURL,
			sshProxyConfig.UAATokenURL,
			sshProxyConfig.UAAUsername,
			sshProxyConfig.UAAPassword,
			permissionsBuilder,
		)
		authens = append(authens, cfAuthenticator)
	}

	authenticator := authenticators.NewCompositeAuthenticator(authens...)

	sshConfig := &ssh.ServerConfig{
		PasswordCallback: authenticator.Authenticate,
		AuthLogCallback: func(cmd ssh.ConnMetadata, method string, err error) {
			if err != nil {
				logger.Error("authentication-failed", err, lager.Data{"user": cmd.User()})
			} else {
				logger.Info("authentication-attempted", lager.Data{"user": cmd.User()})
			}
		},
	}

	if sshProxyConfig.HostKey == "" {
		err := errors.New("hostKey is required")
		logger.Fatal("host-key-required", err)
	}

	key, err := parsePrivateKey(logger, sshProxyConfig.HostKey)
	if err != nil {
		logger.Fatal("failed-to-parse-host-key", err)
	}

	sshConfig.AddHostKey(key)

	if sshProxyConfig.AllowedCiphers != "" {
		sshConfig.Config.Ciphers = strings.Split(sshProxyConfig.AllowedCiphers, ",")
	}
	if sshProxyConfig.AllowedMACs != "" {
		sshConfig.Config.MACs = strings.Split(sshProxyConfig.AllowedMACs, ",")
	}
	if sshProxyConfig.AllowedKeyExchanges != "" {
		sshConfig.Config.KeyExchanges = strings.Split(sshProxyConfig.AllowedKeyExchanges, ",")
	}

	return sshConfig, err
}

func initializeDropsonde(logger lager.Logger, dropsondePort int) {
	dropsondeDestination := fmt.Sprint("localhost:", dropsondePort)
	err := dropsonde.Initialize(dropsondeDestination, dropsondeOrigin)
	if err != nil {
		logger.Error("failed to initialize dropsonde: %v", err)
	}
}

func parsePrivateKey(logger lager.Logger, encodedKey string) (ssh.Signer, error) {
	key, err := ssh.ParsePrivateKey([]byte(encodedKey))
	if err != nil {
		logger.Error("failed-to-parse-private-key", err)
		return nil, err
	}
	return key, nil
}

func newHttpClient(insecureSkipVerify bool, communicationTimeout time.Duration) *http.Client {
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	tlsConfig := &tls.Config{InsecureSkipVerify: insecureSkipVerify}
	return &http.Client{
		Transport: &http.Transport{
			Dial:            dialer.Dial,
			TLSClientConfig: tlsConfig,
		},
		Timeout: communicationTimeout,
	}
}

func initializeBBSClient(
	logger lager.Logger,
	bbsAddress,
	bbsCACert,
	bbsClientCert,
	bbsClientKey string,
	bbsClientSessionCacheSize,
	bbsMaxIdleConnsPerHost int,
) bbs.InternalClient {
	bbsURL, err := url.Parse(bbsAddress)
	if err != nil {
		logger.Fatal("Invalid BBS URL", err)
	}

	if bbsURL.Scheme != "https" {
		return bbs.NewClient(bbsAddress)
	}

	bbsClient, err := bbs.NewSecureClient(bbsAddress, bbsCACert, bbsClientCert, bbsClientKey, bbsClientSessionCacheSize, bbsMaxIdleConnsPerHost)
	if err != nil {
		logger.Fatal("Failed to configure secure BBS client", err)
	}
	return bbsClient
}

func initializeRegistrationRunner(logger lager.Logger, consulClient consuladapter.Client, listenAddress string, clock clock.Clock) ifrit.Runner {
	_, portString, err := net.SplitHostPort(listenAddress)
	if err != nil {
		logger.Fatal("failed-invalid-listen-address", err)
	}
	portNum, err := net.LookupPort("tcp", portString)
	if err != nil {
		logger.Fatal("failed-invalid-listen-port", err)
	}

	registration := &api.AgentServiceRegistration{
		Name: "ssh-proxy",
		Port: portNum,
		Check: &api.AgentServiceCheck{
			TTL: "20s",
		},
	}

	return locket.NewRegistrationRunner(logger, registration, consulClient, locket.RetryInterval, clock)
}
