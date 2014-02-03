package organization_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestOrganization(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Organization Suite")
}
