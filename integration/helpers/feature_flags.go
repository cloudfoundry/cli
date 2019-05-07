package helpers

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

// EnableFeatureFlag enables given feature flag with 'cf enable-feature-flag'.
func EnableFeatureFlag(flagName string) {
	Eventually(CF("enable-feature-flag", flagName)).Should(Exit(0))
}

// DisableFeatureFlag disables given feature flag with 'cf disable-feature-flag'.
func DisableFeatureFlag(flagName string) {
	Eventually(CF("disable-feature-flag", flagName)).Should(Exit(0))
}
