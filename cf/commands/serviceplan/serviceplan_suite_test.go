package serviceplan_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestServicePlan(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Plan Suite")
}
