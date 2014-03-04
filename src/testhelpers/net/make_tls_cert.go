package net

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

func MakeSelfSignedTLSCert() tls.Certificate {
	return generateCert([]string{"127.0.0.1", "::1"}, time.Date(2020, time.December, 1, 0, 0, 0, 0, time.UTC), true)
}

func MakeTLSCertWithInvalidHost() tls.Certificate {
	return generateCert([]string{"example.com"}, time.Date(2020, time.December, 1, 0, 0, 0, 0, time.UTC), true)
}

func MakeExpiredTLSCert() tls.Certificate {
	return generateCert([]string{"127.0.0.1", "::1"}, time.Date(2000, time.December, 1, 0, 0, 0, 0, time.UTC), true)
}

func MakeUnauthorizedTLSCert() tls.Certificate {
	return generateCert([]string{"127.0.0.1", "::1"}, time.Date(2020, time.December, 1, 0, 0, 0, 0, time.UTC), false)
}

func generateCert(hosts []string, notAfter time.Time, isAuthorizedToSign bool) tls.Certificate {
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}

	template := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(0),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC),
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, host := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, host)
		}
	}

	if isAuthorizedToSign {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		panic(err)
	}

	certOut := new(bytes.Buffer)
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyOut := new(bytes.Buffer)
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	cert, err := tls.X509KeyPair(certOut.Bytes(), keyOut.Bytes())
	if err != nil {
		panic(err)
	}

	return cert
}
