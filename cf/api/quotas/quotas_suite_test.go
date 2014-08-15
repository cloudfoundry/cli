package quotas_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestQuotas(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Quotas Suite")
}
