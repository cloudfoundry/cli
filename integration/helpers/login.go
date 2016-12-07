package helpers

import (
	"os"
	"strconv"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func SetAPI() (string, string) {
	apiURL := GetAPI()
	skipSSLValidation := skipSSLValidation()
	Eventually(CF("api", apiURL, skipSSLValidation)).Should(Exit(0))
	return apiURL, skipSSLValidation
}

func UnsetAPI() {
	Eventually(CF("api", "--unset")).Should(Exit(0))
}

func skipSSLValidation() string {
	if skip, err := strconv.ParseBool(os.Getenv("SKIP_SSL_VALIDATION")); err == nil && !skip {
		return ""
	}
	return "--skip-ssl-validation"
}

func GetAPI() string {
	apiURL := os.Getenv("CF_API")
	if apiURL == "" {
		return "https://api.bosh-lite.com"
	}
	return apiURL
}

func LoginCF() string {
	username, password := GetCredentials()
	Eventually(CF("auth", username, password)).Should(Exit(0))

	return username
}

func GetCredentials() (string, string) {
	username := os.Getenv("CF_USERNAME")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("CF_PASSWORD")
	if password == "" {
		password = "admin"
	}
	return username, password
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
