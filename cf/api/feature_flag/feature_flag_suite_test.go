package feature_flag_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFeatureFlag(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FeatureFlag Suite")
}
