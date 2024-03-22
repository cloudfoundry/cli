package terminal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTerminal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Terminal Suite")
}
