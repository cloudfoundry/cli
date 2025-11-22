package models_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestModelHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Model Helpers Suite")
}
