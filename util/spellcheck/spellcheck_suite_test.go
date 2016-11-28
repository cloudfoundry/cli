package spellcheck_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSpellcheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spellcheck Suite")
}
