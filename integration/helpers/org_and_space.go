package helpers

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func SetupReadOnlyOrgAndSpace() (string, string) {
	homeDir := SetHomeDir()
	SetAPI()
	LoginCF()
	orgName := NewOrgName()
	spaceName1 := PrefixedRandomName("SPACE")
	spaceName2 := PrefixedRandomName("SPACE")
	Eventually(CF("create-org", orgName)).Should(Exit(0))
	Eventually(CF("create-space", spaceName1, "-o", orgName)).Should(Exit(0))
	Eventually(CF("create-space", spaceName2, "-o", orgName)).Should(Exit(0))
	DestroyHomeDir(homeDir)
	return orgName, spaceName1
}

func CreateOrgAndSpace(org string, space string) {
	Eventually(CF("create-org", org)).Should(Exit(0))
	Eventually(CF("create-space", space, "-o", org)).Should(Exit(0))
}

func CreateSpace(space string) {
	Eventually(CF("create-space", space)).Should(Exit(0))
}
