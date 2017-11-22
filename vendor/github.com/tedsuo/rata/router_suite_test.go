package rata_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRouter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rata Suite")
}
