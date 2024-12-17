package random_test

import (
	"code.cloudfoundry.org/cli/util/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GenerateHex", func() {
	It("returns random hex of correct length", func() {
		Expect(random.GenerateHex(16)).To(HaveLen(16))
	})
})
