package helpers

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

// SecurityGroup represents a security group API resource
type SecurityGroup struct {
	Name        string `json:"-"`
	Protocol    string `json:"protocol"`
	Destination string `json:"destination"`
	Ports       string `json:"ports"`
	Description string `json:"description"`
}

// NewSecurityGroup returns a new security group with the given attributes
func NewSecurityGroup(name string, protocol string, destination string, ports string, description string) SecurityGroup {
	return SecurityGroup{
		Name:        name,
		Protocol:    protocol,
		Destination: destination,
		Ports:       ports,
		Description: description,
	}
}

// Create Creates a new security group on the API using the 'cf create-security-group'
func (s SecurityGroup) Create() {
	dir, err := ioutil.TempDir("", "simple-security-group")
	Expect(err).ToNot(HaveOccurred())
	defer os.RemoveAll(dir)

	tempfile := filepath.Join(dir, "security-group.json")

	securityGroup, err := json.Marshal([]SecurityGroup{s})
	Expect(err).ToNot(HaveOccurred())

	err = ioutil.WriteFile(tempfile, securityGroup, 0666)
	Expect(err).ToNot(HaveOccurred())
	Eventually(CF("create-security-group", s.Name, tempfile)).Should(Exit(0))
}
