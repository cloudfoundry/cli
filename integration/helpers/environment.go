package helpers

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

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

func EnableDockerSupport() {
	tempHome := SetHomeDir()
	SetAPI()
	LoginCF()
	Eventually(CF("enable-feature-flag", "diego_docker")).Should(Exit(0))
	DestroyHomeDir(tempHome)
}

func CheckEnvironmentTargetedCorrectly(targetedOrganizationRequired bool, targetedSpaceRequired bool, testOrg string, command ...string) {
	LoginCF()

	if targetedOrganizationRequired {
		By("errors if org is not targeted")
		session := CF(command...)
		Eventually(session.Out).Should(Say("FAILED"))
		Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org\\."))
		Eventually(session).Should(Exit(1))

		if targetedSpaceRequired {
			By("errors if space is not targeted")
			TargetOrg(testOrg)
			session := CF(command...)
			Eventually(session.Out).Should(Say("FAILED"))
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
