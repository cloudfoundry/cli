package helpers

import (
	"fmt"
	"strings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

// SetupReadOnlyOrgAndSpace creates a randomly-named org containing two randomly-named
// spaces. It creates a new CF_HOME directory to run these commands, then deletes it
// afterwards.
func SetupReadOnlyOrgAndSpace() (string, string) {
	homeDir := SetHomeDir()
	SetAPI()
	LoginCF()
	orgName := NewOrgName()
	spaceName1 := NewSpaceName()
	spaceName2 := NewSpaceName()
	Eventually(CF("create-org", orgName)).Should(Exit(0))
	// TODO: remove when create-space works with CF_CLI_EXPERIMENTAL=true
	Eventually(CFWithEnv(map[string]string{"CF_CLI_EXPERIMENTAL": "false"}, "create-space", spaceName1, "-o", orgName)).Should(Exit(0))
	Eventually(CFWithEnv(map[string]string{"CF_CLI_EXPERIMENTAL": "false"}, "create-space", spaceName2, "-o", orgName)).Should(Exit(0))
	DestroyHomeDir(homeDir)
	return orgName, spaceName1
}

// CreateAndTargetOrg creates a randomly-named org and targets it.
func CreateAndTargetOrg() string {
	org := NewOrgName()
	CreateOrg(org)
	TargetOrg(org)
	return org
}

// CreateOrgAndSpace creates a randomly-named org and a randomly-named space in that org.
func CreateOrgAndSpace(org string, space string) {
	CreateOrg(org)
	TargetOrg(org)
	CreateSpace(space)
}

// CreateOrg creates an org with the given name using 'cf create-org'.
func CreateOrg(org string) {
	Eventually(CF("create-org", org)).Should(Exit(0))
}

// CreateSpace creates a space with the given name using 'cf create-space'.
func CreateSpace(space string) {
	// TODO: remove when create-space works with CF_CLI_EXPERIMENTAL=true
	Eventually(CFWithEnv(map[string]string{"CF_CLI_EXPERIMENTAL": "false"}, "create-space", space)).Should(Exit(0))
}

// GetOrgGUID gets the GUID of an org with the given name.
func GetOrgGUID(orgName string) string {
	session := CF("org", "--guid", orgName)
	Eventually(session).Should(Exit(0))
	return strings.TrimSpace(string(session.Out.Contents()))
}

// GetSpaceGUID gets the GUID of a space with the given name.
func GetSpaceGUID(spaceName string) string {
	session := CF("space", "--guid", spaceName)
	Eventually(session).Should(Exit(0))
	return strings.TrimSpace(string(session.Out.Contents()))
}

// QuickDeleteOrg deletes the org with the given name, if provided, using
// 'cf curl /v2/organizations... -X DELETE'.
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

// QuickDeleteOrgIfExists deletes the org with the given name, if it exists, using
// 'cf curl /v2/organizations... -X DELETE'.
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

// QuickDeleteSpace deletes the space with the given name, if it exists, using
// 'cf curl /v2/spaces... -X DELETE'.
func QuickDeleteSpace(spaceName string) {
	guid := GetSpaceGUID(spaceName)
	url := fmt.Sprintf("/v2/spaces/%s?recursive=true&async=true", guid)
	session := CF("curl", "-X", "DELETE", url)
	Eventually(session).Should(Exit(0))
}
