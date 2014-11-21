package detection_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDetection(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Detection Suite")
}
