package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("help command", func() {

	Describe("commands that appear in cf help -a", func() {
		It("includes labels", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Say(`labels\s+List all labels \(key-value pairs\) for an API resource`))
			Eventually(session).Should(Exit(0))
		})
		It("includes set-label", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Say(`set-label\s+Set a label \(key-value pairs\) for an API resource`))
			Eventually(session).Should(Exit(0))
		})
		It("includes delete-label", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Say(`delete-label\s+Delete a label \(key-value pairs\) for an API resource`))
			Eventually(session).Should(Exit(0))
		})
	})

})
