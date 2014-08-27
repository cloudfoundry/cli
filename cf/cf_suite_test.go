package cf_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cf Suite")
}
$ Cf v
cf phiên bản 6.2.0-c9d4aaa
