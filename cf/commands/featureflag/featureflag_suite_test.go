package featureflag_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFeatureflag(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FeatureFlag Suite")
}
