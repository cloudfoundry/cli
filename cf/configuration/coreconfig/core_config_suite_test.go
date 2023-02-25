package coreconfig_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCoreConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CoreConfig Suite")
}
