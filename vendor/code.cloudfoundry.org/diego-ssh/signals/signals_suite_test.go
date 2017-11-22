package signals_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSignals(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Signals Suite")
}
