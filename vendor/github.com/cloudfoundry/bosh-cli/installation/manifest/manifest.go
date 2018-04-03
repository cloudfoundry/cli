package manifest

import (
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type Manifest struct {
	Name       string
	Template   ReleaseJobRef
	Properties biproperty.Map
	Mbus       string
	Cert       Certificate
	Registry   Registry
}

type Certificate struct {
	CA string
}

type ReleaseJobRef struct {
	Name    string
	Release string
}

type Registry struct {
	Username  string
	Password  string
	Host      string
	Port      int
	SSHTunnel SSHTunnel
}

func (r Registry) IsEmpty() bool {
	return r == Registry{}
}

type SSHTunnel struct {
	User       string
	Host       string
	Port       int
	Password   string
	PrivateKey string `yaml:"private_key"`
}

func (m *Manifest) PopulateRegistry(username string, password string, host string, port int, sshTunnel SSHTunnel) {
	m.Properties["registry"] = biproperty.Map{
		"host":     host,
		"port":     port,
		"username": username,
		"password": password,
	}
	m.Registry = Registry{
		Username:  username,
		Password:  password,
		Host:      host,
		Port:      port,
		SSHTunnel: sshTunnel,
	}
}
