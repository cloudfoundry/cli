package serviceaccess_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestServiceAccess(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Access Suite")
}
