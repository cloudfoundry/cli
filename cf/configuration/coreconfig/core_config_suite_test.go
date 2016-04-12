package coreconfig_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCoreConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CoreConfig Suite")
}
