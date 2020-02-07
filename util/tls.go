package util

import (
	"crypto/tls"
	"crypto/x509"

	"code.cloudfoundry.org/tlsconfig"
)

func NewTLSConfig(trustedCerts []*x509.Certificate, skipTLSValidation bool) *tls.Config {
	config := &tls.Config{}

	_ = tlsconfig.WithExternalServiceDefaults()(config) //nolint - always returns nil

	if len(trustedCerts) > 0 {
		certPool := x509.NewCertPool()
		for _, tlsCert := range trustedCerts {
			certPool.AddCert(tlsCert)
		}
		config.RootCAs = certPool
	}

	config.InsecureSkipVerify = skipTLSValidation

	return config
}
