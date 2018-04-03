package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"time"

	"github.com/pivotal-cf/paraphernalia/secure/tlsconfig"
)

func NewMutualTLSClient(identity tls.Certificate, caCertPool *x509.CertPool, serverName string) *http.Client {
	tlsConfig := tlsconfig.Build(
		tlsconfig.WithIdentity(identity),
		tlsconfig.WithInternalServiceDefaults(),
	)

	clientConfig := tlsConfig.Client(tlsconfig.WithAuthority(caCertPool))
	clientConfig.BuildNameToCertificate()
	clientConfig.ServerName = serverName

	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       clientConfig,
		},
		Timeout: 10 * time.Second,
	}
}
