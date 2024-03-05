package json_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestJSON(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "JSON Suite")
}
