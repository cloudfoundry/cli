package organizations_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestOrganizations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Organizations Suite")
}
