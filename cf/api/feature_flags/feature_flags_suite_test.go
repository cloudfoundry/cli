package feature_flags_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFeatureFlags(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FeatureFlags Suite")
}
