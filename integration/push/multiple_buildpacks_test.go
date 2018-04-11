package push

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with different buildpack values", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("when the buildpack flag is provided", func() {
		It("sets that buildpack correctly for the pushed app", func() {
			helpers.WithProcfileApp(func(dir string) {
				tempfile := filepath.Join(dir, "index.html")
				err := ioutil.WriteFile(tempfile, []byte(fmt.Sprintf("hello world %d", rand.Int())), 0666)
				Expect(err).ToNot(HaveOccurred())

				By("pushing a ruby app with a static buildpack sets buildpack to static")
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"-b", "staticfile_buildpack",
				)
				Eventually(session).Should(Say("Staticfile Buildpack"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("app", appName)
				Eventually(session).Should(Say("buildpack:\\s+staticfile_buildpack"))
				Eventually(session).Should(Exit(0))

				By("pushing a ruby app with a null buildpack sets buildpack to auto-detected (ruby)")
				session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"-b", "null",
				)
				Eventually(session).Should(Say(`\-\s+buildpack:\s+staticfile_buildpack`))
				Consistently(session).ShouldNot(Say(`\+\s+buildpack:`))
				Eventually(session).Should(Say("Ruby Buildpack"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("app", appName)
				Eventually(session).Should(Say(`buildpack:\s+ruby`))
				Eventually(session).Should(Exit(0))

				By("pushing a ruby app with a static buildpack sets buildpack to static")
				session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"-b", "staticfile_buildpack",
				)
				Eventually(session).Should(Say("Staticfile Buildpack"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("app", appName)
				Eventually(session).Should(Say("buildpack:\\s+staticfile_buildpack"))
				Eventually(session).Should(Exit(0))

				By("pushing a ruby app with a default buildpack sets buildpack to auto-detected (ruby)")
				session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"-b", "default",
				)
				Eventually(session).Should(Say(`\-\s+buildpack:\s+staticfile_buildpack`))
				Consistently(session).ShouldNot(Say(`\+\s+buildpack:`))
				Eventually(session).Should(Say("Ruby Buildpack"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("app", appName)
				Eventually(session).Should(Say(`buildpack:\s+ruby`))
				Eventually(session).Should(Exit(0))

			})
		})
	})
})
