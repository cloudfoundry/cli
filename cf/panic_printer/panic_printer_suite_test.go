package panic_printer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPanicHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PanicHandler Suite")
}
