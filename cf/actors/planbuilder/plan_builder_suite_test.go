package planbuilder_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPlanBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PlanBuilder Suite")
}
