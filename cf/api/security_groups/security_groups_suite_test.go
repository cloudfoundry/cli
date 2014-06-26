package security_groups_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSecurityGroups(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SecurityGroups Suite")
}
