package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with --task", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		helpers.WithHelloWorldApp(func(dir string) {
			session := helpers.CF("push", appName, "-p", dir, "--task")
			Eventually(session).Should(Exit(0))
		})
	})

	When("pushing an app with the --task flag", func() {
		var session *gexec.Session

		BeforeEach(func() {
			session = helpers.CF("app", appName)
			Eventually(session).Should(Exit(0))
		})

		It("pushes the app without starting it", func() {
			Expect(session).To(Say(`name:\s+%s`, appName))
			Expect(session).To(Say(`requested state:\s+stopped`))
		})
		It("pushes the app with no routes", func() {
			Expect(session).To(Say(`name:\s+%s`, appName))
			Expect(session).To(Say(`routes:\s+last uploaded`))
		})
		It("pushes the app with no instances", func() {
			Expect(session).To(Say(`name:\s+%s`, appName))
			Expect(session).To(Say(`type:\s+web\s+sidecars:\s+instances:\s+0/1`))
		})
	})
})
