package helpers

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

// SetAPI sets the API endpoint to the value of the CF_INT_API environment variable,
// or "https://api.bosh-lite.com" if not set. If the SKIP_SSL_VALIDATION environment
// variable is set, it will use the '--skip-ssl-validation' flag. It returns the API
// URL and a boolean indicating if SSL validation was skipped.
func SetAPI() (string, bool) {
	apiURL := GetAPI()
	skipSSLValidation := SkipSSLValidation()
	if skipSSLValidation {
		Eventually(CF("api", apiURL, "--skip-ssl-validation")).Should(Exit(0))
	} else {
		Eventually(CF("api", apiURL)).Should(Exit(0))
	}
	return apiURL, skipSSLValidation
}

// UnsetAPI unsets the currently set API endpoint for the CLI.
func UnsetAPI() {
	Eventually(CF("api", "--unset")).Should(Exit(0))
}

func SkipSSLValidation() bool {
	if skip, err := strconv.ParseBool(os.Getenv("SKIP_SSL_VALIDATION")); err == nil && !skip {
		return false
	}
	return true
}

// GetAPI gets the value of the CF_INT_API environment variable, if set, and prefixes
// it with "https://" if the value doesn't already start with "http". If the variable
// is not set, returns "https://api.bosh-lite.com".
func GetAPI() string {
	apiURL := os.Getenv("CF_INT_API")
	if apiURL == "" {
		return "https://api.bosh-lite.com"
	}
	if !strings.HasPrefix(apiURL, "http") {
		apiURL = fmt.Sprintf("https://%s", apiURL)
	}

	return apiURL
}

// LoginAs logs into the CLI with 'cf auth' and the given username and password,
// retrying up to 3 times on failures.
func LoginAs(username, password string) {
	env := map[string]string{
		"CF_USERNAME": username,
		"CF_PASSWORD": password,
	}

	var session *Session

	for i := 0; i < 3; i++ {
		session = CFWithEnv(env, "auth")
		Eventually(session).Should(Exit())
		if session.ExitCode() == 0 {
			return
		}
		time.Sleep(3 * time.Second)
	}
	Expect(session.ExitCode()).To(Equal(0))
}

// LoginCF logs into the CLI using the username and password from the CF_INT_USERNAME
// and CF_INT_PASSWORD environment variables, respectively, defaulting to "admin" for
// each if either is not set.
func LoginCF() string {
	if ClientCredentialsTestMode() {
		return LoginCFWithClientCredentials()
	}
	username, password := GetCredentials()
	LoginAs(username, password)
	return username
}

// LoginCFWithClientCredentials logs into the CLI using client credentials from the CF_INT_CLIENT_ID and
// CF_INT_CLIENT_SECRET environment variables and returns the client ID. If these environment variables
// are not set, it skips the current test.
func LoginCFWithClientCredentials() string {
	username, password := SkipIfClientCredentialsNotSet()
	env := map[string]string{
		"CF_USERNAME": username,
		"CF_PASSWORD": password,
	}
	Eventually(CFWithEnv(env, "auth", "--client-credentials")).Should(Exit(0))

	return username
}

// GetCredentials returns back the credentials for the user or client to authenticate with Cloud Foundry.
func GetCredentials() (string, string) {
	if ClientCredentialsTestMode() {
		return SkipIfClientCredentialsNotSet()
	}

	username := os.Getenv("CF_INT_USERNAME")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("CF_INT_PASSWORD")
	if password == "" {
		password = "admin"
	}
	return username, password
}

// SkipIfOIDCCredentialsNotSet returns back the username and the password for
// OIDC origin, or skips the test if those values are not set.
func SkipIfOIDCCredentialsNotSet() (string, string) {
	oidcUsername := os.Getenv("CF_INT_OIDC_USERNAME")
	oidcPassword := os.Getenv("CF_INT_OIDC_PASSWORD")

	if oidcUsername == "" || oidcPassword == "" {
		Skip("CF_INT_OIDC_USERNAME or CF_INT_OIDC_PASSWORD is not set")
	}

	return oidcUsername, oidcPassword
}

// LogoutCF logs out of the CLI.
func LogoutCF() {
	Eventually(CF("logout")).Should(Exit(0))
}

// TargetOrgAndSpace targets the given org and space with 'cf target'.
func TargetOrgAndSpace(org string, space string) {
	Eventually(CF("target", "-o", org, "-s", space)).Should(Exit(0))
}

// TargetOrg targets the given org with 'cf target'.
func TargetOrg(org string) {
	Eventually(CF("target", "-o", org)).Should(Exit(0))
}

// ClearTarget logs out and logs back into the CLI using LogoutCF and LoginCF.
func ClearTarget() {
	LogoutCF()
	LoginCF()
}

// SetupCF logs into the CLI with LoginCF, creates the given org and space, and targets that
// org and space.
func SetupCF(org string, space string) {
	LoginCF()
	CreateOrgAndSpace(org, space)
	TargetOrgAndSpace(org, space)
}

// SetupCFWithOrgOnly logs into the CLI with LoginCF, creates the given org, and targets it.
func SetupCFWithOrgOnly(org string) {
	LoginCF()
	CreateOrg(org)
	TargetOrg(org)
}

// SetupCFWithGeneratedOrgAndSpaceNames logs into the CLI with LoginCF, creates the org and
// space with generated names, and targets that org and space. Returns the generated org so
// that it can be deleted easily in cleanup step of the test.
func SetupCFWithGeneratedOrgAndSpaceNames() string {
	org := NewOrgName()
	space := NewSpaceName()

	SetupCF(org, space)
	return org
}

// SwitchToNoRole logs out of the CLI and logs back in as a newly-created user without a role.
func SwitchToNoRole() string {
	username, password := CreateUser()
	LogoutCF()
	LoginAs(username, password)
	return username
}

// SwitchToOrgRole logs out of the CLI and logs back in as a newly-created user with the given
// org role in the given org.
func SwitchToOrgRole(org, role string) string {
	username, password := CreateUserInOrgRole(org, role)
	LogoutCF()
	LoginAs(username, password)
	return username
}

// SwitchToSpaceRole logs out of the CLI and logs back in as a newly-created user with the given
// space role in the given space and org.
func SwitchToSpaceRole(org, space, role string) string {
	username, password := CreateUserInSpaceRole(org, space, role)
	LogoutCF()
	LoginAs(username, password)
	return username
}
