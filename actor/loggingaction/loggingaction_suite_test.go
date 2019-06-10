package loggingaction_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLoggingaction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Loggingaction Suite")
}
