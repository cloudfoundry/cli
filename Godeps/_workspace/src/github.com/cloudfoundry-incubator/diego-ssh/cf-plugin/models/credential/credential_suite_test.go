package credential_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCredential(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Credential Suite")
}
