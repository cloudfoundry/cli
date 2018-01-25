package helpers

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func EnableFeatureFlag(flagName string) {
	session := CF("enable-feature-flag", flagName)
	Eventually(session).Should(Exit(0))
}

func DisableFeatureFlag(flagName string) {
	session := CF("disable-feature-flag", flagName)
	Eventually(session).Should(Exit(0))
}
