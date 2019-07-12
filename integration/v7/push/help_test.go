package push

import (
	"regexp"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("help", func() {
	When("--help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF(PushCommandName, "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("%s - Push a new app or sync changes to an existing app", PushCommandName))
			Eventually(session).Should(Say("USAGE:"))

			buildpackAppUsage := []string{
				"cf",
				PushCommandName,
				"APP_NAME",
				"[-b BUILDPACK_NAME]",
				"[-c COMMAND]",
				"[-f MANIFEST_PATH | --no-manifest]",
				"[--no-start]",
				"[--no-wait]",
				"[-i NUM_INSTANCES]",
				"[-k DISK]",
				"[-m MEMORY]",
				"[-p PATH]",
				"[-s STACK]",
				"[-t HEALTH_TIMEOUT]",
				"[-u (process | port | http)]",
				"[--no-route | --random-route]",
				"[--var KEY=VALUE]",
				"[--vars-file VARS_FILE_PATH]...",
			}

			dockerAppUsage := []string{
				"cf",
				PushCommandName,
				"APP_NAME",
				"--docker-image",
				"[REGISTRY_HOST:PORT/]IMAGE[:TAG]",
				"[--docker-username USERNAME]",
				"[-c COMMAND]",
				"[-f MANIFEST_PATH | --no-manifest]",
				"[--no-start]",
				"[--no-wait]",
				"[-i NUM_INSTANCES]",
				"[-k DISK]",
				"[-m MEMORY]",
				"[-p PATH]",
				"[-s STACK]",
				"[-t HEALTH_TIMEOUT]",
				"[-u (process | port | http)]",
				"[--no-route | --random-route ]",
				"[--var KEY=VALUE]",
				"[--vars-file VARS_FILE_PATH]...",
			}

			assertUsage(session, buildpackAppUsage, dockerAppUsage)

			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say(`-b\s+Custom buildpack by name \(e\.g\. my-buildpack\) or Git URL \(e\.g\. 'https://github.com/cloudfoundry/java-buildpack.git'\) or Git URL with a branch or tag \(e\.g\. 'https://github.com/cloudfoundry/java-buildpack\.git#v3.3.0' for 'v3.3.0' tag\)\. To use built-in buildpacks only, specify 'default' or 'null'`))
			Eventually(session).Should(Say(`--docker-image, -o\s+Docker image to use \(e\.g\. user/docker-image-name\)`))
			Eventually(session).Should(Say(`--docker-username\s+Repository username; used with password from environment variable CF_DOCKER_PASSWORD`))
			Eventually(session).Should(Say(`--no-route\s+Do not map a route to this app`))
			Eventually(session).Should(Say(`--no-start\s+Do not stage and start the app after pushing`))
			Eventually(session).Should(Say(`-p\s+Path to app directory or to a zip file of the contents of the app directory`))
			Eventually(session).Should(Say("ENVIRONMENT:"))
			Eventually(session).Should(Say(`CF_DOCKER_PASSWORD=\s+Password used for private docker repository`))
			Eventually(session).Should(Say(`CF_STAGING_TIMEOUT=15\s+Max wait time for staging, in minutes`))
			Eventually(session).Should(Say(`CF_STARTUP_TIMEOUT=5\s+Max wait time for app instance startup, in minutes`))

			Eventually(session).Should(Exit(0))
		})
	})
})

func assertUsage(session *Session, usages ...[]string) {
	for _, usage := range usages {
		for k, v := range usage {
			usage[k] = regexp.QuoteMeta(v)
		}
		Eventually(session).Should(Say(
			strings.Join(usage, `\s+`)),
		)
	}
}
