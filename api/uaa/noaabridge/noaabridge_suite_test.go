package noaabridge_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestNOAABridge(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NOAA Bridge Suite")
}
