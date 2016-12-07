package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("space command", func() {
	BeforeEach(func() {
		setupCF(ReadOnlyOrg, ReadOnlySpace)
	})

	It("displays a space", func() {
		session := helpers.CF("space", ReadOnlySpace)
		Eventually(session).Should(Say("Org:\\s+%s", ReadOnlyOrg))
		Eventually(session).Should(Exit(0))
	})
})
