package options_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestOptions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Options Suite")
}
