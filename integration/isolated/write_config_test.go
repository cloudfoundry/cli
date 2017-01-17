package isolated

import (
	"path/filepath"
	"runtime"
	"syscall"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("writing the config after command execution", func() {
	BeforeEach(func() {
		if runtime.GOOS == "windows" {
			Skip("no way to make the config readable but not writable on Windows")
		}
	})

	Context("when an error occurs writing to the config", func() {
		BeforeEach(func() {
			homeDir := helpers.SetHomeDir()

			// This command call is purely to create the cf config for the first time.
			session := helpers.CF("api", helpers.GetAPI(), "--skip-ssl-validation")
			Eventually(session).Should(Exit(0))

			configFilepath := filepath.Join(homeDir, ".cf", "config.json")

			// Make the config file immutable. Chmod is not enough because root can
			// write a read-only file.
			err := syscall.Chflags(configFilepath, 2) // UF_IMMUTABLE
			Expect(err).ToNot(HaveOccurred())
		})

		It("displays the error to stderr", func() {
			session := helpers.CF("api")
			Eventually(session.Err).Should(Say("Error writing config: open .+/.cf/config.json: operation not permitted"))
			Eventually(session).Should(Exit(0))
		})
	})
})
