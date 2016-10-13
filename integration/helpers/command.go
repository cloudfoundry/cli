package helpers

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func CF(args ...string) *Session {
	session, err := Start(
		exec.Command("cf", args...),
		NewPrefixedWriter("OUT: ", GinkgoWriter),
		NewPrefixedWriter("ERR: ", GinkgoWriter))
	Expect(err).NotTo(HaveOccurred())
	return session
}
