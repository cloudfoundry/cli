package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Push with app directory", func() {
	Context("when the specified app directory does not exist", func() {
		It("displays a path does not exist error, help, and exits 1", func() {
			session := helpers.CF("push", "-f", "./non-existant-dir/")
			Eventually(session.Err).Should(Say("Incorrect Usage: The specified path './non-existant-dir/' does not exist."))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session.Out).Should(Say("USAGE:"))
			Eventually(session).Should(Exit(1))
		})
	})
})
