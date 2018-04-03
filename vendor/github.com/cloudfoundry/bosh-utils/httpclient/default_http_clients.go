package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"time"

	proxy "github.com/cloudfoundry/socks5-proxy"
)

var (
	DefaultClient = CreateDefaultClientInsecureSkipVerify()
	defaultDialer = SOCKS5DialFuncFromEnvironment((&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).Dial, proxy.NewSocks5Proxy(proxy.NewHostKeyGetter()))
)

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

func CreateDefaultClient(certPool *x509.CertPool) *http.Client {
	insecureSkipVerify := false
	return factory{}.New(insecureSkipVerify, certPool)
}

func CreateDefaultClientInsecureSkipVerify() *http.Client {
	insecureSkipVerify := true
	return factory{}.New(insecureSkipVerify, nil)
}

type factory struct{}

func (f factory) New(insecureSkipVerify bool, certPool *x509.CertPool) *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
			TLSClientConfig: &tls.Config{
				RootCAs:            certPool,
				InsecureSkipVerify: insecureSkipVerify,
			},

			Proxy: http.ProxyFromEnvironment,
			Dial:  defaultDialer,

			TLSHandshakeTimeout: 30 * time.Second,
			DisableKeepAlives:   true,
		},
	}

	return client
}
