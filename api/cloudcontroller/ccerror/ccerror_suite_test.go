package ccerror_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCcerror(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cloud Controller Errors Suite")
}
