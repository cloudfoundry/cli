package trace_test

import (
	"code.cloudfoundry.org/cli/v9/util/trace"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GenerateHex", func() {
	It("returns random trace id", func() {
		Expect(trace.GenerateUUIDTraceID()).To(HaveLen(32))
	})
})

var _ = Describe("GenerateRandomTraceID", func() {
	It("returns random trace id", func() {
		Expect(trace.GenerateRandomTraceID(16)).To(HaveLen(16))
	})
})
