package pluginaction_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPluginaction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Plugin Action Suite")
}
