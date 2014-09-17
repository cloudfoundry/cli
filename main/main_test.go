package main_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Integration")
}

var _ = Describe("exit codes", func() {
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
