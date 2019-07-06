package isolated

import (
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"

	"github.com/blang/semver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("version command", func() {
	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
	})

	DescribeTable("displays version",
		func(arg string) {
			session := helpers.CF(arg)
			Eventually(session).Should(Exit(0))
			output := string(session.Out.Contents())
			version := strings.Split(output, " ")[2]
			versionNumber := strings.Split(version, "+")[0]
			_, err := semver.Make(versionNumber)
			Expect(err).To(Not(HaveOccurred()))
			Eventually(session).ShouldNot(Say("cf version 0.0.0-unknown-version"))
		},

		Entry("when passed version", "version"),
		Entry("when passed -v", "-v"),
		Entry("when passed --version", "--version"),
	)
})
