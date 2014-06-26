package running_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRunning(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Running Suite")
}
