package helpers

import (
	"fmt"
	"strings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func SetupReadOnlyOrgAndSpace() (string, string) {
	homeDir := SetHomeDir()
	SetAPI()
	LoginCF()
	orgName := NewOrgName()
	spaceName1 := NewSpaceName()
	spaceName2 := NewSpaceName()
	Eventually(CF("create-org", orgName)).Should(Exit(0))
	Eventually(CF("create-space", spaceName1, "-o", orgName)).Should(Exit(0))
	Eventually(CF("create-space", spaceName2, "-o", orgName)).Should(Exit(0))
	DestroyHomeDir(homeDir)
	return orgName, spaceName1
}

func CreateOrgAndSpace(org string, space string) {
	CreateOrg(org)
	TargetOrg(org)
	CreateSpace(space)
}

func CreateOrg(org string) {
	Eventually(CF("create-org", org)).Should(Exit(0))
}

func CreateSpace(space string) {
	Eventually(CF("create-space", space)).Should(Exit(0))
}

func GetOrgGUID(orgName string) string {
	session := CF("org", "--guid", orgName)
	Eventually(session).Should(Exit(0))
	return strings.TrimSpace(string(session.Out.Contents()))
}

func GetSpaceGUID(spaceName string) string {
	session := CF("space", "--guid", spaceName)
	Eventually(session).Should(Exit(0))
	return strings.TrimSpace(string(session.Out.Contents()))
}

func QuickDeleteOrg(orgName string) {
	// If orgName is empty, the BeforeSuite has failed and attempting to delete
	// will produce a meaningless error.
	if orgName == "" {
		fmt.Println("Empty org name. Skipping deletion.")
		return
	}

	guid := GetOrgGUID(orgName)
	url := fmt.Sprintf("/v2/organizations/%s?recursive=true&async=true", guid)
	session := CF("curl", "-X", "DELETE", url)
	Eventually(session).Should(Exit(0))
}

func QuickDeleteOrgIfExists(orgName string) {
	session := CF("org", "--guid", orgName)
	Eventually(session).Should(Exit())
	if session.ExitCode() != 0 {
		return
	}
	guid := strings.TrimSpace(string(session.Out.Contents()))
	url := fmt.Sprintf("/v2/organizations/%s?recursive=true&async=true", guid)
	session = CF("curl", "-X", "DELETE", url)
	Eventually(session).Should(Exit())
}

func QuickDeleteSpace(spaceName string) {
	guid := GetSpaceGUID(spaceName)
	url := fmt.Sprintf("/v2/spaces/%s?recursive=true&async=true", guid)
	session := CF("curl", "-X", "DELETE", url)
	Eventually(session).Should(Exit(0))
}
