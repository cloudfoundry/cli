package helpers

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// AddOrReplaceEnvironment will update environment if it already exists or will add
// a new environment with the given environment name and details.
func AddOrReplaceEnvironment(env []string, newEnvName string, newEnvVal string) []string {
	var found bool
	for i, envPair := range env {
		splitENV := strings.Split(envPair, "=")
		if splitENV[0] == newEnvName {
			env[i] = fmt.Sprintf("%s=%s", newEnvName, newEnvVal)
			found = true
		}
	}

	if !found {
		env = append(env, fmt.Sprintf("%s=%s", newEnvName, newEnvVal))
	}
	return env
}

// CheckEnvironmentTargetedCorrectly will confirm if the command requires an
// API to be targeted and logged in to run. It can optionally check if the
// command requires org and space to be targeted.
func CheckEnvironmentTargetedCorrectly(targetedOrganizationRequired bool, targetedSpaceRequired bool, testOrg string, command ...string) {
	LoginCF()

	if targetedOrganizationRequired {
		By("errors if org is not targeted")
		session := CF(command...)
		Eventually(session).Should(Say("FAILED"))
		Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org\\."))
		Eventually(session).Should(Exit(1))

		if targetedSpaceRequired {
			By("errors if space is not targeted")
			TargetOrg(testOrg)
			session := CF(command...)
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space\\."))
			Eventually(session).Should(Exit(1))
		}
	}

	By("errors if user not logged in")
	LogoutCF()
	session := CF(command...)
	Eventually(session).Should(Say("FAILED"))
	Eventually(session.Err).Should(Say("Not logged in\\. Use 'cf login' to log in\\."))
	Eventually(session).Should(Exit(1))

	By("errors if cli not targeted")
	UnsetAPI()
	session = CF(command...)
	Eventually(session).Should(Say("FAILED"))
	Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
	Eventually(session).Should(Exit(1))
}

// UnrefactoredCheckEnvironmentTargetedCorrectly will confirm if the command requires an
// API to be targeted and logged in to run. It can optionally check if the
// command requires org and space to be targeted.
func UnrefactoredCheckEnvironmentTargetedCorrectly(targetedOrganizationRequired bool, targetedSpaceRequired bool, testOrg string, command ...string) {
	LoginCF()

	if targetedOrganizationRequired {
		By("errors if org is not targeted")
		session := CF(command...)
		Eventually(session).Should(Say("FAILED"))
		Eventually(session).Should(Say("No org targeted, use 'cf target -o ORG' to target an org\\."))
		Eventually(session).Should(Exit(1))

		if targetedSpaceRequired {
			By("errors if space is not targeted")
			TargetOrg(testOrg)
			session := CF(command...)
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space\\."))
			Eventually(session).Should(Exit(1))
		}
	}

	By("errors if user not logged in")
	LogoutCF()
	session := CF(command...)
	Eventually(session).Should(Say("FAILED"))
	Eventually(session).Should(Say("Not logged in\\. Use 'cf login' to log in\\."))
	Eventually(session).Should(Exit(1))

	By("errors if cli not targeted")
	UnsetAPI()
	session = CF(command...)
	Eventually(session).Should(Say("FAILED"))
	Eventually(session).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
	Eventually(session).Should(Exit(1))
}
