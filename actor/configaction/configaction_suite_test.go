package configaction_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConfigAction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Actions Suite")
}
