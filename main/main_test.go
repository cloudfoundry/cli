package main_test

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Integration")
}

var _ = Describe("main", func() {
	Describe("exit codes", func() {
		It("exits non-zero when an unknown command is invoked", func() {
			result := Cf("some-command-that-should-never-actually-be-a-real-thing-i-can-use")

			Eventually(result).Should(Say("not a registered command"))
			Eventually(result).Should(Exit(1))
		})

		It("exits non-zero when known command is invoked with invalid option", func() {
			result := Cf("push", "--crazy")
			Eventually(result).Should(Exit(1))
		})
	})

	Describe("Plugins", func() {
		var (
			old_CF_HOME string
		)

		BeforeEach(func() {
			old_CF_HOME = os.Getenv("CF_HOME")

			dir, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			cmd := exec.Command("go", "build", "-a", "-o", path.Join(dir, "..", "fixtures", "plugins", "test"), path.Join(dir, "..", "fixtures", "plugins", "test.go"))
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			fullDir := filepath.Join(dir, "..", "fixtures", "config", "plugin-config")
			err = os.Setenv("CF_HOME", fullDir)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.Setenv("CF_HOME", old_CF_HOME)
			Expect(err).NotTo(HaveOccurred())
		})

		It("can print help for all core commands by executing only the command `cf`", func() {
			output := Cf()
			Eventually(output.Out.Contents).Should(ContainSubstring("A command line tool to interact with Cloud Foundry"))
		})

		It("informs user for any invalid commands", func() {
			output := Cf("foo-bar")
			Eventually(output.Out).Should(Say("'foo-bar' is not a registered command"))
		})

		Context("plugins", func() {
			It("Can call a executable from the Plugins configuration if it does not exist as a cf command", func() {
				output := Cf("valid-plugin").Wait(5 * time.Second)
				Eventually(output.Out).Should(Say("HaHaHaHa you called the push plugin"))
			})

			It("Calls core cf command if the plugin shares the same name", func() {
				output := Cf("help")
				Eventually(output.Out).ShouldNot(Say("HaHaHaHa you called the push plugin"))
			})
		})
	})
})

func Cf(args ...string) *Session {
	path, err := Build("github.com/cloudfoundry/cli/main")
	Expect(err).NotTo(HaveOccurred())

	session, err := Start(exec.Command(path, args...), GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}

// gexec.Build leaves a compiled binary behind in /tmp.
var _ = AfterSuite(func() {
	CleanupBuildArtifacts()
})
