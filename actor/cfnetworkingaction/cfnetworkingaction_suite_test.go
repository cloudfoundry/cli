package cfnetworkingaction_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCfnetworkingaction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CF Networking Action Suite")
}
