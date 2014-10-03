package fileutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFileutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fileutils Suite")
}
