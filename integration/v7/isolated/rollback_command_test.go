package isolated

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/types"
)

var _ = Describe("rollback command", func() {

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("rollback", "EXPERIMENTAL COMMANDS", "Rollback to the specified revision of an app"))
			})

			It("Displays rollback command usage to output", func() {
				session := helpers.CF("rollback", "--help")

				Eventually(session).Should(Exit(0))

				Expect(session).To(Say("NAME:"))
				Expect(session).To(Say("rollback - Rollback to the specified revision of an app"))
				Expect(session).To(Say("USAGE:"))
				Expect(session).To(Say(`cf rollback APP_NAME \[--version VERSION\]`))
				Expect(session).To(Say("OPTIONS:"))
				Expect(session).To(Say("-f             Force rollback without confirmation"))
				Expect(session).To(Say("--version      Roll back to the specified revision"))
				Expect(session).To(Say("SEE ALSO:"))
				Expect(session).To(Say("revisions"))
			})
		})
	})

	When("the environment is set up correctly", func() {
		var (
			appName   string
			orgName   string
			spaceName string
			userName  string
		)

		BeforeEach(func() {
			appName = helpers.PrefixedRandomName("app")
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			userName, _ = helpers.GetCredentials()
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Describe("the app does not exist", func() {
			It("errors with app not found", func() {
				session := helpers.CF("rollback", appName, "--version", "1")
				Eventually(session).Should(Exit(1))

				Expect(session).ToNot(Say("Are you sure you want to continue?"))

				Expect(session.Err).To(Say("App '%s' not found.", appName))
				Expect(session).To(Say("FAILED"))
			})
		})

		Describe("the app exists with revisions", func() {
			When("the app is started and has 2 instances", func() {
				var domainName string

				BeforeEach(func() {
					domainName = helpers.DefaultSharedDomain()
					helpers.WithHelloWorldApp(func(appDir string) {
						manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  memory: 128M
  instances: 2
  disk_quota: 128M
  routes:
  - route: %s.%s
`, appName, appName, domainName))
						manifestPath := filepath.Join(appDir, "manifest.yml")
						err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
						Expect(err).ToNot(HaveOccurred())

						Eventually(helpers.CF("push", appName, "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")).Should(Exit(0))
						Eventually(helpers.CF("push", appName, "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")).Should(Exit(0))
					})
				})

				When("the desired revision does not exist", func() {
					It("errors with 'revision not found'", func() {
						session := helpers.CF("rollback", appName, "--version", "5")
						Eventually(session).Should(Exit(1))

						Expect(session.Err).To(Say("Revision '5' not found"))
						Expect(session).To(Say("FAILED"))
					})
				})

				When("the -f flag is provided", func() {
					It("does not prompt the user, and just rolls back", func() {
						session := helpers.CF("rollback", appName, "--version", "1", "-f")
						Eventually(session).Should(Exit(0))

						Expect(session).To(HaveRollbackOutput(appName, orgName, spaceName, userName))

						session = helpers.CF("revisions", appName)
						Eventually(session).Should(Exit(0))
					})
				})

				Describe("the -f flag is not provided", func() {
					var buffer *Buffer

					BeforeEach(func() {
						buffer = NewBuffer()
					})

					When("the user enters y", func() {
						BeforeEach(func() {
							_, err := buffer.Write([]byte("y\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("prompts the user to rollback, then successfully rolls back", func() {
							session := helpers.CFWithStdin(buffer, "rollback", appName, "--version", "1")
							Eventually(session).Should(Exit(0))

							Expect(session).To(HaveRollbackPrompt())
							Expect(session).To(HaveRollbackOutput(appName, orgName, spaceName, userName))
							Expect(session).To(Say("OK"))

							session = helpers.CF("revisions", appName)
							Eventually(session).Should(Exit(0))

							Expect(session).To(Say(`3\(deployed\)\s+New droplet deployed.`))
						})
					})

					When("the user enters n", func() {
						BeforeEach(func() {
							_, err := buffer.Write([]byte("n\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("prompts the user to rollback, then does not rollback", func() {
							session := helpers.CFWithStdin(buffer, "rollback", appName, "--version", "1")
							Eventually(session).Should(Exit(0))

							Expect(session).To(HaveRollbackPrompt())
							Expect(session).To(Say("App '%s' has not been rolled back to revision '1'", appName))

							session = helpers.CF("revisions", appName)
							Eventually(session).Should(Exit(0))

							Expect(session).ToNot(Say(`3\s+[\w\-]+\s+New droplet deployed.`))
						})
					})
				})
			})
		})
	})
})

type Line struct {
	str  string
	args []interface{}
}

func HaveRollbackPrompt() *CLIMatcher {
	return &CLIMatcher{Lines: []Line{
		{"Are you sure you want to continue?", nil},
	}}
}

func HaveRollbackOutput(appName, orgName, spaceName, userName string) *CLIMatcher {
	return &CLIMatcher{Lines: []Line{
		AppInOrgSpaceAsUser(appName, orgName, spaceName, userName),
		KeyValue("name", appName),
		KeyValue("routes", fmt.Sprintf("%s.%s", appName, helpers.DefaultSharedDomain())),
	}}
}

// Per-style guide: https://github.com/cloudfoundry/cli/wiki/CF-CLI-Style-Guide#system-feedback--transparency
func AppInOrgSpaceAsUser(appName, orgName, spaceName, userName string) Line {
	return Line{`app %s in org %s / space %s as %s`, []interface{}{appName, orgName, spaceName, userName}}
}

// Per-style guide: https://github.com/cloudfoundry/cli/wiki/CF-CLI-Style-Guide#keyvalue-pairs
func KeyValue(key, value string) Line {
	return Line{`%s:\s+%s`, []interface{}{key, value}}
}

type CLIMatcher struct {
	Lines          []Line
	wrappedMatcher types.GomegaMatcher
}

func (cm *CLIMatcher) Match(actual interface{}) (bool, error) {
	for _, line := range cm.Lines {
		cm.wrappedMatcher = types.GomegaMatcher(Say(line.str, line.args...))
		success, err := cm.wrappedMatcher.Match(actual)
		if err != nil {
			return false, err
		}

		if success != true {
			return false, nil
		}
	}

	return true, nil
}

func (cm CLIMatcher) FailureMessage(interface{}) string {
	return cm.wrappedMatcher.FailureMessage(nil)
}

func (cm CLIMatcher) NegatedFailureMessage(interface{}) string {
	return cm.wrappedMatcher.NegatedFailureMessage(nil)
}
