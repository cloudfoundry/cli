package i18n_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestI18n(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "I18n Suite")
}
