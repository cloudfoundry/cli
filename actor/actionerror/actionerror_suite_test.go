package actionerror_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestActionerror(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ActionError Suite")
}
