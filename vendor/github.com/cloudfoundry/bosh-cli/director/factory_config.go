package director

import (
	"crypto/x509"
	gonet "net"
	gourl "net/url"
	"strconv"
	"strings"

	"github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type FactoryConfig struct {
	Host string
	Port int

	// CA certificate is not required
	CACert string

	Client       string
	ClientSecret string

	TokenFunc func(bool) (string, error)
}

func NewConfigFromURL(url string) (FactoryConfig, error) {
	if len(url) == 0 {
		return FactoryConfig{}, bosherr.Error("Expected non-empty Director URL")
	}

	parsedURL, err := gourl.Parse(url)
	if err != nil {
		return FactoryConfig{}, bosherr.WrapErrorf(err, "Parsing Director URL '%s'", url)
	}

	host := parsedURL.Host
	port := 25555

	if len(host) == 0 {
		host = url
	}

	if strings.Contains(host, ":") {
		var portStr string

		host, portStr, err = gonet.SplitHostPort(host)
		if err != nil {
			return FactoryConfig{}, bosherr.WrapErrorf(
				err, "Extracting host/port from URL '%s'", parsedURL.Host)
		}

		port, err = strconv.Atoi(portStr)
		if err != nil {
			return FactoryConfig{}, bosherr.WrapErrorf(
				err, "Extracting port from URL '%s'", parsedURL.Host)
		}
	}

	if len(host) == 0 {
		return FactoryConfig{}, bosherr.Errorf("Expected to extract host from URL '%s'", url)
	}

	return FactoryConfig{Host: host, Port: port}, nil
}

func (c FactoryConfig) Validate() error {
	if len(c.Host) == 0 {
		return bosherr.Error("Missing 'Host'")
	}

	if c.Port == 0 {
		return bosherr.Error("Missing 'Port'")
	}

	if _, err := c.CACertPool(); err != nil {
		return err
	}

	// Don't validate credentials since Info call does not require authentication.

	return nil
}

func (c FactoryConfig) CACertPool() (*x509.CertPool, error) {
	if len(c.CACert) == 0 {
		return nil, nil
	}

	return crypto.CertPoolFromPEM([]byte(c.CACert))
}
