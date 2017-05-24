package push

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with hostname", func() {
	Context("when the default domain is a shared domain", func() {
		DescribeTable("creates and binds the route as neccessary",
			func(existingRoute bool, boundRoute bool, setup func(appName string, dir string) *Session) {
				appName := helpers.NewAppName()

				if existingRoute {
					session := helpers.CF("create-route", space, defaultSharedDomain(), "-n", appName)
					Eventually(session).Should(Exit(0))
				}

				if boundRoute {
					helpers.WithHelloWorldApp(func(dir string) {
						// TODO: Add --no-start
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
						Eventually(session).Should(Exit(0))
					})
				}

				helpers.WithHelloWorldApp(func(dir string) {
					session := setup(appName, dir)

					Eventually(session).Should(Say("routes:"))
					if existingRoute && boundRoute {
						Eventually(session).Should(Say("(?i)%s.%s", appName, defaultSharedDomain()))
					} else {
						Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", appName, defaultSharedDomain()))
					}
					Eventually(session).Should(Say("Mapping routes..."))

					Eventually(session).Should(Exit(0))
				})

				resp, err := http.Get(fmt.Sprintf("http://%s.%s", appName, defaultSharedDomain()))
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			},

			Entry("when the hostname is provided via the appName and route does not exist", false, false, func(appName string, dir string) *Session {
				// TODO: Add --no-start
				// TODO: Add --path
				return helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
			}),

			Entry("when the hostname is provided via the appName and the unbound route exists", true, false, func(appName string, dir string) *Session {
				// TODO: Add --no-start
				// TODO: Add --path
				return helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
			}),

			Entry("when the hostname is provided via the appName and the bound route exists", true, true, func(appName string, dir string) *Session {
				// TODO: Add --no-start
				// TODO: Add --path
				return helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
			}),
		)
	})
})
