package versioncheck_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVersioncheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Versioncheck Suite")
}
