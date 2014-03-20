package net

import (
	"cf/errors"
	"code.google.com/p/go.net/websocket"
	"crypto/tls"
	"crypto/x509"
	"net/url"
)

func NewTLSConfig(trustedCerts []tls.Certificate, disableSSL bool) (TLSConfig *tls.Config) {
	TLSConfig = &tls.Config{}

	if len(trustedCerts) > 0 {
		certPool := x509.NewCertPool()
		for _, tlsCert := range trustedCerts {
			cert, _ := x509.ParseCertificate(tlsCert.Certificate[0])
			certPool.AddCert(cert)
		}
		TLSConfig.RootCAs = certPool
	}

	TLSConfig.InsecureSkipVerify = disableSSL

	return
}

func WrapSSLErrors(host string, err error) error {
	urlError, ok := err.(*url.Error)
	if ok {
		return wrapSSLErrorInternal(host, urlError.Err)
	}

	websocketError, ok := err.(*websocket.DialError)
	if ok {
		return wrapSSLErrorInternal(host, websocketError.Err)
	}

	return errors.NewErrorWithError("Error performing request", err)
}

func wrapSSLErrorInternal(host string, err error) error {
	switch err.(type) {
	case x509.UnknownAuthorityError:
		return errors.NewInvalidSSLCert(host, "unknown authority")
	case x509.HostnameError:
		return errors.NewInvalidSSLCert(host, "not valid for the requested host")
	case x509.CertificateInvalidError:
		return errors.NewInvalidSSLCert(host, "")
	default:
		return errors.NewErrorWithError("Error performing request", err)
	}
}
