package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push flag combination errors", func() {
	DescribeTable("path and",
		func(expectedError string, flags ...string) {
			appName := helpers.NewAppName()

			args := append([]string{PushCommandName, appName, "--no-start", "-p", realDir}, flags...)
			session := helpers.CF(args...)
			Eventually(session.Err).Should(Say(expectedError))
			Eventually(session).Should(Exit(1))
		},
		Entry("docker image", "Incorrect Usage: The following arguments cannot be used together: --docker-image, -o, -p", "-o", "some-image"),
	)

	DescribeTable("everything else",
		func(expectedError string, flags ...string) {
			appName := helpers.NewAppName()

			args := append([]string{PushCommandName, appName, "--no-start"}, flags...)
			session := helpers.CF(args...)
			Eventually(session.Err).Should(Say(expectedError))
			Eventually(session).Should(Exit(1))
		},
		Entry("no-route and domain", "The following arguments cannot be used together: -d, --no-route", "--no-route", "-d", "some-domain"),
		Entry("no-route and no-hostname", "The following arguments cannot be used together: --no-hostname, --no-route", "--no-route", "--no-hostname"),
		Entry("no-route and hostname", "The following arguments cannot be used together: --hostname, -n, --no-route", "--no-route", "--hostname", "some-hostname"),
		Entry("hostname and no-hostname", "The following arguments cannot be used together: --hostname, -n, --no-hostname", "--hostname", "some-hostname", "--no-hostname"),
		Entry("random-route and hostname", "The following arguments cannot be used together: --hostname, -n, --random-route", "--hostname", "some-hostname", "--random-route"),
		Entry("random-route and no-hostname", "The following arguments cannot be used together: --no-hostname, --random-route", "--no-hostname", "--random-route"),
		Entry("random-route and no-route", "The following arguments cannot be used together: --no-route, --random-route", "--no-route", "--random-route"),
		Entry("random-route and route path", "The following arguments cannot be used together: --random-route, --route-path", "--random-route", "--route-path", "some-route-path"),
		Entry("docker-username without image", "Incorrect Usage: '--docker-image, -o' and '--docker-username' must be used together.", "--docker-username", "some-user"),
		Entry("docker-image and buildpack", "Incorrect Usage: The following arguments cannot be used together: -b, --docker-image, -o", "-o", "some-image", "-b", "some-buidpack"),
	)
})
