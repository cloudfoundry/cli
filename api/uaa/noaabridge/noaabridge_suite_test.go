package noaabridge_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestNOAABridge(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NOAA Bridge Suite")
}
