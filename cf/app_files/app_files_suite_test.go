package app_files_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestAppFiles(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "App Files Suite")
}
