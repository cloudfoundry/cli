package spellcheck_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSpellcheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spellcheck Suite")
}
