package global

import (
	"fmt"
	"math/rand"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("env command", func() {
	var (
		appName string
		orgName string

		key1 string
		key2 string
		key3 string
		key4 string
		key5 string
		key6 string

		val1 string
		val2 int
		val3 string
		val4 int
		val5 string
		val6 string
	)

	BeforeEach(func() {
		spaceName := helpers.NewSpaceName()
		orgName = helpers.NewOrgName()
		setupCF(orgName, spaceName)

		appName = helpers.PrefixedRandomName("app")

		key1 = helpers.PrefixedRandomName("key1")
		key2 = helpers.PrefixedRandomName("key2")
		val1 = helpers.PrefixedRandomName("val1")
		val2 = rand.Intn(2000)
		json := fmt.Sprintf(`{"%s":"%s", "%s":%d}`, key1, val1, key2, val2)
		session := helpers.CF("set-staging-environment-variable-group", json)
		Eventually(session).Should(Exit(0))

		key3 = helpers.PrefixedRandomName("key3")
		key4 = helpers.PrefixedRandomName("key4")
		val3 = helpers.PrefixedRandomName("val3")
		val4 = rand.Intn(2000)
		json = fmt.Sprintf(`{"%s":"%s", "%s":%d}`, key3, val3, key4, val4)
		session = helpers.CF("set-running-environment-variable-group", json)
		Eventually(session).Should(Exit(0))

		helpers.WithHelloWorldApp(func(appDir string) {
			Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
		})

		key5 = helpers.PrefixedRandomName("key5")
		key6 = helpers.PrefixedRandomName("key6")
		val5 = helpers.PrefixedRandomName("val5")
		val6 = fmt.Sprint(rand.Intn(2000))
		session = helpers.CF("set-env", appName, key5, val5)
		Eventually(session).Should(Exit(0))
		session = helpers.CF("set-env", appName, key6, val6)
		Eventually(session).Should(Exit(0))
	})

	AfterEach(func() {
		session := helpers.CF("set-staging-environment-variable-group", "{}")
		Eventually(session).Should(Exit(0))
		session = helpers.CF("set-running-environment-variable-group", "{}")
		Eventually(session).Should(Exit(0))

		helpers.QuickDeleteOrg(orgName)
	})

	It("displays all environment variables", func() {
		session := helpers.CF("env", appName)

		Eventually(session).Should(Say("System-Provided:"))
		Eventually(session).Should(Say("VCAP_APPLICATION"))

		Eventually(session).Should(Say("User-Provided:"))
		Eventually(session).Should(Say("%s: %s", key5, val5))
		Eventually(session).Should(Say("%s: %s", key6, val6))

		Eventually(session).Should(Say("Running Environment Variable Groups:"))
		Eventually(session).Should(Say("%s: %s", key3, val3))
		Eventually(session).Should(Say("%s: %d", key4, val4))

		Eventually(session).Should(Say("Staging Environment Variable Groups:"))
		Eventually(session).Should(Say("%s: %s", key1, val1))
		Eventually(session).Should(Say("%s: %d", key2, val2))

		Eventually(session).Should(Exit(0))
	})
})
