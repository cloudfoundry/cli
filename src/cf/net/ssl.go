package net

import (
	"cf/errors"
	"crypto/tls"
	"crypto/x509"
	"net/url"
)

func TLSConfigWithTrustedCerts(trustedCerts []tls.Certificate) *tls.Config {
	if len(trustedCerts) > 0 {
		certPool := x509.NewCertPool()
		for _, tlsCert := range trustedCerts {
			cert, _ := x509.ParseCertificate(tlsCert.Certificate[0])
			certPool.AddCert(cert)
		}
		return &tls.Config{RootCAs: certPool}
	} else {
		return &tls.Config{}
	}
}

func wrapSSLErrors(host string, err error) errors.Error {
	urlError, ok := err.(*url.Error)
	if ok {
		switch urlError.Err.(type) {
		case x509.UnknownAuthorityError:
			return errors.NewInvalidSSLCert(host, "unknown authority")
		case x509.HostnameError:
			return errors.NewInvalidSSLCert(host, "not valid for the requested host")
		case x509.CertificateInvalidError:
			return errors.NewInvalidSSLCert(host, "")
		}
	}
	return errors.NewErrorWithError("Error performing request", err)
}
