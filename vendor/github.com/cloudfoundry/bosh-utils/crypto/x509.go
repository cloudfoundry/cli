package crypto

import (
	"crypto/x509"
	"encoding/pem"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

func CertPoolFromPEM(pemCerts []byte) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()

	for pemCertsIdx := 1; len(pemCerts) > 0; pemCertsIdx++ {
		var block *pem.Block

		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			if strings.TrimSpace(string(pemCerts)) != "" {
				return nil, bosherr.Errorf("Parsing certificate %d: Missing PEM block", pemCertsIdx)
			}

			break
		}

		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			return nil, bosherr.Errorf("Parsing certificate %d: Not a certificate", pemCertsIdx)
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Parsing certificate %d", pemCertsIdx)
		}

		certPool.AddCert(cert)
	}

	return certPool, nil
}
