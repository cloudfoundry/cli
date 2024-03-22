package push

import (
	"regexp"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
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
				"[-l LOG_RATE_LIMIT]",
				"[-p PATH]",
				"[-s STACK]",
				"[-t HEALTH_TIMEOUT]",
				"[--task TASK]",
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
				"[-l LOG_RATE_LIMIT]",
				"[-p PATH]",
				"[-s STACK]",
				"[-t HEALTH_TIMEOUT]",
				"[--task TASK]",
				"[-u (process | port | http)]",
				"[--no-route | --random-route ]",
				"[--var KEY=VALUE]",
				"[--vars-file VARS_FILE_PATH]...",
			}

			assertUsage(session, buildpackAppUsage, dockerAppUsage)

			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say(`--app-start-timeout, -t`))
			Eventually(session).Should(Say(`--buildpack, -b`))
			Eventually(session).Should(Say(`--disk, -k`))
			Eventually(session).Should(Say(`--docker-image, -o`))
			Eventually(session).Should(Say(`--docker-username`))
			Eventually(session).Should(Say(`--droplet`))
			Eventually(session).Should(Say(`--endpoint`))
			Eventually(session).Should(Say(`--health-check-type, -u`))
			Eventually(session).Should(Say(`--instances, -i`))
			Eventually(session).Should(Say(`--log-rate-limit, -l\s+Log rate limit per second, in bytes \(e.g. 128B, 4K, 1M\). -l=-1 represents unlimited`))
			Eventually(session).Should(Say(`--manifest, -f`))
			Eventually(session).Should(Say(`--memory, -m`))
			Eventually(session).Should(Say(`--no-manifest`))
			Eventually(session).Should(Say(`--no-route`))
			Eventually(session).Should(Say(`--no-start`))
			Eventually(session).Should(Say(`--no-wait`))
			Eventually(session).Should(Say(`--path, -p`))
			Eventually(session).Should(Say(`--random-route`))
			Eventually(session).Should(Say(`--stack, -s`))
			Eventually(session).Should(Say(`--start-command, -c`))
			Eventually(session).Should(Say(`--strategy`))
			Eventually(session).Should(Say(`--task`))
			Eventually(session).Should(Say(`--var`))
			Eventually(session).Should(Say(`--vars-file`))
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
