package scp_test

import (
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestScp(t *testing.T) {
	RegisterFailHandler(Fail)
	BeforeEach(func() {
		if runtime.GOOS == "windows" {
			Skip("scp isn't supported on windows")
		}
	})
	RunSpecs(t, "Scp Suite")
}
