package global

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("running-environment-variable-group command", func() {
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

		json := fmt.Sprintf(`{"%s":"%s", "%s":"%s"}`, key1, val1, key2, val2)
		session := helpers.CF("set-running-environment-variable-group", json)
		Eventually(session).Should(Exit(0))
	})

	AfterEach(func() {
		session := helpers.CF("set-running-environment-variable-group", "{}")
		Eventually(session).Should(Exit(0))
	})

	It("gets running environment variables", func() {
		session := helpers.CF("running-environment-variable-group")
		Eventually(session).Should(Exit(0))
		stdout := string(session.Out.Contents())
		Expect(stdout).To(MatchRegexp(`%s\s+%s`, key1, val1))
		Expect(stdout).To(MatchRegexp(`%s\s+%s`, key2, val2))
	})
})
