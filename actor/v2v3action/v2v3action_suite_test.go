package v2v3action_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestV2v3action(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "v2v3action Suite")
}
