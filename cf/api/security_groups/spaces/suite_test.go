package spaces_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSecurityGroupSpaces(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SecurityGroupSpaces Suite")
}
