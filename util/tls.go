package util

import (
	"crypto/tls"
	"crypto/x509"
)

func NewTLSConfig(trustedCerts []*x509.Certificate, skipTLSValidation bool) *tls.Config {
	config := &tls.Config{
		MinVersion: tls.VersionTLS10,
		MaxVersion: tls.VersionTLS12,
	}

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
