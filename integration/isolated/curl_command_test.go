package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("error handling", func() {
	Describe("exit codes", func() {
		Context("when an unknown command is invoked", func() {
			It("exits 1", func() {
				session := helpers.CF("some-command-that-should-never-actually-be-a-real-thing-i-can-use")

				Eventually(session).Should(Exit(1))
				Eventually(session).Should(Say("not a registered command"))
			})
		})

		Context("when a known command is invoked with an invalid option", func() {
			It("exits 1", func() {
				session := helpers.CF("push", "--crazy")

				Eventually(session).Should(Exit(1))
			})
		})
	})

	Describe("incorrect usage", func() {
		Context("when a command is invoked with an invalid options", func() {
			It("does not display requirement errors twice", func() {
				session := helpers.CF("space")

				Eventually(session.Err).Should(Say("the required argument `SPACE` was not provided"))
				Consistently(session.Err).ShouldNot(Say("the required argument `SPACE` was not provided"))
				Consistently(session.Out).ShouldNot(Say("the required argument `SPACE` was not provided"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
