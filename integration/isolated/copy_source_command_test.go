package isolated

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("copy-source command", func() {
	var (
		appName1  string
		appName2  string
		orgName   string
		spaceName string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()

		setupCF(orgName, spaceName)

		appName1 = helpers.PrefixedRandomName("hello")
		appName2 = helpers.PrefixedRandomName("banana")

		helpers.WithHelloWorldApp(func(appDir string) {
			Eventually(helpers.CF("push", appName1, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
		})

		helpers.WithBananaPantsApp(func(appDir string) {
			Eventually(helpers.CF("push", appName2, "--no-start", "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
		})
	})

	AfterEach(func() {
		helpers.QuickDeleteOrg(orgName)
	})

	It("copies the app", func() {
		session := helpers.CF("copy-source", appName1, appName2)
		Eventually(session).Should(Say("Copying source from app %s to target app %s", appName1, appName2))
		Eventually(session).Should(Say("Showing health and status for app %s", appName2))
		Eventually(session).Should(Exit(0))

		resp, err := http.Get(fmt.Sprintf("http://%s.%s", appName2, defaultSharedDomain()))
		Expect(err).ToNot(HaveOccurred())
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(body)).To(MatchRegexp("hello world"))
	})
})
