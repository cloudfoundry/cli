package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/resources"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

// SecurityGroup represents a security group API resource
type SecurityGroup struct {
	Name  string `json:"-"`
	Rules []resources.Rule
}

// NewSecurityGroup returns a new security group with the given attributes
func NewSecurityGroup(name string, protocol string, destination string, ports *string, description *string) SecurityGroup {
	return SecurityGroup{
		Name: name,
		Rules: []resources.Rule{{
			Protocol:    protocol,
			Destination: destination,
			Ports:       ports,
			Description: description,
		}},
	}
}

// Create Creates a new security group on the API using the 'cf create-security-group'
func (s SecurityGroup) Create() {
	dir, err := ioutil.TempDir("", "simple-security-group")
	Expect(err).ToNot(HaveOccurred())
	defer os.RemoveAll(dir)

	tempfile := filepath.Join(dir, "security-group.json")

	securityGroup, err := json.Marshal(s.Rules)
	Expect(err).ToNot(HaveOccurred())

	err = ioutil.WriteFile(tempfile, securityGroup, 0666)
	Expect(err).ToNot(HaveOccurred())
	Eventually(CF("create-security-group", s.Name, tempfile)).Should(Exit(0))
}

// Delete Deletes a security group on the API using the 'cf delete-security-group'
func (s SecurityGroup) Delete() {
	if s.Name == "" {
		fmt.Println("Empty security group name. Skipping deletion.")
		return
	}
	Eventually(CF("delete-security-group", s.Name, "-f")).Should(Exit(0))
}
