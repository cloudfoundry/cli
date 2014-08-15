package space_quotas_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSpaceQuotas(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SpaceQuotas Suite")
}
