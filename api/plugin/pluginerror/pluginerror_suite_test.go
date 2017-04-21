package pluginerror_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCcerror(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Plugin Errors Suite")
}
