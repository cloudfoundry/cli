package uaa

import (
	"crypto/x509"
	gonet "net"
	gourl "net/url"
	"strconv"
	"strings"

	"github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Config struct {
	Host string
	Port int
	Path string

	Client       string
	ClientSecret string

	CACert string
}

func NewConfigFromURL(url string) (Config, error) {
	if len(url) == 0 {
		return Config{}, bosherr.Error("Expected non-empty UAA URL")
	}

	parsedURL, err := gourl.Parse(url)
	if err != nil {
		return Config{}, bosherr.WrapErrorf(err, "Parsing UAA URL '%s'", url)
	}

	host := parsedURL.Host
	port := 443
	path := parsedURL.Path

	if len(host) == 0 {
		host = url
		path = ""
	}

	if strings.Contains(host, ":") {
		var portStr string

		host, portStr, err = gonet.SplitHostPort(host)
		if err != nil {
			return Config{}, bosherr.WrapErrorf(
				err, "Extracting host/port from URL '%s'", url)
		}

		port, err = strconv.Atoi(portStr)
		if err != nil {
			return Config{}, bosherr.WrapErrorf(
				err, "Extracting port from URL '%s'", url)
		}
	}

	if len(host) == 0 {
		return Config{}, bosherr.Errorf("Expected to extract host from URL '%s'", url)
	}

	return Config{Host: host, Port: port, Path: path}, nil
}

func (c Config) Validate() error {
	if len(c.Host) == 0 {
		return bosherr.Error("Missing 'Host'")
	}

	if c.Port == 0 {
		return bosherr.Error("Missing 'Port'")
	}

	if len(c.Client) == 0 {
		return bosherr.Error("Missing 'Client'")
	}

	if _, err := c.CACertPool(); err != nil {
		return err
	}

	return nil
}

func (c Config) CACertPool() (*x509.CertPool, error) {
	if len(c.CACert) == 0 {
		return nil, nil
	}

	return crypto.CertPoolFromPEM([]byte(c.CACert))
}
