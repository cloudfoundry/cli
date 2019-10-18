package global

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-staging-environment-variable-group command", func() {
	var (
		key1 string
		key2 string
		val1 string
		val2 string
	)

	BeforeEach(func() {
		helpers.LoginCF()

		key1 = helpers.PrefixedRandomName("key1")
		key2 = helpers.PrefixedRandomName("key2")
		val1 = helpers.PrefixedRandomName("val1")
		val2 = helpers.PrefixedRandomName("val2")
	})

	AfterEach(func() {
		session := helpers.CF("set-staging-environment-variable-group", "{}")
		Eventually(session).Should(Exit(0))
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("set-staging-environment-variable-group", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("set-staging-environment-variable-group - Pass parameters as JSON to create a staging environment variable group"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf set-staging-environment-variable-group '{"name":"value","name":"value"}'`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("set-env, staging-environment-variable-group"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	It("sets staging environment variables", func() {
		json := fmt.Sprintf(`{"%s":"%s", "%s":"%s"}`, key1, val1, key2, val2)
		session := helpers.CF("set-staging-environment-variable-group", json)
		Eventually(session).Should(Say("Setting the contents of the staging environment variable group as"))
		Eventually(session).Should(Say("OK"))
		Eventually(session).Should(Exit(0))

		session = helpers.CF("staging-environment-variable-group")
		Eventually(session).Should(Exit(0))
		// We cannot use `Say()`, for the results are returned in a random order
		Expect(string(session.Out.Contents())).To(MatchRegexp(`%s\s+%s`, key1, val1))
		Expect(string(session.Out.Contents())).To(MatchRegexp(`%s\s+%s`, key2, val2))
	})

	When("user passes in '{}'", func() {
		BeforeEach(func() {
			json := fmt.Sprintf(`{"%s":"%s", "%s":"%s"}`, key1, val1, key2, val2)
			session := helpers.CF("set-staging-environment-variable-group", json)
			Eventually(session).Should(Exit(0))
		})

		It("clears the environment group", func() {
			json := fmt.Sprintf(`{}`)
			session := helpers.CF("set-staging-environment-variable-group", json)
			Eventually(session).Should(Say("Setting the contents of the staging environment variable group as"))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))

			session = helpers.CF("staging-environment-variable-group")
			Eventually(session).Should(Exit(0))
			Expect(string(session.Out.Contents())).ToNot(MatchRegexp(fmt.Sprintf(`%s\s+%s`, key1, val1)))
			Expect(string(session.Out.Contents())).ToNot(MatchRegexp(fmt.Sprintf(`%s\s+%s`, key2, val2)))
		})
	})

	When("user unsets some, but not all. variables", func() {
		BeforeEach(func() {
			json := fmt.Sprintf(`{"%s":"%s", "%s":"%s"}`, key1, val1, key2, val2)
			session := helpers.CF("set-staging-environment-variable-group", json)
			Eventually(session).Should(Exit(0))
		})

		It("clears the removed variables", func() {
			json := fmt.Sprintf(`{"%s":"%s"}`, key1, val1)
			session := helpers.CF("set-staging-environment-variable-group", json)
			Eventually(session).Should(Say("Setting the contents of the staging environment variable group as"))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))

			session = helpers.CF("staging-environment-variable-group")
			Eventually(session).Should(Exit(0))
			Expect(string(session.Out.Contents())).To(MatchRegexp(fmt.Sprintf(`%s\s+%s`, key1, val1)))
			Expect(string(session.Out.Contents())).ToNot(MatchRegexp(fmt.Sprintf(`%s\s+%s`, key2, val2)))
		})
	})

	When("user passes invalid JSON", func() {
		It("fails helpfully", func() {
			session := helpers.CF("set-staging-environment-variable-group", `not json...`)
			Eventually(session).Should(Say("Setting the contents of the staging environment variable group as"))
			Eventually(session.Err).Should(Say("Invalid environment variable group provided. Please provide a valid JSON object."))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Exit(1))
		})
	})
})
