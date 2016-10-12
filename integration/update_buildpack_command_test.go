package integration

import (
	. "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unbind-service command", func() {
	Context("when the buildpack is not provided", func() {
		It("returns a buildpack argument not provided error", func() {
			session := CF("update-buildpack", "-p", ".")
			Eventually(session).Should(Exit(1))

			Expect(session.Err.Contents()).To(BeEquivalentTo("Incorrect Usage: the required argument `BUILDPACK` was not provided\n\n"))
		})
	})
})
