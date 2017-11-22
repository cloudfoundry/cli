package signals_test

import (
	"code.cloudfoundry.org/diego-ssh/signals"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Signals", func() {
	Describe("Signal Mapping", func() {
		It("should have the same length map", func() {
			Expect(signals.SyscallSignals).To(HaveLen(len(signals.SSHSignals)))
		})

		It("has the correct mapping", func() {
			for k, v := range signals.SyscallSignals {
				Expect(k).To(Equal(signals.SSHSignals[v]))
			}

			for k, v := range signals.SSHSignals {
				Expect(k).To(Equal(signals.SyscallSignals[v]))
			}
		})
	})
})
