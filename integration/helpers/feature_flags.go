package helpers

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func EnableFeatureFlag(flagName string) {
	Eventually(CF("enable-feature-flag", flagName)).Should(Exit(0))
}

func DisableFeatureFlag(flagName string) {
	Eventually(CF("disable-feature-flag", flagName)).Should(Exit(0))
}
