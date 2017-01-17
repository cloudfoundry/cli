package isolated

import (
	"fmt"
	"os/exec"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("version command", func() {
	DescribeTable("displays version",
		func(arg string) {
			session := helpers.CF(arg)
			Eventually(session).Should(Say("cf version [\\w0-9.+]+-[\\w0-9]+"))
			Eventually(session).Should(Exit(0))
		},

		Entry("when passed version", "version"),
		Entry("when passed -v", "-v"),
		Entry("when passed --version", "--version"),
	)

	DescribeTable("binary version substitution",
		func(version string, sha string, date string, expectedOutput string) {
			var ldFlags []string
			if version != "" {
				ldFlags = append(ldFlags,
					fmt.Sprintf("-X code.cloudfoundry.org/cli/version.binaryVersion=%s", version))
			}
			if sha != "" {
				ldFlags = append(ldFlags,
					fmt.Sprintf("-X code.cloudfoundry.org/cli/version.binarySHA=%s", sha))
			}
			if date != "" {
				ldFlags = append(ldFlags,
					fmt.Sprintf("-X code.cloudfoundry.org/cli/version.binaryBuildDate=%s", date))
			}

			path, err := Build("code.cloudfoundry.org/cli",
				"-ldflags",
				strings.Join(ldFlags, " "),
			)
			Expect(err).ToNot(HaveOccurred())

			session, err := Start(
				exec.Command(path, "version"),
				GinkgoWriter,
				GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session.Out).Should(Say(expectedOutput))
			Eventually(session).Should(Exit(0))

			CleanupBuildArtifacts()
		},

		Entry("when passed no ldflags", "", "", "", "cli(\\.exe)? version 0.0.0-unknown-version"),
		Entry("when passed just a build-sha", "", "deadbeef", "", "cli(\\.exe)? version 0.0.0-unknown-version\\+deadbeef"),
		Entry("when passed just a build-date", "", "", "2001-01-01", "cli(\\.exe)? version 0.0.0-unknown-version\\+2001-01-01"),
		Entry("when passed a sha and build-date", "", "deadbeef", "2001-01-01", "cli(\\.exe)? version 0.0.0-unknown-version\\+deadbeef.2001-01-01"),
		Entry("when passed just a version", "1.1.1", "", "", "cli(\\.exe)? version 1.1.1"),
		Entry("when passed a version and build-sha", "1.1.1", "deadbeef", "", "cli(\\.exe)? version 1.1.1\\+deadbeef"),
		Entry("when passed a version and a build-date", "1.1.1", "", "2001-01-01", "cli(\\.exe)? version 1.1.1\\+2001-01-01"),
		Entry("when passed a version, build-sha, and build-date", "1.1.1", "deadbeef", "2001-01-01", "cli(\\.exe)? version 1.1.1\\+deadbeef.2001-01-01"),
		Entry("when passed a wacky version", "#$%{@+&*!", "deadbeef", "2001-01-01", "cli(\\.exe)? version 0.0.0-unknown-version\\+deadbeef.2001-01-01"),
	)
})
