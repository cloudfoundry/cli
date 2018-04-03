package cmd

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"gopkg.in/yaml.v2"

	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
)

type VarsCertLoader struct {
	vars boshtpl.Variables
}

func NewVarsCertLoader(vars boshtpl.Variables) VarsCertLoader {
	return VarsCertLoader{vars}
}

func (l VarsCertLoader) LoadCerts(name string) (*x509.Certificate, *rsa.PrivateKey, error) {
	val, found, err := l.vars.Get(boshtpl.VariableDefinition{Name: name})
	if err != nil {
		return nil, nil, err
	} else if !found {
		return nil, nil, fmt.Errorf("Expected to find variable '%s' with a certificate", name)
	}

	// Convert to YAML for easier struct parsing
	valBytes, err := yaml.Marshal(val)
	if err != nil {
		return nil, nil, bosherr.WrapErrorf(err, "Expected variable '%s' to be serializable", name)
	}

	type CertVal struct {
		Certificate string
		PrivateKey  string `yaml:"private_key"`
	}

	var certVal CertVal

	err = yaml.Unmarshal(valBytes, &certVal)
	if err != nil {
		return nil, nil, bosherr.WrapErrorf(err, "Expected variable '%s' to be deserializable", name)
	}

	crt, err := l.parseCertificate(certVal.Certificate)
	if err != nil {
		return nil, nil, err
	}

	key, err := l.parsePrivateKey(certVal.PrivateKey)
	if err != nil {
		return nil, nil, err
	}

	return crt, key, nil
}

func (VarsCertLoader) parseCertificate(data string) (*x509.Certificate, error) {
	cpb, _ := pem.Decode([]byte(data))
	if cpb == nil {
		return nil, bosherr.Error("Certificate did not contain PEM formatted block")
	}

	crt, err := x509.ParseCertificate(cpb.Bytes)
	if err != nil {
		return nil, bosherr.WrapError(err, "Parsing certificate")
	}

	return crt, nil
}

func (VarsCertLoader) parsePrivateKey(data string) (*rsa.PrivateKey, error) {
	kpb, _ := pem.Decode([]byte(data))
	if kpb == nil {
		return nil, bosherr.Error("Private key did not contain PEM formatted block")
	}

	key, err := x509.ParsePKCS1PrivateKey(kpb.Bytes)
	if err != nil {
		return nil, bosherr.WrapError(err, "Parsing private key")
	}

	return key, nil
}
