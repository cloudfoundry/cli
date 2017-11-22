package authenticators_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAuthenticators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Authenticators Suite")
}
