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

func SetAPI() (string, bool) {
	apiURL := GetAPI()
	skipSSLValidation := skipSSLValidation()
	if skipSSLValidation {
		Eventually(CF("api", apiURL, "--skip-ssl-validation")).Should(Exit(0))
	} else {
		Eventually(CF("api", apiURL)).Should(Exit(0))
	}
	return apiURL, skipSSLValidation
}

func UnsetAPI() {
	Eventually(CF("api", "--unset")).Should(Exit(0))
}

func skipSSLValidation() bool {
	if skip, err := strconv.ParseBool(os.Getenv("SKIP_SSL_VALIDATION")); err == nil && !skip {
		return false
	}
	return true
}

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

func LoginAs(username, password string) {
	env := map[string]string{
		"CF_USERNAME": username,
		"CF_PASSWORD": password,
	}

	for i := 0; i < 3; i++ {
		session := CFWithEnv(env, "auth")
		Eventually(session).Should(Exit())
		if session.ExitCode() == 0 {
			break
		}
		time.Sleep(3 * time.Second)
	}
}

func LoginCF() string {
	username, password := GetCredentials()
	LoginAs(username, password)
	return username
}

func LoginCFWithClientCredentials() string {
	username, password := SkipIfClientCredentialsNotSet()
	env := map[string]string{
		"CF_USERNAME": username,
		"CF_PASSWORD": password,
	}
	Eventually(CFWithEnv(env, "auth", "--client-credentials")).Should(Exit(0))

	return username
}

// GetCredentials returns back the username and the password.
func GetCredentials() (string, string) {
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

func LogoutCF() {
	Eventually(CF("logout")).Should(Exit(0))
}

func TargetOrgAndSpace(org string, space string) {
	Eventually(CF("target", "-o", org, "-s", space)).Should(Exit(0))
}

func TargetOrg(org string) {
	Eventually(CF("target", "-o", org)).Should(Exit(0))
}

func ClearTarget() {
	LogoutCF()
	LoginCF()
}

func SetupCF(org string, space string) {
	LoginCF()
	CreateOrgAndSpace(org, space)
	TargetOrgAndSpace(org, space)
}

func SwitchToNoRole() string {
	username, password := CreateUser()
	LogoutCF()
	LoginAs(username, password)
	return username
}

func SwitchToOrgRole(org, role string) string {
	username, password := CreateUserInOrgRole(org, role)
	LogoutCF()
	LoginAs(username, password)
	return username
}

func SwitchToSpaceRole(org, space, role string) string {
	username, password := CreateUserInSpaceRole(org, space, role)
	LogoutCF()
	LoginAs(username, password)
	return username
}
