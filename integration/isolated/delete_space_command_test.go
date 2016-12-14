package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-space command", func() {
	Describe("help", func() {
		It("shows usage", func() {
			session := helpers.CF("help", "delete-space")
			Eventually(session).Should(Exit(0))
			Expect(session).To(Say("delete-space SPACE \\[-o ORG\\] \\[-f\\]"))
		})
	})
})
