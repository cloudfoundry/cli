package helpers

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func EnableFeatureFlag(flagName string) {
	tempHome := SetHomeDir()
	SetAPI()
	LoginCF()
	Eventually(CF("enable-feature-flag", flagName)).Should(Exit(0))
	DestroyHomeDir(tempHome)
}

func DisableFeatureFlag(flagName string) {
	tempHome := SetHomeDir()
	SetAPI()
	LoginCF()
	session := CF("disable-feature-flag", flagName)
	Eventually(session).Should(Exit(0))
	DestroyHomeDir(tempHome)
}
