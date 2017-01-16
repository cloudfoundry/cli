package isolated

import (
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("writing the config after command execution", func() {
	Context("when an error occurs writing to the config", func() {
		BeforeEach(func() {
			homeDir := helpers.SetHomeDir()

			// This command call is purely to create the cf config for the first time.
			session := helpers.CF("api", helpers.GetAPI(), "--skip-ssl-validation")
			Eventually(session).Should(Exit(0))

			configFilepath := filepath.Join(homeDir, ".cf", "config.json")

			// Make the config read-only so reading it succeeds but writing fails.
			err := os.Chmod(configFilepath, 0400)
			Expect(err).ToNot(HaveOccurred())
		})

		It("displays the error to stderr", func() {
			session := helpers.CF("api")
			Eventually(session.Err).Should(Say("Error writing config: open .+/.cf/config.json: permission denied"))
			Eventually(session).Should(Exit(0))
		})
	})
})
