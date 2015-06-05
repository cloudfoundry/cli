package ui_helpers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUiHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UiHelpers Suite")
}
